//go:build windows

package report

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/soderasen-au/go-common/loggers"
)

// TestExcelToPDFWin_Basic tests basic conversion with default settings
func TestExcelToPDFWin_Basic(t *testing.T) {
	// Skip if no Excel file available
	testFile := "test/test_report.xlsx"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("test file not found: %s", testFile)
	}

	outputDir := "test-reports"
	os.MkdirAll(outputDir, 0755)
	outputPath := filepath.Join(outputDir, "test_basic.pdf")

	config := ExcelToPDFWinConfig{
		InputExcelPath: testFile,
		OutputPDFPath:  outputPath,
		Logger:         loggers.CoreDebugLogger,
	}

	converter, res := NewExcelToPDFWin(config)
	if res != nil {
		t.Fatalf("NewExcelToPDFWin failed: %v", res)
	}

	if res := converter.Convert(); res != nil {
		t.Fatalf("Convert failed: %v", res)
	}

	// Verify output exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("output PDF not created: %s", outputPath)
	}

	t.Logf("✓ Basic conversion successful: %s", outputPath)
}

// TestExcelToPDFWin_Portrait tests portrait orientation
func TestExcelToPDFWin_Portrait(t *testing.T) {
	testFile := "test/test_report.xlsx"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("test file not found: %s", testFile)
	}

	outputDir := "test-reports"
	os.MkdirAll(outputDir, 0755)
	outputPath := filepath.Join(outputDir, "test_portrait.pdf")

	config := ExcelToPDFWinConfig{
		InputExcelPath: testFile,
		OutputPDFPath:  outputPath,
		Orientation:    xlPortrait,
		Logger:         loggers.CoreDebugLogger,
	}

	converter, res := NewExcelToPDFWin(config)
	if res != nil {
		t.Fatalf("NewExcelToPDFWin failed: %v", res)
	}

	if res := converter.Convert(); res != nil {
		t.Fatalf("Convert failed: %v", res)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("output PDF not created: %s", outputPath)
	}

	t.Logf("✓ Portrait conversion successful: %s", outputPath)
}

// TestExcelToPDFWin_CustomMargins tests custom margin settings
func TestExcelToPDFWin_CustomMargins(t *testing.T) {
	testFile := "test/test_report.xlsx"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("test file not found: %s", testFile)
	}

	outputDir := "test-reports"
	os.MkdirAll(outputDir, 0755)
	outputPath := filepath.Join(outputDir, "test_custom_margins.pdf")

	config := ExcelToPDFWinConfig{
		InputExcelPath: testFile,
		OutputPDFPath:  outputPath,
		LeftMargin:     0.25,
		RightMargin:    0.25,
		TopMargin:      0.5,
		BottomMargin:   0.5,
		Logger:         loggers.CoreDebugLogger,
	}

	converter, res := NewExcelToPDFWin(config)
	if res != nil {
		t.Fatalf("NewExcelToPDFWin failed: %v", res)
	}

	if res := converter.Convert(); res != nil {
		t.Fatalf("Convert failed: %v", res)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("output PDF not created: %s", outputPath)
	}

	t.Logf("✓ Custom margins conversion successful: %s", outputPath)
}

// TestExcelToPDFWin_MultipleSheets tests exporting multiple sheets to separate PDFs
func TestExcelToPDFWin_MultipleSheets(t *testing.T) {
	testFile := "test/test_report.xlsx"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("test file not found: %s", testFile)
	}

	outputDir := "test-reports/multi-sheets"
	os.MkdirAll(outputDir, 0755)

	config := ExcelToPDFWinConfig{
		InputExcelPath:     testFile,
		OutputPDFPath:      outputDir,
		ExportMultiplePDFs: true,
		Logger:             loggers.CoreDebugLogger,
	}

	converter, res := NewExcelToPDFWin(config)
	if res != nil {
		t.Fatalf("NewExcelToPDFWin failed: %v", res)
	}

	if res := converter.Convert(); res != nil {
		t.Fatalf("Convert failed: %v", res)
	}

	// Check that output directory has PDF files
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("failed to read output directory: %v", err)
	}

	pdfCount := 0
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".pdf" {
			pdfCount++
			t.Logf("  - %s", entry.Name())
		}
	}

	if pdfCount == 0 {
		t.Errorf("no PDF files created in %s", outputDir)
	}

	t.Logf("✓ Multiple sheets conversion successful: %d PDFs created", pdfCount)
}

