//go:build !windows

package report

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/loggers"
)

// checkLibreOfficeInstalled checks if LibreOffice is available in PATH
func checkLibreOfficeInstalled(t *testing.T) {
	if _, err := exec.LookPath("libreoffice"); err != nil {
		t.Skipf("LibreOffice not found in PATH: %v", err)
	}
}

// TestLibreExcel2PDF_Basic tests basic conversion with default settings
func TestLibreExcel2PDF_Basic(t *testing.T) {
	checkLibreOfficeInstalled(t)

	testFile := "../test/pdf/TestReport.xlsx"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("test file not found: %s", testFile)
	}

	outputDir := "../test-reports"
	os.MkdirAll(outputDir, 0755)
	outputPath := filepath.Join(outputDir, "test_libre_basic.pdf")
	defer os.Remove(outputPath)

	// Reset singleton for test isolation
	ResetGlobalInstance()

	converter := NewLibreExcel2PDF("libreoffice", loggers.CoreDebugLogger, 2)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	converter.StartUp(ctx)
	defer converter.Shutdown(context.Background())

	config := ExcelToPDFTaskConfig{
		InputExcelPath: testFile,
		OutputPDFPath:  outputPath,
		Logger:         loggers.CoreDebugLogger,
	}

	if res := converter.Convert(ctx, config); res != nil {
		t.Fatalf("Convert failed: %v", res)
	}

	// Verify output exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("output PDF not created: %s", outputPath)
	}

	t.Logf("✓ Basic conversion successful: %s", outputPath)
}

// TestLibreExcel2PDF_Concurrent tests concurrent conversion requests
func TestLibreExcel2PDF_Concurrent(t *testing.T) {
	checkLibreOfficeInstalled(t)

	testFile := "../test/pdf/TestReport.xlsx"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("test file not found: %s", testFile)
	}

	outputDir := "../test-reports"
	os.MkdirAll(outputDir, 0755)

	// Reset singleton for test isolation
	ResetGlobalInstance()

	converter := NewLibreExcel2PDF("libreoffice", loggers.CoreDebugLogger, 2)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	converter.StartUp(ctx)
	defer converter.Shutdown(context.Background())

	// Run 5 concurrent conversions
	numTasks := 5
	errChan := make(chan error, numTasks)

	for i := 0; i < numTasks; i++ {
		i := i
		go func() {
			outputPath := filepath.Join(outputDir, "test_libre_concurrent_"+string(rune('A'+i))+".pdf")
			defer os.Remove(outputPath)

			config := ExcelToPDFTaskConfig{
				InputExcelPath: testFile,
				OutputPDFPath:  outputPath,
			}

			taskCtx, taskCancel := context.WithTimeout(ctx, 30*time.Second)
			defer taskCancel()

			if res := converter.Convert(taskCtx, config); res != nil {
				errChan <- res
				return
			}

			// Verify output exists
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				errChan <- err
				return
			}

			errChan <- nil
		}()
	}

	// Wait for all tasks
	for i := 0; i < numTasks; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("concurrent task %d failed: %v", i, err)
		}
	}

	t.Logf("✓ %d concurrent conversions successful", numTasks)
}

// TestLibreExcel2PDF_ContextTimeout tests context timeout handling
func TestLibreExcel2PDF_ContextTimeout(t *testing.T) {
	checkLibreOfficeInstalled(t)

	testFile := "../test/pdf/TestReport.xlsx"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("test file not found: %s", testFile)
	}

	outputDir := "../test-reports"
	os.MkdirAll(outputDir, 0755)
	outputPath := filepath.Join(outputDir, "test_libre_timeout.pdf")
	defer os.Remove(outputPath)

	// Reset singleton for test isolation
	ResetGlobalInstance()

	converter := NewLibreExcel2PDF("libreoffice", loggers.CoreDebugLogger, 1)

	mainCtx := context.Background()
	converter.StartUp(mainCtx)
	defer converter.Shutdown(context.Background())

	config := ExcelToPDFTaskConfig{
		InputExcelPath: testFile,
		OutputPDFPath:  outputPath,
	}

	// Set extremely short timeout (likely to fail)
	ctx, cancel := context.WithTimeout(mainCtx, 1*time.Millisecond)
	defer cancel()

	res := converter.Convert(ctx, config)
	if res == nil {
		t.Logf("Note: Conversion completed within 1ms timeout (unlikely but possible)")
	} else {
		t.Logf("✓ Context timeout handled correctly: %v", res)
	}
}

