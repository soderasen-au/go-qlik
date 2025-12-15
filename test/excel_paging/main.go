package main

import (
	"flag"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/soderasen-au/go-common/crypto"
	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/qlik/engine"
	"github.com/soderasen-au/go-qlik/report"
)

var (
	appID             = flag.String("app-id", "0569bf97-812d-455b-9fce-83c7bb6a018d", "Application ID (Consumer Sales.qvf)")
	bmID              = flag.String("bm-id", "d3a086a7-633c-4c35-94ff-ec009a602bca", "Bookmark ID (Bookmark2)")
	objID             = flag.String("obj-id", "KnASd", "Object ID to export")
	name              = flag.String("name", "PaginatedReport", "Report name")
	outputFolder      = flag.String("output-folder", ".", "Output folder path")
	certsPath         = flag.String("certs-path", "../certs/sa-win2k25", "Path to certificates directory")
	rowsPerPage       = flag.Int("rows-per-page", 50, "Number of rows per page")
	reportTitle       = flag.String("title", "", "Report title (appears on each page)")
	totalRecordsLabel = flag.String("total-label", "Total Records Found", "Label for total records count")
	showColumnNumbers = flag.Bool("show-col-nums", false, "Show column sequence numbers")
	showSubtotals     = flag.Bool("show-subtotals", false, "Show page subtotals for numeric columns")
	allBorders        = flag.Bool("all-borders", false, "Add borders to all cells")
	outputSelection   = flag.Bool("output-selection", true, "Output current selection")
	help              = flag.Bool("h", false, "Show help message")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", "excel_paging")
		fmt.Fprintf(flag.CommandLine.Output(), "\nPaginated Excel Report Generator\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Exports Qlik objects to paginated Excel sheets.\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExamples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Basic usage with defaults (50 rows per page)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel_paging\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Custom rows per page\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel_paging -rows-per-page 100\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # With report title and subtotals\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel_paging -title \"Sales Report Q4\" -show-subtotals\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Full customization\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel_paging -obj-id \"MyObject\" -rows-per-page 25 -title \"Monthly Report\" \\\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    -show-col-nums -show-subtotals -all-borders -output-folder ./reports\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Specify custom app and bookmark\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./excel_paging -app-id \"your-app-id\" -bm-id \"your-bookmark-id\" -obj-id \"obj1\"\n\n")
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

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	logFile := path.Join(*outputFolder, "excel_paging.log")
	logger, _ := loggers.GetLogger(logFile)

	doc, err := getDoc()
	if err != nil {
		logger.Err(err).Msg("getDoc")
		fmt.Printf("Error: %v\n", err)
		return
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

	// If title is set, use it as the report name for the filename
	reportName := *name
	if *reportTitle != "" {
		// Sanitize the title for use as filename
		reportName = sanitizeFilename(*reportTitle)
	}

	printer := report.NewExcelPagingPrinter(pagingConfig)
	r := report.Report{
		ID:        util.Ptr(reportName),
		Name:      util.Ptr(reportName),
		AppId:     *appID,
		Doc:       doc,
		Target:    report.TARGET_OBJECTS,
		TargetIDs: []string{*objID},
		Headers: []report.CustomHeader{
			{Label: "Generated By", Text: "Excel Paging Driver"},
			{Label: "App ID", Text: *appID},
		},
		OutputCurrentSelection: *outputSelection,
		OutputFormat:           util.Ptr(report.REPORT_FORMAT_XLSX),
		OutputFolder:           util.Ptr(*outputFolder),
		AllBorders:             *allBorders,
		Logger:                 logger,
	}

	res := printer.Print(r)
	if res != nil {
		logger.Err(res).Msg("Print")
		fmt.Printf("Error: %v\n", res)
		return
	}

	result, _ := printer.GetReportResult(reportName)
	if result != nil && result.ReportFile != nil {
		fmt.Printf("Report generated: %s\n", *result.ReportFile)
	}
	logger.Info().Msg("done")
}

// sanitizeFilename removes or replaces characters that are invalid in filenames
func sanitizeFilename(name string) string {
	// Replace characters that are invalid on Windows/Unix
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
