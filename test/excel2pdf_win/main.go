//go:build windows

package main

import (
	"flag"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/soderasen-au/go-common/crypto"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/qlik/engine"
	"github.com/soderasen-au/go-qlik/report"
)

var (
	appID             = flag.String("app-id", "0569bf97-812d-455b-9fce-83c7bb6a018d", "Application ID (Consumer Sales.qvf)")
	bmID              = flag.String("bm-id", "d3a086a7-633c-4c35-94ff-ec009a602bca", "Bookmark ID (Bookmark2)")
	objID             = flag.String("obj-id", "KnASd", "Object ID to export")
	name              = flag.String("name", "Excel2PDFTest", "Report name")
	outputFolder      = flag.String("output-folder", ".", "Output folder path")
	certsPath         = flag.String("certs-path", "../certs/sa-win2k25", "Path to certificates directory")
	rowsPerPage       = flag.Int("rows-per-page", 50, "Number of rows per page")
	reportTitle       = flag.String("title", "Excel to PDF Test Report", "Report title (appears on each page)")
	totalRecordsLabel = flag.String("total-label", "Total Records Found", "Label for total records count")
	showColumnNumbers = flag.Bool("show-col-nums", false, "Show column sequence numbers")
	showSubtotals     = flag.Bool("show-subtotals", false, "Show page subtotals for numeric columns")
	allBorders        = flag.Bool("all-borders", true, "Add borders to all cells")
	outputSelection   = flag.Bool("output-selection", true, "Output current selection")

	// PDF conversion options
	convertToPDF      = flag.Bool("convert-pdf", true, "Convert Excel to PDF after generation")
	pdfOrientation    = flag.String("pdf-orientation", "landscape", "PDF orientation (landscape/portrait)")
	pdfFitToWidth     = flag.Int("pdf-fit-width", 1, "Fit PDF to N pages wide (0 = no fit)")
	pdfFitToHeight    = flag.Int("pdf-fit-height", 0, "Fit PDF to N pages tall (0 = no fit)")
	pdfLeftMargin     = flag.Float64("pdf-left-margin", 0.5, "PDF left margin in inches")
	pdfRightMargin    = flag.Float64("pdf-right-margin", 0.5, "PDF right margin in inches")
	pdfTopMargin      = flag.Float64("pdf-top-margin", 0.75, "PDF top margin in inches")
	pdfBottomMargin   = flag.Float64("pdf-bottom-margin", 0.75, "PDF bottom margin in inches")

	help              = flag.Bool("h", false, "Show help message")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", "excel2pdf_win")
		fmt.Fprintf(flag.CommandLine.Output(), "\nExcel to PDF Test - Component Test for Excel-to-PDF Conversion\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Generates a paginated Excel report and converts it to PDF using Windows COM automation.\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExamples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Generate Excel and convert to PDF with defaults\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel2pdf_win\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Custom page settings\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel2pdf_win -rows-per-page 100 -pdf-fit-width 1 -pdf-fit-height 0\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Portrait orientation\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel2pdf_win -pdf-orientation portrait\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Excel only (no PDF conversion)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel2pdf_win -convert-pdf=false\n\n")
	}
}

func getDoc() (*enigma.Doc, *util.Result) {
	cfg := engine.Config{
		EngineURI:     "wss://sa-win2k25:4747",
		AppID:         *appID,
		UserName:      "qliksa",
		UserDirectory: "sa-win2k25",
		AuthMode:      engine.AUTH_MODE_CERT,
		ServerType:    engine.ST_ON_PREM,
		Certs: crypto.Certificates{
			ClientFile:    filepath.Join(*certsPath, "client.pem"),
			ClientkeyFile: filepath.Join(*certsPath, "client_key.pem"),
			CAFile:        filepath.Join(*certsPath, "root.pem"),
		},
	}
	res := cfg.QCSEngineURIAppendAppID(*appID)
	if res != nil {
		return nil, res.With("AppendAppID")
	}

	conn, err := engine.NewConn(cfg)
	if err != nil {
		return nil, util.Error("NewConn", err)
	}

	doc, err := conn.Global.OpenDoc(engine.ConnCtx, *appID, "", "", "", false)
	if err != nil {
		return nil, util.Error("OpenDoc", err)
	}

	if *bmID != "" {
		if ok, err := doc.ApplyBookmark(engine.ConnCtx, *bmID); err != nil {
			return nil, util.Error("ApplyBookmark", err)
		} else if !ok {
			return nil, util.MsgError("ApplyBookmark", "failed")
		}
	}

	return doc, nil
}

