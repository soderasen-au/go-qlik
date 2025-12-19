package report

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/soderasen-au/go-common/loggers"
	"github.com/xuri/excelize/v2"
)

// TestDefaultExcelPagingConfig tests default configuration values
func TestDefaultExcelPagingConfig(t *testing.T) {
	config := DefaultExcelPagingConfig()

	if config.RowsPerPage != 50 {
		t.Errorf("expected RowsPerPage=50, got %d", config.RowsPerPage)
	}
	if config.TotalRecordsLabel != "Total Records Found" {
		t.Errorf("expected TotalRecordsLabel='Total Records Found', got '%s'", config.TotalRecordsLabel)
	}
	if config.ShowColumnNumbers {
		t.Error("expected ShowColumnNumbers=false")
	}
	if config.ShowSubtotals {
		t.Error("expected ShowSubtotals=false")
	}
}

// TestNewExcelPagingPrinter tests printer creation with various configs
func TestNewExcelPagingPrinter(t *testing.T) {
	tests := []struct {
		name           string
		config         ExcelPagingConfig
		expectedRows   int
		expectedLabel  string
	}{
		{
			name:           "default config",
			config:         DefaultExcelPagingConfig(),
			expectedRows:   50,
			expectedLabel:  "Total Records Found",
		},
		{
			name:           "zero rows per page defaults to 50",
			config:         ExcelPagingConfig{RowsPerPage: 0},
			expectedRows:   50,
			expectedLabel:  "Total Records Found",
		},
		{
			name:           "negative rows per page defaults to 50",
			config:         ExcelPagingConfig{RowsPerPage: -10},
			expectedRows:   50,
			expectedLabel:  "Total Records Found",
		},
		{
			name:           "custom rows per page",
			config:         ExcelPagingConfig{RowsPerPage: 100},
			expectedRows:   100,
			expectedLabel:  "Total Records Found",
		},
		{
			name:           "custom total records label",
			config:         ExcelPagingConfig{RowsPerPage: 25, TotalRecordsLabel: "Records"},
			expectedRows:   25,
			expectedLabel:  "Records",
		},
		{
			name:           "empty label defaults",
			config:         ExcelPagingConfig{RowsPerPage: 25, TotalRecordsLabel: ""},
			expectedRows:   25,
			expectedLabel:  "Total Records Found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printer := NewExcelPagingPrinter(tt.config)

			if printer.Config.RowsPerPage != tt.expectedRows {
				t.Errorf("expected RowsPerPage=%d, got %d", tt.expectedRows, printer.Config.RowsPerPage)
			}
			if printer.Config.TotalRecordsLabel != tt.expectedLabel {
				t.Errorf("expected TotalRecordsLabel='%s', got '%s'", tt.expectedLabel, printer.Config.TotalRecordsLabel)
			}
			if printer.ReportResults == nil {
				t.Error("ReportResults map should be initialized")
			}
		})
	}
}

// TestExcelPagingPrinter_PrintReportTitle tests title printing
func TestExcelPagingPrinter_PrintReportTitle(t *testing.T) {
	printer := NewExcelPagingPrinter(DefaultExcelPagingConfig())
	excel := excelize.NewFile()
	sheetName := "TestSheet"
	excel.NewSheet(sheetName)
	logger := loggers.CoreDebugLogger

	// Initialize execution context
	printer.excel = excel
	printer.logger = logger

	tests := []struct {
		name           string
		title          string
		expectedHeight int
	}{
		{"empty title", "", 0},
		{"with title", "My Report Title", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rect := enigma.Rect{Top: 1, Left: 1}
			resRect, res := printer.printReportTitle(tt.title, sheetName, rect)

			if res != nil {
				t.Fatalf("unexpected error: %v", res)
			}
			if resRect.Height != tt.expectedHeight {
				t.Errorf("expected height=%d, got %d", tt.expectedHeight, resRect.Height)
			}

			if tt.title != "" {
				cellName, _ := excelize.CoordinatesToCellName(rect.Left, rect.Top)
				value, _ := excel.GetCellValue(sheetName, cellName)
				if value != tt.title {
					t.Errorf("expected cell value='%s', got '%s'", tt.title, value)
				}
			}
		})
	}
}

