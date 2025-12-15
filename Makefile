.PHONY: all build build-jwt build-pdfprinter build-excel-paging test clean deps deps-python fmt vet lint help test-pdf compare-pdf test-pivot compare-pivot test-excel-paging compare-excel-paging

# Python interpreter (use virtual environment if available)
PYTHON := $(shell if [ -f .venv/bin/python3 ]; then echo .venv/bin/python3; else echo python3; fi)

# Default target
all: deps fmt vet build test

# Build all binaries
build: build-jwt build-pdfprinter build-excel-paging

# Build JWT encoder/decoder tool
build-jwt:
	@echo "Building JWT tool..."
	@go build -o bin/jwt ./cmd/jwt

# Build PDF printer tool
build-pdfprinter:
	@echo "Building PDF printer..."
	@go build -o bin/pdfprinter ./test/pdf

# Build Excel paging printer tool
build-excel-paging:
	@echo "Building Excel paging printer..."
	@go build -o bin/excel_paging ./test/excel_paging

# Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests for specific package
test-engine:
	@echo "Running engine tests..."
	@go test -v ./qlik/engine

test-qrs:
	@echo "Running QRS tests..."
	@go test -v ./qlik/managed/qrs

test-report:
	@echo "Running report tests..."
	@go test -v ./report

# Generate test reports using pdfprinter
test-pdf: build-pdfprinter
	@echo "Generating test reports..."
	@mkdir -p test-reports
	@echo "  - Generating Excel report..."
	@./bin/pdfprinter -format xlsx -name "TestReport" -output-folder test-reports -certs-path test/certs/sa-win2k25
	@echo "  - Generating PDF report..."
	@./bin/pdfprinter -format pdf -name "TestReport" -output-folder test-reports -orientation portrait -certs-path test/certs/sa-win2k25
	@echo "✓ Reports generated in test-reports/"
	@ls -lh test-reports/TestReport.*

# Generate test pivot table report
test-pivot: build-pdfprinter
	@echo "Generating pivot table test report..."
	@mkdir -p test-reports
	@echo "  - Generating pivot table PDF report..."
	@./bin/pdfprinter -format pdf -name "TestReportPivotTable" -output-folder test-reports -obj-ids "RfEbJ" -orientation landscape -certs-path test/certs/sa-win2k25
	@echo "✓ Pivot table report generated in test-reports/"
	@ls -lh test-reports/TestReportPivotTable.*

# Compare generated reports with reference files
# Requirements: python3, poppler (pdftotext), uv
# Python packages: uv pip install -r scripts/requirements.txt
# Install uv: pip install uv
# macOS: brew install poppler
compare-pdf: test-pdf deps-python
	@echo "Comparing generated reports with reference..."
	@PYTHON=$(PYTHON) bash scripts/compare-reports.sh || (echo "❌ Comparison failed. Check error messages above." && exit 1)
	@echo ""
	@echo "Generating detailed analysis..."
	@$(PYTHON) scripts/extract-pdf-issues.py || echo "⚠️  Python analysis skipped"
	@echo ""
	@echo "📊 Comparison complete! View results:"
	@echo "  📄 Text Report:  cat test-reports/comparison-report.txt"
	@if [ -f test-reports/comparison-report.html ]; then \
		echo "  🌐 HTML Report:  open test-reports/comparison-report.html"; \
	fi
	@if [ -f test-reports/comparison-data.json ]; then \
		echo "  📊 JSON Data:    test-reports/comparison-data.json"; \
	fi

# Generate paginated Excel report using excel_paging tool
test-excel-paging: build-excel-paging
	@echo "Generating paginated Excel report..."
	@mkdir -p test-reports
	@echo "  - Generating paginated report with 100 rows per page..."
	@./bin/excel_paging -rows-per-page 100 -title "Paginated Sales Report" \
		-show-col-nums -show-subtotals -all-borders \
		-name "TestPaginatedReport" -output-folder test-reports \
		-certs-path test/certs/sa-win2k25
	@echo "✓ Paginated report generated in test-reports/"
	@ls -lh test-reports/Paginated*.xlsx 2>/dev/null || ls -lh test-reports/TestPaginatedReport.xlsx 2>/dev/null || echo "  (check test-reports/ for output)"

# Compare paginated Excel report with reference
# Uses Python script to compare cell values across all sheets
compare-excel-paging: test-excel-paging
	@echo "Comparing paginated Excel report with reference..."
	@if [ ! -f test/excel_paging/PaginatedSalesReport_reference.xlsx ]; then \
		echo "❌ Reference file not found: test/excel_paging/PaginatedSalesReport_reference.xlsx"; \
		exit 1; \
	fi
	@GENERATED_FILE=$$(ls test-reports/Paginated*.xlsx 2>/dev/null | head -1); \
	if [ -z "$$GENERATED_FILE" ]; then \
		GENERATED_FILE="test-reports/TestPaginatedReport.xlsx"; \
	fi; \
	if [ ! -f "$$GENERATED_FILE" ]; then \
		echo "❌ Generated file not found. The tool may have crashed."; \
		echo "  Check test-reports/excel_paging.log for error details."; \
		exit 1; \
	fi; \
	echo "  - Reference: test/excel_paging/PaginatedSalesReport_reference.xlsx"; \
	echo "  - Generated: $$GENERATED_FILE"; \
	echo ""; \
	$(PYTHON) scripts/compare-excel.py \
		test/excel_paging/PaginatedSalesReport_reference.xlsx \
		"$$GENERATED_FILE"

