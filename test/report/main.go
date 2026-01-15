package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/crypto"
	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-common/util"
	"gopkg.in/yaml.v3"

	"github.com/soderasen-au/go-qlik/qlik/engine"
	"github.com/soderasen-au/go-qlik/report"
)

var (
	configFile = flag.String("config", "config.yaml", "Path to YAML configuration file")
	help       = flag.Bool("h", false, "Show help message")
)

type SystemConfig struct {
	LogFolder    string `yaml:"log_folder"`
	OutputFolder string `yaml:"output_folder"`
}

type ReportConfig struct {
	Driver          string                    `yaml:"driver"`
	Format          string                    `yaml:"format"`
	Name            string                    `yaml:"name"`
	AppID           string                    `yaml:"app_id"`
	BookmarkID      string                    `yaml:"bookmark_id"`
	Target          string                    `yaml:"target"`
	TargetIDs       []string                  `yaml:"target_ids"`
	Orientation     string                    `yaml:"orientation"`
	AllBorders      bool                      `yaml:"all_borders"`
	OutputSelection bool                      `yaml:"output_selection"`
	OutputOffset    *enigma.Rect              `yaml:"output_offset"`
	ExcelPaging     *report.ExcelPagingConfig `yaml:"excel_paging"`
	ExcelToPDF      *ExcelToPDFConfig         `yaml:"excel_to_pdf"`
	Headers         []report.CustomHeader     `yaml:"headers"`
	HeadersOffset   *enigma.Rect              `yaml:"headers_offset"`
	Footers         []report.CustomHeader     `yaml:"footers"`
	FootersOffset   *enigma.Rect              `yaml:"footers_offset"`
	Legends         []report.Legend           `yaml:"legends"`
	LegendOffset    *enigma.Rect              `yaml:"legend_offset"`
}

type ExcelToPDFConfig struct {
	LibreOfficeBin string `yaml:"libreoffice_bin"`
	MaxConcurrent  int    `yaml:"max_concurrent"`
	Timeout        int    `yaml:"timeout"`
	Profile        string `yaml:"profile"`
}

type EngineConfig struct {
	EngineURI     string `yaml:"engine_uri"`
	UserName      string `yaml:"user_name"`
	UserDirectory string `yaml:"user_directory"`
	AuthMode      string `yaml:"auth_mode"`
	ServerType    string `yaml:"server_type"`
	CertsPath     string `yaml:"certs_path"`
}

type Config struct {
	System SystemConfig `yaml:"system"`
	Report ReportConfig `yaml:"report"`
	Engine EngineConfig `yaml:"engine"`
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", "report-test")
		fmt.Fprintf(flag.CommandLine.Output(), "\nUnified Report Test Tool\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Tests all report drivers (excel, pdf, excel_paging, pdf_paging, excel_to_pdf_libre) via YAML config.\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExamples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Use default config.yaml\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./report-test\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Specify custom config file\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ./report-test -config my-config.yaml\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\nConfig File Structure:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  system:           # Application settings\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    log_folder:     # Where to write logs\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    output_folder:  # Where to write reports\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  engine:           # Qlik Engine connection settings\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    engine_uri:     # WebSocket URI\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    auth_mode:      # cert, jwt, or desktop\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    certs_path:     # Path to certificate files\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  report:           # Report generation settings\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    driver:         # built_in or sense\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    format:         # xlsx, paged_xlsx, pdf, csv, tsv\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    excel_paging:   # Config for paged_xlsx format\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    excel_to_pdf:   # Config for Excel->PDF conversion\n\n")
	}
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}

	// Set defaults
	if cfg.System.LogFolder == "" {
		cfg.System.LogFolder = "."
	}
	if cfg.System.OutputFolder == "" {
		cfg.System.OutputFolder = "."
	}
	if cfg.Report.Driver == "" {
		cfg.Report.Driver = report.DRIVER_BUILT_IN
	}
	if cfg.Report.Format == "" {
		cfg.Report.Format = string(report.REPORT_FORMAT_XLSX)
	}
	if cfg.Report.Target == "" {
		cfg.Report.Target = report.TARGET_OBJECTS
	}

	return &cfg, nil
}