// generatePagedExcel creates a paginated Excel report and returns the file path
func generatePagedExcel(logger *zerolog.Logger) (string, *util.Result) {
	fmt.Println("Step 1: Connecting to Qlik Sense and generating Excel report...")

	doc, err := getDoc()
	if err != nil {
		return "", err.With("getDoc")
	}
	defer doc.DisconnectFromServer()

	// Create paging config
	pagingConfig := report.ExcelPagingConfig{
		RowsPerPage:       *rowsPerPage,
		ReportTitle:       *reportTitle,
		TotalRecordsLabel: *totalRecordsLabel,
		ShowColumnNumbers: *showColumnNumbers,
		ShowSubtotals:     *showSubtotals,
	}

	reportName := sanitizeFilename(*name)

	printer := report.NewExcelPagingPrinter(pagingConfig)
	r := report.Report{
		ID:        util.Ptr(reportName),
		Name:      util.Ptr(reportName),
		AppId:     *appID,
		Doc:       doc,
		Target:    report.TARGET_OBJECTS,
		TargetIDs: []string{*objID},
		Headers: []report.CustomHeader{
			{Label: "Generated By", Text: "Excel to PDF Test"},
			{Label: "App ID", Text: *appID},
			{Label: "Test Purpose", Text: "Component test for Excel-to-PDF conversion"},
		},
		OutputCurrentSelection: *outputSelection,
		OutputFormat:           util.Ptr(report.REPORT_FORMAT_XLSX),
		OutputFolder:           util.Ptr(*outputFolder),
		AllBorders:             *allBorders,
		Logger:                 logger,
	}

	res := printer.Print(r)
	if res != nil {
		return "", res.With("Print")
	}

	result, _ := printer.GetReportResult(reportName)
	if result == nil || result.ReportFile == nil {
		return "", util.MsgError("generatePagedExcel", "no report file generated")
	}

	fmt.Printf("✓ Excel report generated: %s\n", *result.ReportFile)
	return *result.ReportFile, nil
}

// convertExcelToPDF converts the Excel file to PDF using Windows COM automation
func convertExcelToPDF(excelPath string, logger *zerolog.Logger) (string, *util.Result) {
	fmt.Println("\nStep 2: Converting Excel to PDF using Windows COM automation...")

	// Determine orientation constant
	orientation := 2 // xlLandscape
	if *pdfOrientation == "portrait" {
		orientation = 1 // xlPortrait
	}

	// Build output PDF path (same name, .pdf extension)
	pdfPath := strings.TrimSuffix(excelPath, filepath.Ext(excelPath)) + ".pdf"

	// Create conversion config
	config := report.ExcelToPDFWinConfig{
		InputExcelPath:   excelPath,
		OutputPDFPath:    pdfPath,
		PaperSize:        9, // xlPaperA4
		Orientation:      orientation,
		FitToWidth:       *pdfFitToWidth,
		FitToHeight:      *pdfFitToHeight,
		LeftMargin:       *pdfLeftMargin,
		RightMargin:      *pdfRightMargin,
		TopMargin:        *pdfTopMargin,
		BottomMargin:     *pdfBottomMargin,
		ExportMultiplePDFs: false,
		IncludeDocProperties: true,
		OpenAfterPublish: false,
		Logger:           logger,
	}

	converter, res := report.NewExcelToPDFWin(config)
	if res != nil {
		return "", res.With("NewExcelToPDFWin")
	}

	res = converter.Convert()
	if res != nil {
		return "", res.With("Convert")
	}

	fmt.Printf("✓ PDF generated: %s\n", pdfPath)
	return pdfPath, nil
}

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	logFile := path.Join(*outputFolder, "excel2pdf_win.log")
	logger, _ := loggers.GetLogger(logFile)

	fmt.Println("=== Excel to PDF Component Test ===")
	fmt.Println()

	// Step 1: Generate paged Excel file
	excelPath, err := generatePagedExcel(logger)
	if err != nil {
		logger.Err(err).Msg("generatePagedExcel failed")
		fmt.Printf("\n✗ Error generating Excel: %v\n", err)
		return
	}

	// Step 2: Convert to PDF (if enabled)
	if *convertToPDF {
		pdfPath, err := convertExcelToPDF(excelPath, logger)
		if err != nil {
			logger.Err(err).Msg("convertExcelToPDF failed")
			fmt.Printf("\n✗ Error converting to PDF: %v\n", err)
			return
		}

		fmt.Println("\n=== Test Complete ===")
		fmt.Printf("Excel file: %s\n", excelPath)
		fmt.Printf("PDF file:   %s\n", pdfPath)
	} else {
		fmt.Println("\n=== Test Complete (Excel only) ===")
		fmt.Printf("Excel file: %s\n", excelPath)
	}

	logger.Info().Msg("excel2pdf_win test completed successfully")
}

// sanitizeFilename removes or replaces characters that are invalid in filenames
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"<", "_",
		">", "_",
		":", "_",
		"\"", "_",
		"/", "_",
		"\\", "_",
		"|", "_",
		"?", "_",
		"*", "_",
	)
	return replacer.Replace(name)
}
