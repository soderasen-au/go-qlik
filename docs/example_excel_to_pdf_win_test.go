//go:build windows

package report_test

import (
	"fmt"

	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-qlik/report"
)

// ExampleExcelToPDFWin_basic demonstrates basic usage
func ExampleExcelToPDFWin_basic() {
	config := report.ExcelToPDFWinConfig{
		InputExcelPath: "C:/reports/sales.xlsx",
		OutputPDFPath:  "C:/reports/sales.pdf",
		Logger:         loggers.CoreDebugLogger,
	}

	converter, res := report.NewExcelToPDFWin(config)
	if res != nil {
		fmt.Printf("Error: %v\n", res)
		return
	}

	if res := converter.Convert(); res != nil {
		fmt.Printf("Conversion failed: %v\n", res)
		return
	}

	fmt.Println("✓ PDF created successfully")
}

// ExampleExcelToPDFWin_multipleSheets demonstrates exporting to separate PDFs
func ExampleExcelToPDFWin_multipleSheets() {
	config := report.ExcelToPDFWinConfig{
		InputExcelPath:     "C:/reports/quarterly.xlsx",
		OutputPDFPath:      "C:/reports/output/",
		ExportMultiplePDFs: true,
		Logger:             loggers.CoreDebugLogger,
	}

	converter, res := report.NewExcelToPDFWin(config)
	if res != nil {
		fmt.Printf("Error: %v\n", res)
		return
	}

	if res := converter.Convert(); res != nil {
		fmt.Printf("Conversion failed: %v\n", res)
		return
	}

	fmt.Println("✓ Multiple PDFs created successfully")
}

// ExampleExcelToPDFWin_customSettings demonstrates advanced configuration
func ExampleExcelToPDFWin_customSettings() {
	config := report.ExcelToPDFWinConfig{
		InputExcelPath: "C:/reports/data.xlsx",
		OutputPDFPath:  "C:/reports/data_portrait.pdf",
		Password:       "secret123",

		// Portrait with tight margins
		Orientation:  1, // xlPortrait
		LeftMargin:   0.25,
		RightMargin:  0.25,
		TopMargin:    0.5,
		BottomMargin: 0.5,

		// Fit to 1 page wide, unlimited height
		FitToWidth:  1,
		FitToHeight: 0,

		// Export specific sheets only
		SheetNames: []string{"Summary", "Details"},

		Logger: loggers.CoreDebugLogger,
	}

	converter, res := report.NewExcelToPDFWin(config)
	if res != nil {
		fmt.Printf("Error: %v\n", res)
		return
	}

	if res := converter.Convert(); res != nil {
		fmt.Printf("Conversion failed: %v\n", res)
		return
	}

	fmt.Println("✓ Custom PDF created successfully")
}