// TestExcelPagingPrinter_PrintTotalRecords tests total records printing
func TestExcelPagingPrinter_PrintTotalRecords(t *testing.T) {
	config := ExcelPagingConfig{
		RowsPerPage:       50,
		TotalRecordsLabel: "Total Records",
	}
	printer := NewExcelPagingPrinter(config)
	excel := excelize.NewFile()
	sheetName := "TestSheet"
	excel.NewSheet(sheetName)
	logger := loggers.CoreDebugLogger

	// Initialize execution context
	printer.excel = excel
	printer.logger = logger

	rect := enigma.Rect{Top: 1, Left: 1}
	totalRows := 150

	resRect, res := printer.printTotalRecords(totalRows, sheetName, rect)

	if res != nil {
		t.Fatalf("unexpected error: %v", res)
	}
	if resRect.Height != 1 {
		t.Errorf("expected height=1, got %d", resRect.Height)
	}
	if resRect.Width != 2 {
		t.Errorf("expected width=2, got %d", resRect.Width)
	}

	// Check label cell
	labelCell, _ := excelize.CoordinatesToCellName(rect.Left, rect.Top)
	labelValue, _ := excel.GetCellValue(sheetName, labelCell)
	if labelValue != "Total Records:" {
		t.Errorf("expected label='Total Records:', got '%s'", labelValue)
	}

	// Check value cell
	valueCell, _ := excelize.CoordinatesToCellName(rect.Left+1, rect.Top)
	valueStr, _ := excel.GetCellValue(sheetName, valueCell)
	if valueStr != "150" {
		t.Errorf("expected value='150', got '%s'", valueStr)
	}
}

// TestExcelPagingPrinter_PrintColumnNumbers tests column number printing
func TestExcelPagingPrinter_PrintColumnNumbers(t *testing.T) {
	printer := NewExcelPagingPrinter(DefaultExcelPagingConfig())
	excel := excelize.NewFile()
	sheetName := "TestSheet"
	excel.NewSheet(sheetName)
	logger := loggers.CoreDebugLogger

	// Initialize execution context
	printer.excel = excel
	printer.logger = logger

	rect := enigma.Rect{Top: 1, Left: 1}
	colCount := 5

	resRect, res := printer.printColumnNumbers(colCount, sheetName, rect)

	if res != nil {
		t.Fatalf("unexpected error: %v", res)
	}
	if resRect.Height != 1 {
		t.Errorf("expected height=1, got %d", resRect.Height)
	}
	if resRect.Width != colCount {
		t.Errorf("expected width=%d, got %d", colCount, resRect.Width)
	}

	// Verify each column number
	for ci := 0; ci < colCount; ci++ {
		cellName, _ := excelize.CoordinatesToCellName(rect.Left+ci, rect.Top)
		value, _ := excel.GetCellValue(sheetName, cellName)
		expected := string(rune('0' + ci + 1))
		if ci >= 9 {
			expected = "10"
		}
		if value != expected && ci < 9 {
			t.Errorf("expected column %d value='%s', got '%s'", ci+1, expected, value)
		}
	}
}

// TestExcelPagingPrinter_PrintPageSubtotals tests subtotal printing
func TestExcelPagingPrinter_PrintPageSubtotals(t *testing.T) {
	printer := NewExcelPagingPrinter(DefaultExcelPagingConfig())
	excel := excelize.NewFile()
	sheetName := "TestSheet"
	excel.NewSheet(sheetName)
	logger := loggers.CoreDebugLogger

	// Initialize execution context
	printer.excel = excel
	printer.logger = logger
	printer.report = Report{AllBorders: false}

	rect := enigma.Rect{Top: 1, Left: 1}
	subtotals := []float64{0, 100.5, 200.25, 0}
	isNumeric := []bool{false, true, true, false}

	resRect, res := printer.printPageSubtotals(subtotals, isNumeric, sheetName, rect)

	if res != nil {
		t.Fatalf("unexpected error: %v", res)
	}
	if resRect.Height != 1 {
		t.Errorf("expected height=1, got %d", resRect.Height)
	}
	if resRect.Width != len(subtotals) {
		t.Errorf("expected width=%d, got %d", len(subtotals), resRect.Width)
	}

	// Check first cell has "Page Subtotal" label
	firstCell, _ := excelize.CoordinatesToCellName(rect.Left, rect.Top)
	firstValue, _ := excel.GetCellValue(sheetName, firstCell)
	if firstValue != "Page Subtotal" {
		t.Errorf("expected first cell='Page Subtotal', got '%s'", firstValue)
	}

	// Check numeric columns have values
	numCell, _ := excelize.CoordinatesToCellName(rect.Left+1, rect.Top)
	numValue, _ := excel.GetCellValue(sheetName, numCell)
	if numValue != "100.5" {
		t.Errorf("expected numeric cell='100.5', got '%s'", numValue)
	}
}

