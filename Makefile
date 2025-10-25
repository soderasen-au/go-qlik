.PHONY: all build build-jwt build-pdfprinter test clean deps fmt vet lint help test-pdf compare-pdf

# Default target
all: deps fmt vet build test

# Build all binaries
build: build-jwt build-pdfprinter

# Build JWT encoder/decoder tool
build-jwt:
	@echo "Building JWT tool..."
	@go build -o bin/jwt ./cmd/jwt

# Build PDF printer tool
build-pdfprinter:
	@echo "Building PDF printer..."
	@go build -o bin/pdfprinter ./test/pdf

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
	@./bin/pdfprinter -format xlsx -name "TestReport" -output-folder test-reports
	@echo "  - Generating PDF report..."
	@./bin/pdfprinter -format pdf -name "TestReport" -output-folder test-reports -orientation portrait
	@echo "✓ Reports generated in test-reports/"
	@ls -lh test-reports/TestReport.*

# Compare generated reports with reference files
compare-pdf: test-pdf
	@echo "Comparing generated reports with reference..."
	@bash scripts/compare-reports.sh
	@echo ""
	@echo "Generating detailed analysis..."
	@python3 scripts/extract-pdf-issues.py
	@echo ""
	@echo "📊 Comparison complete! View results:"
	@echo "  📄 Text Report:  cat test-reports/comparison-report.txt"
	@echo "  🌐 HTML Report:  open test-reports/comparison-report.html"
	@echo "  📊 JSON Data:    test-reports/comparison-data.json"

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
	@rm -f test/pdf/*.pdf test/pdf/*.xlsx test/pdf/*.csv test/pdf/*.tsv
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
	@echo "  all              - Run deps, fmt, vet, build, and test (default)"
	@echo "  build            - Build all binaries (jwt, pdfprinter)"
	@echo "  build-jwt        - Build JWT tool only"
	@echo "  build-pdfprinter - Build PDF printer only"
	@echo "  test             - Run all tests"
	@echo "  test-engine      - Run engine package tests"
	@echo "  test-qrs         - Run QRS package tests"
	@echo "  test-report      - Run report package tests"
	@echo "  test-pdf         - Generate test reports (1 xlsx, 1 pdf) using pdfprinter"
	@echo "  compare-pdf      - Generate reports and compare with reference (creates detailed analysis)"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  deps             - Download and tidy dependencies"
	@echo "  fmt              - Format Go code"
	@echo "  vet              - Run go vet"
	@echo "  lint             - Run golint (if installed)"
	@echo "  staticcheck      - Run staticcheck (if installed)"
	@echo "  clean            - Remove build artifacts and generated files"
	@echo "  install          - Build and install binaries to bin/"
	@echo "  help             - Show this help message"