// TestExcelToPDFWin_SpecificSheets tests exporting specific sheets by name
func TestExcelToPDFWin_SpecificSheets(t *testing.T) {
	testFile := "test/test_report.xlsx"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("test file not found: %s", testFile)
	}

	outputDir := "test-reports"
	os.MkdirAll(outputDir, 0755)
	outputPath := filepath.Join(outputDir, "test_specific_sheets.pdf")

	// Specify sheet names - adjust based on your test file
	config := ExcelToPDFWinConfig{
		InputExcelPath: testFile,
		OutputPDFPath:  outputPath,
		SheetNames:     []string{"TestReport"}, // Adjust to actual sheet name
		Logger:         loggers.CoreDebugLogger,
	}

	converter, res := NewExcelToPDFWin(config)
	if res != nil {
		t.Fatalf("NewExcelToPDFWin failed: %v", res)
	}

	if res := converter.Convert(); res != nil {
		t.Fatalf("Convert failed: %v", res)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("output PDF not created: %s", outputPath)
	}

	t.Logf("✓ Specific sheets conversion successful: %s", outputPath)
}

// TestExcelToPDFWin_PrintArea tests setting a custom print area
func TestExcelToPDFWin_PrintArea(t *testing.T) {
	testFile := "test/test_report.xlsx"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("test file not found: %s", testFile)
	}

	outputDir := "test-reports"
	os.MkdirAll(outputDir, 0755)
	outputPath := filepath.Join(outputDir, "test_print_area.pdf")

	config := ExcelToPDFWinConfig{
		InputExcelPath: testFile,
		OutputPDFPath:  outputPath,
		PrintArea:      "A1:M50", // Export only specific range
		Logger:         loggers.CoreDebugLogger,
	}

	converter, res := NewExcelToPDFWin(config)
	if res != nil {
		t.Fatalf("NewExcelToPDFWin failed: %v", res)
	}

	if res := converter.Convert(); res != nil {
		t.Fatalf("Convert failed: %v", res)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("output PDF not created: %s", outputPath)
	}

	t.Logf("✓ Print area conversion successful: %s", outputPath)
}

// TestExcelToPDFWin_Validation tests configuration validation
func TestExcelToPDFWin_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      ExcelToPDFWinConfig
		shouldError bool
	}{
		{
			name: "missing input path",
			config: ExcelToPDFWinConfig{
				OutputPDFPath: "output.pdf",
			},
			shouldError: true,
		},
		{
			name: "missing output path",
			config: ExcelToPDFWinConfig{
				InputExcelPath: "input.xlsx",
			},
			shouldError: true,
		},
		{
			name: "conflicting sheet selection",
			config: ExcelToPDFWinConfig{
				InputExcelPath: "input.xlsx",
				OutputPDFPath:  "output.pdf",
				SheetNames:     []string{"Sheet1"},
				SheetIndices:   []int{1},
			},
			shouldError: true,
		},
		{
			name: "valid config",
			config: ExcelToPDFWinConfig{
				InputExcelPath: "input.xlsx",
				OutputPDFPath:  "output.pdf",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, res := NewExcelToPDFWin(tt.config)
			if tt.shouldError && res == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.shouldError && res != nil {
				t.Errorf("expected no error but got: %v", res)
			}
		})
	}
}

// TestSanitizeFilename tests filename sanitization
func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Sheet1", "Sheet1"},
		{"Sheet:Name", "Sheet_Name"},
		{"My<Sheet>", "My_Sheet_"},
		{"Sheet/Name", "Sheet_Name"},
		{"Sheet|Name", "Sheet_Name"},
		{"Sheet?Name", "Sheet_Name"},
		{"Sheet*Name", "Sheet_Name"},
		{`Sheet\Name`, "Sheet_Name"},
		{`"SheetName"`, "_SheetName_"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
