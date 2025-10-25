package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/soderasen-au/go-common/crypto"
	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/qlik/engine"
	"github.com/soderasen-au/go-qlik/report"
)

var (
	appID        = flag.String("app-id", "0569bf97-812d-455b-9fce-83c7bb6a018d", "Application ID (Consumer Sales.qvf)")
	bmID         = flag.String("bm-id", "35929e08-2288-420b-8f2d-fee01c8e7a94", "Bookmark ID (Bookmark1)")
	format       = flag.String("format", "pdf", "Output format: pdf, xlsx, csv, tsv")
	orientation  = flag.String("orientation", "portrait", "PDF orientation: portrait, landscape (only for PDF format)")
	name         = flag.String("name", "TestPdfStackObject", "Report name")
	outputFolder = flag.String("output-folder", ".", "Output folder path")
	certsPath    = flag.String("certs-path", "../certs/sa-win2k25", "Path to certificates directory")
	help         = flag.Bool("h", false, "Show help message")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", "pdfprinter")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExamples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Use defaults (PDF format, portrait orientation, current directory)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./pdfprinter\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Specify custom app-id\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./pdfprinter -app-id \"your-app-id\"\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Specify both app-id and bookmark-id\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./pdfprinter -app-id \"your-app-id\" -bm-id \"your-bookmark-id\"\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Generate Excel report\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./pdfprinter -format xlsx\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Generate CSV report\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./pdfprinter -format csv\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Generate TSV report\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./pdfprinter -format tsv\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # PDF with landscape orientation\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./pdfprinter -format pdf -orientation landscape\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Specify custom report name\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./pdfprinter -name \"MonthlySalesReport\"\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Specify custom output folder\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./pdfprinter -output-folder \"/tmp/reports\"\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Custom name and output folder\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./pdfprinter -name \"Q4Report\" -output-folder \"./output\"\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Specify custom certificates path\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./pdfprinter -certs-path \"/home/sa/certs/sa-win2k25\"\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Combine all parameters\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./pdfprinter -app-id \"your-app-id\" -bm-id \"your-bookmark-id\" -format pdf -orientation landscape -name \"MyReport\" -output-folder \"./reports\" -certs-path \"/path/to/certs\"\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Note: orientation is ignored for non-PDF formats\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./pdfprinter -format xlsx -orientation landscape  # orientation has no effect\n")
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

	if ok, err := doc.ApplyBookmark(engine.ConnCtx, *bmID); err != nil {
		return nil, util.Error("ApplyBookmark", err)
	} else if !ok {
		return nil, util.MsgError("ApplyBookmark", "failed")
	}

	return doc, nil
}

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	logger, _ := loggers.GetLogger("pdf.log")
	doc, err := getDoc()
	if err != nil {
		logger.Err(err).Msg("getDoc")
		return
	}
	defer doc.DisconnectFromServer()

	printer := report.NewBuiltInReportPrinter()
	r := report.Report{
		ID:        util.Ptr(*name),
		Name:      util.Ptr(*name),
		AppId:     *appID,
		Doc:       doc,
		Target:    report.TARGET_OBJECTS,
		TargetIDs: []string{"KnASd"},
		Headers: []report.CustomHeader{
			{Label: "label1", Text: "Text1"},
			{Label: "label2", Text: "Text2"},
		},
		OutputCurrentSelection: true,
		OutputFormat:           util.Ptr(report.ReportFormat(*format)),
		OutputFolder:           util.Ptr(*outputFolder),
		OutputOffset: &enigma.Rect{
			Left: 3,
			Top:  3,
		},
		Logger: logger,
	}

	if *format == "pdf" {
		r.OutputPDFOrientation = util.Ptr(*orientation)
	}

	res := printer.Print(r)
	if res != nil {
		logger.Err(res).Msg("Print")
	}
	logger.Info().Msg("done")
}