func getDoc(cfg *Config, logger *zerolog.Logger) (*enigma.Doc, *util.Result) {
	engineCfg := engine.Config{
		EngineURI:     cfg.Engine.EngineURI,
		AppID:         cfg.Report.AppID,
		UserName:      cfg.Engine.UserName,
		UserDirectory: cfg.Engine.UserDirectory,
	}

	switch cfg.Engine.AuthMode {
	case "cert":
		engineCfg.AuthMode = engine.AUTH_MODE_CERT
	case "jwt":
		engineCfg.AuthMode = engine.AUTH_MODE_JWT
	case "desktop":
		engineCfg.AuthMode = engine.AUTH_MODE_DESKTOP
	default:
		engineCfg.AuthMode = engine.AUTH_MODE_CERT
	}

	switch cfg.Engine.ServerType {
	case "on_prem":
		engineCfg.ServerType = engine.ST_ON_PREM
	case "cloud":
		engineCfg.ServerType = engine.ST_CLOUD
	default:
		engineCfg.ServerType = engine.ST_ON_PREM
	}

	if engineCfg.AuthMode == engine.AUTH_MODE_CERT && cfg.Engine.CertsPath != "" {
		engineCfg.Certs = crypto.Certificates{
			ClientFile:    filepath.Join(cfg.Engine.CertsPath, "client.pem"),
			ClientkeyFile: filepath.Join(cfg.Engine.CertsPath, "client_key.pem"),
			CAFile:        filepath.Join(cfg.Engine.CertsPath, "root.pem"),
		}
	}

	res := engineCfg.QCSEngineURIAppendAppID(cfg.Report.AppID)
	if res != nil {
		return nil, res.LogWith(logger, "AppendAppID")
	}
	conn, err := engine.NewConn(engineCfg)
	if err != nil {
		return nil, util.LogError(logger, "NewConn", err)
	}
	doc, err := conn.Global.OpenDoc(engine.ConnCtx, cfg.Report.AppID, "", "", "", false)
	if err != nil {
		return nil, util.LogError(logger, "OpenDoc", err)
	}

	if cfg.Report.BookmarkID != "" {
		if ok, err := doc.ApplyBookmark(engine.ConnCtx, cfg.Report.BookmarkID); err != nil {
			return nil, util.LogError(logger, "ApplyBookmark", err)
		} else if !ok {
			return nil, util.LogMsgError(logger, "ApplyBookmark", "failed")
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

	cfg, err := loadConfig(*configFile)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	logFile := filepath.Join(cfg.System.LogFolder, "report-test.log")
	logger, _ := loggers.GetLogger(logFile)
	logger.Info().
		Str("config", *configFile).
		Str("driver", cfg.Report.Driver).
		Str("format", cfg.Report.Format).
		Msg("starting report test")

	doc, res := getDoc(cfg, logger)
	if res != nil {
		logger.Err(res).Msg("getDoc")
		fmt.Printf("Error: %v\n", res)
		os.Exit(1)
	}
	defer doc.DisconnectFromServer()

	var printer report.IReportPrinter
	reportFormat := report.ReportFormat(cfg.Report.Format)

	if reportFormat.IsPagedExcel() {
		if cfg.Report.ExcelPaging == nil {
			cfg.Report.ExcelPaging = &report.ExcelPagingConfig{
				RowsPerPage: 50,
			}
		}
		printer = report.NewExcelPagingPrinter(*cfg.Report.ExcelPaging)
		logger.Info().Msg("using Excel paging printer")
	} else {
		printer = report.NewBuiltInReportPrinter()
		logger.Info().Msg("using built-in report printer")
	}

	r := report.Report{
		ID:                     util.Ptr(cfg.Report.Name),
		Name:                   util.Ptr(cfg.Report.Name),
		AppId:                  cfg.Report.AppID,
		Doc:                    doc,
		Target:                 cfg.Report.Target,
		TargetIDs:              cfg.Report.TargetIDs,
		Headers:                cfg.Report.Headers,
		HeadersOffset:          cfg.Report.HeadersOffset,
		Footers:                cfg.Report.Footers,
		FootersOffset:          cfg.Report.FootersOffset,
		Legends:                cfg.Report.Legends,
		LegendOffset:           cfg.Report.LegendOffset,
		OutputCurrentSelection: cfg.Report.OutputSelection,
		OutputFormat:           util.Ptr(reportFormat),
		OutputFolder:           util.Ptr(cfg.System.OutputFolder),
		AllBorders:             cfg.Report.AllBorders,
		OutputOffset:           cfg.Report.OutputOffset,
		Logger:                 logger,
	}
	if cfg.Report.Driver != "" {
		r.Driver = util.Ptr(cfg.Report.Driver)
	}
	if reportFormat.IsPdf() && cfg.Report.Orientation != "" {
		r.OutputPDFOrientation = util.Ptr(cfg.Report.Orientation)
	}

	res = printer.Print(r)
	if res != nil {
		logger.Err(res).Msg("Print")
		fmt.Printf("Error: %v\n", res)
		os.Exit(1)
	}

	result, _ := printer.GetReportResult(cfg.Report.Name)
	if result != nil && result.ReportFile != nil {
		fmt.Printf("Report generated: %s\n", *result.ReportFile)
		logger.Info().Str("file", *result.ReportFile).Msg("report generated")

		// Excel to PDF conversion if configured
		if reportFormat.IsExcel() || reportFormat.IsPagedExcel() {
			if cfg.Report.ExcelToPDF != nil {
				logger.Info().Msg("starting Excel to PDF conversion")
				convertExcelToPDF(cfg, *result.ReportFile, logger)
			}
		}
	}

	logger.Info().Msg("done")
}

func convertExcelToPDF(cfg *Config, excelFile string, logger *zerolog.Logger) {
	if cfg.Report.ExcelToPDF == nil {
		return
	}

	e2pCfg := cfg.Report.ExcelToPDF
	if e2pCfg.LibreOfficeBin == "" {
		e2pCfg.LibreOfficeBin = "libreoffice"
	}
	if e2pCfg.MaxConcurrent == 0 {
		e2pCfg.MaxConcurrent = 2
	}
	if e2pCfg.Timeout == 0 {
		e2pCfg.Timeout = 30
	}

	pdfFile := excelFile[:len(excelFile)-len(filepath.Ext(excelFile))] + ".pdf"
	converter := report.NewLibreExcel2PDF(e2pCfg.LibreOfficeBin, logger, e2pCfg.MaxConcurrent, e2pCfg.Profile)

	ctx := context.Background()
	converter.StartUp(ctx)
	defer converter.Shutdown(context.Background())

	taskCfg := report.ExcelToPDFTaskConfig{
		InputExcelPath: excelFile,
		OutputPDFPath:  pdfFile,
		Logger:         logger,
	}

	res := converter.Convert(ctx, taskCfg)
	if res != nil {
		logger.Err(res).Msg("Excel to PDF conversion failed")
		fmt.Printf("Excel to PDF conversion failed: %v\n", res)
		return
	}

	logger.Info().Str("pdf", pdfFile).Msg("Excel to PDF conversion completed")
	fmt.Printf("PDF generated: %s\n", pdfFile)
}