# Compare pivot table report with reference
# Generate PDF with obj-ids=RfEbJ and compare content with baseline XLSX
compare-pivot: test-pivot
	@echo "Comparing pivot table report with reference..."
	@mkdir -p test-reports
	@if [ ! -f test/pdf/TestReportPivotTable.xlsx ]; then \
		echo "❌ Reference file not found: test/pdf/TestReportPivotTable.xlsx"; \
		exit 1; \
	fi
	@echo "  - Reference (baseline): test/pdf/TestReportPivotTable.xlsx"
	@echo "  - Generated PDF: test-reports/TestReportPivotTable.pdf"
	@if [ ! -f test-reports/TestReportPivotTable.pdf ]; then \
		echo "❌ Generated PDF not found. The tool may have crashed."; \
		echo "  Check test-reports/pdf.log for error details."; \
		exit 1; \
	fi
	@if command -v pdftotext >/dev/null 2>&1 && command -v libreoffice >/dev/null 2>&1; then \
		echo "  - Extracting text from PDF..."; \
		pdftotext test-reports/TestReportPivotTable.pdf test-reports/TestReportPivotTable_pdf.txt 2>/dev/null || true; \
		echo "  - Converting reference XLSX to CSV..."; \
		libreoffice --headless --convert-to csv --outdir test-reports test/pdf/TestReportPivotTable.xlsx >/dev/null 2>&1; \
		mv test-reports/TestReportPivotTable.csv test-reports/TestReportPivotTable_reference.csv 2>/dev/null || true; \
		echo ""; \
		echo "📊 Comparison files generated:"; \
		echo "  - PDF text: test-reports/TestReportPivotTable_pdf.txt"; \
		echo "  - Reference CSV: test-reports/TestReportPivotTable_reference.csv"; \
		echo ""; \
		echo "Manual comparison needed - verify PDF content matches baseline XLSX data"; \
	else \
		echo "⚠️  Tools not found. Install: brew install poppler (pdftotext) and libreoffice"; \
	fi

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Install Python dependencies for report comparison
deps-python:
	@echo "Installing Python dependencies for report comparison..."
	@if [ ! -d .venv ]; then \
		echo "Creating virtual environment..."; \
		uv venv; \
	fi
	@uv pip install -r scripts/requirements.txt
	@echo "✓ Python dependencies installed"

# Format Go code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run golint (if installed)
lint:
	@echo "Running golint..."
	@which golint > /dev/null 2>&1 && golint ./... || echo "golint not installed. Run: go install golang.org/x/lint/golint@latest"

# Run staticcheck (if installed)
staticcheck:
	@echo "Running staticcheck..."
	@which staticcheck > /dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed. Run: go install honnef.co/go/tools/cmd/staticcheck@latest"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf test-reports/
	@rm -f pdfprinter jwt
	@rm -f coverage.out coverage.html
	@rm -f test/pdf/pdfprinter
	@rm -f test/pdf/*.pdf test/pdf/*.csv test/pdf/*.tsv
	@rm -f cmd/jwt/jwt
	@find . -name "*.log" -type f -delete
	@echo "Clean complete"

# Install tools to bin directory
install: build
	@echo "Installing binaries to bin/..."
	@mkdir -p bin
	@echo "Binaries installed in ./bin/"

# Show help
help:
	@echo "Available targets:"
	@echo "  all                - Run deps, fmt, vet, build, and test (default)"
	@echo "  build              - Build all binaries (jwt, pdfprinter, excel_paging)"
	@echo "  build-jwt          - Build JWT tool only"
	@echo "  build-pdfprinter   - Build PDF printer only"
	@echo "  build-excel-paging - Build Excel paging printer only"
	@echo "  test               - Run all tests"
	@echo "  test-engine        - Run engine package tests"
	@echo "  test-qrs           - Run QRS package tests"
	@echo "  test-report        - Run report package tests"
	@echo "  test-pdf           - Generate test reports (1 xlsx, 1 pdf) using pdfprinter"
	@echo "  test-pivot         - Generate pivot table PDF report (obj-ids=RfEbJ)"
	@echo "  test-excel-paging  - Generate paginated Excel report with multiple sheets"
	@echo "                       Each sheet contains rows_per_page rows with headers,"
	@echo "                       selections, column numbers, and page subtotals"
	@echo "  compare-excel-paging - Generate paginated Excel and compare with reference"
	@echo "                       (requires: python3, openpyxl)"
	@echo "                       Install: pip install openpyxl"
	@echo "  compare-pdf        - Generate reports and compare with reference"
	@echo "                       (requires: python3, poppler, uv)"
	@echo "                       Install: pip install uv, brew install poppler (macOS)"
	@echo "  compare-pivot      - Generate pivot table PDF and compare with baseline XLSX"
	@echo "                       (requires: poppler, libreoffice)"
	@echo "                       Use this to debug and fix pivot table rendering issues"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  deps               - Download and tidy Go dependencies"
	@echo "  deps-python        - Install Python dependencies for report comparison"
	@echo "  fmt                - Format Go code"
	@echo "  vet                - Run go vet"
	@echo "  lint               - Run golint (if installed)"
	@echo "  staticcheck        - Run staticcheck (if installed)"
	@echo "  clean              - Remove build artifacts and generated files"
	@echo "  install            - Build and install binaries to bin/"
	@echo "  help               - Show this help message"