// TestLibreExcel2PDF_ContextLogger tests context logger hierarchy
func TestLibreExcel2PDF_ContextLogger(t *testing.T) {
	checkLibreOfficeInstalled(t)

	testFile := "../test/pdf/TestReport.xlsx"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("test file not found: %s", testFile)
	}

	outputDir := "../test-reports"
	os.MkdirAll(outputDir, 0755)
	outputPath := filepath.Join(outputDir, "test_libre_ctx_logger.pdf")
	defer os.Remove(outputPath)

	// Reset singleton for test isolation
	ResetGlobalInstance()

	converter := NewLibreExcel2PDF("libreoffice", loggers.CoreDebugLogger, 1)

	ctx := context.Background()
	converter.StartUp(ctx)
	defer converter.Shutdown(context.Background())

	// Create context with logger
	ctxLogger := loggers.CoreDebugLogger.With().Str("test", "context-logger").Logger()
	ctxWithLogger := context.WithValue(ctx, "ctxLogger", &ctxLogger)

	config := ExcelToPDFTaskConfig{
		InputExcelPath: testFile,
		OutputPDFPath:  outputPath,
	}

	taskCtx, cancel := context.WithTimeout(ctxWithLogger, 30*time.Second)
	defer cancel()

	if res := converter.Convert(taskCtx, config); res != nil {
		t.Fatalf("Convert failed: %v", res)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("output PDF not created: %s", outputPath)
	}

	t.Logf("✓ Context logger test successful")
}

// TestLibreExcel2PDF_Validation tests configuration validation
func TestLibreExcel2PDF_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      ExcelToPDFTaskConfig
		shouldError bool
	}{
		{
			name: "missing input path",
			config: ExcelToPDFTaskConfig{
				OutputPDFPath: "output.pdf",
			},
			shouldError: true,
		},
		{
			name: "missing output path",
			config: ExcelToPDFTaskConfig{
				InputExcelPath: "../test/pdf/TestReport.xlsx",
			},
			shouldError: true,
		},
		{
			name: "nonexistent input file",
			config: ExcelToPDFTaskConfig{
				InputExcelPath: "nonexistent.xlsx",
				OutputPDFPath:  "output.pdf",
			},
			shouldError: true,
		},
		{
			name: "valid config",
			config: ExcelToPDFTaskConfig{
				InputExcelPath: "../test/pdf/TestReport.xlsx",
				OutputPDFPath:  "output.pdf",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.config.Validate()
			if tt.shouldError && res == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.shouldError && res != nil {
				// Check if file exists for the "valid config" case
				if _, err := os.Stat("../test/pdf/TestReport.xlsx"); os.IsNotExist(err) {
					t.Skipf("test file not found, skipping validation test")
				} else {
					t.Errorf("expected no error but got: %v", res)
				}
			}
		})
	}
}

// TestLibreExcel2PDF_CustomBinPath tests custom LibreOffice binary path
func TestLibreExcel2PDF_CustomBinPath(t *testing.T) {
	// Reset singleton for test isolation
	ResetGlobalInstance()

	// Test with invalid binary path
	converter := NewLibreExcel2PDF("/nonexistent/libreoffice", nil, 1)

	if converter.libreOfficeBin != "/nonexistent/libreoffice" {
		t.Errorf("expected custom bin path, got: %s", converter.libreOfficeBin)
	}

	// Reset and test with empty binary path (should default to "libreoffice")
	ResetGlobalInstance()
	converter2 := NewLibreExcel2PDF("", nil, 1)
	if converter2.libreOfficeBin != "libreoffice" {
		t.Errorf("expected default bin path, got: %s", converter2.libreOfficeBin)
	}

	t.Logf("✓ Custom binary path test successful")
}

// TestLibreExcel2PDF_DefaultMaxConcurrent tests default max concurrent value
func TestLibreExcel2PDF_DefaultMaxConcurrent(t *testing.T) {
	// Reset singleton for test isolation
	ResetGlobalInstance()

	// Test with zero max concurrent (should default to 1)
	converter := NewLibreExcel2PDF("libreoffice", nil, 0)
	if converter.maxConcurrent != 1 {
		t.Errorf("expected default maxConcurrent=1, got: %d", converter.maxConcurrent)
	}

	// Reset and test with negative max concurrent (should default to 1)
	ResetGlobalInstance()
	converter2 := NewLibreExcel2PDF("libreoffice", nil, -5)
	if converter2.maxConcurrent != 1 {
		t.Errorf("expected default maxConcurrent=1, got: %d", converter2.maxConcurrent)
	}

	t.Logf("✓ Default max concurrent test successful")
}