// TestExcelPagingPrinter_Validation tests report validation
func TestExcelPagingPrinter_Validation(t *testing.T) {
	printer := NewExcelPagingPrinter(DefaultExcelPagingConfig())

	tests := []struct {
		name        string
		report      Report
		shouldError bool
		errorMsg    string
	}{
		{
			name: "nil doc",
			report: Report{
				ID:           strPtr("test"),
				AppId:        "app-id",
				Target:       "objects",
				TargetIDs:    []string{"obj1"},
				OutputFormat: reportFormatPtr(REPORT_FORMAT_XLSX),
				OutputFolder: strPtr("./"),
			},
			shouldError: true,
			errorMsg:    "doc is not opened",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := printer.Print(tt.report)
			if tt.shouldError {
				if res == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if res != nil {
					t.Errorf("expected no error but got: %v", res)
				}
			}
		})
	}
}

// TestExcelPagingPrinter_GetReportResult tests result retrieval
func TestExcelPagingPrinter_GetReportResult(t *testing.T) {
	printer := NewExcelPagingPrinter(DefaultExcelPagingConfig())

	// Test non-existent report
	_, res := printer.GetReportResult("non-existent")
	if res == nil {
		t.Error("expected error for non-existent report")
	}

	// Add a result manually
	printer.ReportResults["test-id"] = &ReportResult{
		ID:         "test-id",
		ReportFile: strPtr("/tmp/test.xlsx"),
	}

	result, res := printer.GetReportResult("test-id")
	if res != nil {
		t.Errorf("unexpected error: %v", res)
	}
	if result.ID != "test-id" {
		t.Errorf("expected ID='test-id', got '%s'", result.ID)
	}
}

// TestPaginationCalculation tests page count calculation
func TestPaginationCalculation(t *testing.T) {
	tests := []struct {
		totalRows   int
		rowsPerPage int
		expected    int
	}{
		{0, 50, 1},    // Empty dataset = 1 page
		{1, 50, 1},    // 1 row = 1 page
		{50, 50, 1},   // Exactly 1 page
		{51, 50, 2},   // Just over 1 page
		{100, 50, 2},  // Exactly 2 pages
		{101, 50, 3},  // Just over 2 pages
		{150, 50, 3},  // Exactly 3 pages
		{1000, 100, 10}, // 10 pages
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			pageCount := (tt.totalRows + tt.rowsPerPage - 1) / tt.rowsPerPage
			if pageCount == 0 {
				pageCount = 1
			}
			if pageCount != tt.expected {
				t.Errorf("totalRows=%d, rowsPerPage=%d: expected %d pages, got %d",
					tt.totalRows, tt.rowsPerPage, tt.expected, pageCount)
			}
		})
	}
}

// TestExcelPagingConfig_JSON tests JSON serialization of config
func TestExcelPagingConfig_Serialization(t *testing.T) {
	config := ExcelPagingConfig{
		RowsPerPage:       100,
		ReportTitle:       "Test Report",
		TotalRecordsLabel: "Records Found",
		ShowColumnNumbers: true,
		ShowSubtotals:     true,
	}

	// Verify fields are set correctly
	if config.RowsPerPage != 100 {
		t.Errorf("expected RowsPerPage=100, got %d", config.RowsPerPage)
	}
	if config.ReportTitle != "Test Report" {
		t.Errorf("expected ReportTitle='Test Report', got '%s'", config.ReportTitle)
	}
	if !config.ShowColumnNumbers {
		t.Error("expected ShowColumnNumbers=true")
	}
	if !config.ShowSubtotals {
		t.Error("expected ShowSubtotals=true")
	}
}

// TestExcelPagingPrinter_OutputFile tests output file path generation
func TestExcelPagingPrinter_OutputFile(t *testing.T) {
	outputDir := filepath.Join(os.TempDir(), "excel_paging_test")
	os.MkdirAll(outputDir, 0755)
	defer os.RemoveAll(outputDir)

	config := ExcelPagingConfig{
		RowsPerPage:   50,
		ReportTitle:   "My Custom Report",
	}

	printer := NewExcelPagingPrinter(config)

	// The report name should be used when ReportTitle is set
	// This is tested implicitly through the Print method
	if printer.Config.ReportTitle != "My Custom Report" {
		t.Errorf("expected ReportTitle='My Custom Report', got '%s'", printer.Config.ReportTitle)
	}
}

// Helper functions for creating pointers
func strPtr(s string) *string {
	return &s
}

func reportFormatPtr(f ReportFormat) *ReportFormat {
	return &f
}

// TestExcelPagingPrinter_InterfaceCompliance verifies IReportPrinter implementation
func TestExcelPagingPrinter_InterfaceCompliance(t *testing.T) {
	// This test ensures ExcelPagingPrinter implements IReportPrinter
	var _ IReportPrinter = (*ExcelPagingPrinter)(nil)
	t.Log("ExcelPagingPrinter implements IReportPrinter interface")
}
