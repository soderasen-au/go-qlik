//go:build !windows

package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/soderasen-au/go-common/loggers"

	"github.com/soderasen-au/go-qlik/report"
)

var (
	inputExcel     = flag.String("input", "../../test/pdf/TestReport.xlsx", "Input Excel file path")
	outputPDF      = flag.String("output", "test-reports/TestReport_libre.pdf", "Output PDF file path")
	libreOfficeBin = flag.String("libreoffice", "libreoffice", "Path to LibreOffice binary")
	maxConcurrent  = flag.Int("max-concurrent", 2, "Maximum number of concurrent conversions")
	timeout        = flag.Int("timeout", 30, "Conversion timeout in seconds")
	help           = flag.Bool("h", false, "Show help message")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", "excel_to_pdf_libre")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExamples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Use defaults\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel_to_pdf_libre\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Custom input and output\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel_to_pdf_libre -input myfile.xlsx -output output.pdf\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Custom LibreOffice path (macOS)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel_to_pdf_libre -libreoffice /Applications/LibreOffice.app/Contents/MacOS/soffice\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Custom LibreOffice path (Linux)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel_to_pdf_libre -libreoffice /usr/bin/libreoffice\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Increase concurrent conversions\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel_to_pdf_libre -max-concurrent 4\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Custom timeout\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel_to_pdf_libre -timeout 60\n\n")
	}
}

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	// Setup logger
	logFile := filepath.Join(filepath.Dir(*outputPDF), "excel_to_pdf_libre.log")
	logger, _ := loggers.GetLogger(logFile)

	logger.Info().
		Str("input", *inputExcel).
		Str("output", *outputPDF).
		Str("libreoffice", *libreOfficeBin).
		Int("max_concurrent", *maxConcurrent).
		Int("timeout_sec", *timeout).
		Msg("starting Excel to PDF conversion test")

	// Create converter
	converter := report.NewLibreExcel2PDF(*libreOfficeBin, logger, *maxConcurrent)

	// Start up converter
	ctx := context.Background()
	converter.StartUp(ctx)
	defer converter.Shutdown(context.Background())

	logger.Info().Msg("converter started successfully")

	// Prepare conversion config
	config := report.ExcelToPDFTaskConfig{
		InputExcelPath: *inputExcel,
		OutputPDFPath:  *outputPDF,
		Logger:         logger,
	}

	// Create context with timeout
	convCtx, cancel := context.WithTimeout(ctx, time.Duration(*timeout)*time.Second)
	defer cancel()

	// Perform conversion
	logger.Info().Msg("starting conversion")
	startTime := time.Now()

	result := converter.Convert(convCtx, config)

	elapsed := time.Since(startTime)

	if result != nil {
		logger.Err(result).Msg("conversion failed")
		fmt.Printf("❌ Conversion failed: %v\n", result)
		return
	}

	logger.Info().
		Dur("elapsed", elapsed).
		Msg("conversion completed successfully")

	fmt.Printf("✅ Conversion successful!\n")
	fmt.Printf("   Input:   %s\n", *inputExcel)
	fmt.Printf("   Output:  %s\n", *outputPDF)
	fmt.Printf("   Elapsed: %v\n", elapsed)
}