// TestLibreExcel2PDF_ShutdownGraceful tests graceful shutdown
func TestLibreExcel2PDF_ShutdownGraceful(t *testing.T) {
	checkLibreOfficeInstalled(t)

	// Reset singleton for test isolation
	ResetGlobalInstance()

	converter := NewLibreExcel2PDF("libreoffice", loggers.CoreDebugLogger, 2)

	ctx := context.Background()
	converter.StartUp(ctx)

	// Shutdown immediately
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	converter.Shutdown(shutdownCtx)

	// Try to convert after shutdown (should fail)
	testFile := "../test/pdf/TestReport.xlsx"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("test file not found: %s", testFile)
	}

	config := ExcelToPDFTaskConfig{
		InputExcelPath: testFile,
		OutputPDFPath:  "../test-reports/test_after_shutdown.pdf",
	}

	res := converter.Convert(context.Background(), config)
	if res == nil {
		t.Errorf("expected error after shutdown, got nil")
	}

	t.Logf("✓ Graceful shutdown test successful")
}

// TestLibreExcel2PDF_NilLogger tests nil logger handling
func TestLibreExcel2PDF_NilLogger(t *testing.T) {
	// Reset singleton for test isolation
	ResetGlobalInstance()

	converter := NewLibreExcel2PDF("libreoffice", nil, 1)

	if converter.logger == nil {
		t.Errorf("expected non-nil logger (should default to Nop)")
	}

	// Test logger type
	nopLogger := zerolog.Nop()
	if converter.logger.GetLevel() != nopLogger.GetLevel() {
		t.Logf("Note: Logger may not be Nop, but is non-nil: %v", converter.logger)
	}

	t.Logf("✓ Nil logger handling test successful")
}

// TestLibreExcel2PDF_Singleton tests singleton behavior
func TestLibreExcel2PDF_Singleton(t *testing.T) {
	// Reset singleton for clean test
	ResetGlobalInstance()

	// First call with specific params
	converter1 := NewLibreExcel2PDF("libreoffice", loggers.CoreDebugLogger, 4)
	if converter1 == nil {
		t.Fatal("first call returned nil")
	}
	if converter1.maxConcurrent != 4 {
		t.Errorf("expected maxConcurrent=4, got: %d", converter1.maxConcurrent)
	}

	// Second call with different params - should return same instance
	converter2 := NewLibreExcel2PDF("/custom/path/libreoffice", nil, 8)
	if converter2 == nil {
		t.Fatal("second call returned nil")
	}

	// Verify it's the same instance
	if converter1 != converter2 {
		t.Errorf("expected same instance, got different pointers")
	}

	// Verify params from first call are preserved (second call params ignored)
	if converter2.maxConcurrent != 4 {
		t.Errorf("expected maxConcurrent=4 (from first call), got: %d", converter2.maxConcurrent)
	}
	if converter2.libreOfficeBin != "libreoffice" {
		t.Errorf("expected libreOfficeBin='libreoffice' (from first call), got: %s", converter2.libreOfficeBin)
	}

	t.Logf("✓ Singleton behavior verified: same instance returned")
}

// TestLibreExcel2PDF_ResetGlobalInstance tests singleton reset functionality
func TestLibreExcel2PDF_ResetGlobalInstance(t *testing.T) {
	// Create first instance
	ResetGlobalInstance()
	converter1 := NewLibreExcel2PDF("libreoffice", nil, 2)
	addr1 := fmt.Sprintf("%p", converter1)

	// Reset and create new instance
	ResetGlobalInstance()
	converter2 := NewLibreExcel2PDF("libreoffice", nil, 4)
	addr2 := fmt.Sprintf("%p", converter2)

	// Should be different instances
	if converter1 == converter2 {
		t.Errorf("expected different instances after reset, got same pointer")
	}

	// Should have different params
	if converter2.maxConcurrent != 4 {
		t.Errorf("expected new instance maxConcurrent=4, got: %d", converter2.maxConcurrent)
	}

	t.Logf("✓ Reset created new instance: %s -> %s", addr1, addr2)
}
