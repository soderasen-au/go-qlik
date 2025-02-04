package report

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-common/util"
)

type ReportFormat string

const (
	REPORT_FORMAT_XLSX ReportFormat = "xlsx"
	REPORT_FORMAT_CSV  ReportFormat = "csv"
	REPORT_FORMAT_TSV  ReportFormat = "tsv"
	REPORT_FORMAT_PDF  ReportFormat = "pdf"

	TARGET_OBJECTS string = "objects"
	TARGET_SHEET   string = "sheet"

	DRIVER_SENSE    string = "sense"
	DRIVER_BUILT_IN string = "built_in"

	COL_TYPE string = "static"
)

func (f ReportFormat) IsExcel() bool {
	return f == REPORT_FORMAT_XLSX
}

func (f ReportFormat) IsCsv() bool {
	return f == REPORT_FORMAT_CSV || f == REPORT_FORMAT_TSV
}

func (f ReportFormat) IsTsv() bool {
	return f == REPORT_FORMAT_TSV
}

func (f ReportFormat) IsPdf() bool {
	return f == REPORT_FORMAT_PDF
}

func (f ReportFormat) IsValid() bool {
	return f.IsExcel() || f.IsCsv() || f.IsPdf()
}

func (f *ReportFormat) MaybeDefault() {
	if !f.IsValid() {
		*f = REPORT_FORMAT_XLSX
	}
}

type ReportPrinterBase struct {
	ReportResults map[string]*ReportResult //report-id -> report-results
	Doc           *enigma.Doc
	R             Report
	Logger        *zerolog.Logger
}

type IReportPrinter interface {
	Print(r Report) *util.Result
	GetReportResult(id string) (*ReportResult, *util.Result)
}

type CustomHeader struct {
	Label string `json:"label,omitempty"`
	Text  string `json:"text,omitempty"`
}

type ColumnHeaderFormat struct {
	Order       int    `json:"order"`
	Label       string `json:"label"`
	FgColor     string `json:"fg_color"`
	BgColor     string `json:"bg_color"`
	NumFmt      string `json:"num_fmt"`
	DateFmt     string `json:"date_fmt"`
	ColumnType  string `json:"column_type,omitempty"`
	StaticValue string `json:"static_value"`
}

// user has to apply any needed selection before printing report
type Report struct {
	ID    *string `json:"id,omitempty" yaml:"id,omitempty" bson:"id,omitempty"`
	Name  *string `json:"name,omitempty" yaml:"name,omitempty" bson:"name,omitempty"`
	IsSub bool    `json:"is_sub,omitempty" yaml:"is_sub,omitempty" bson:"is_sub,omitempty"`

	// report target
	// `Target` can be `objects` or `sheet`
	// `TargetIDs` contains either:
	//  - array of object ids, when `Target` is `objects`
	//  - or TargetIDs[0] = sheetID, when `Target` is `sheet`
	Doc       *enigma.Doc `json:"-,omitempty" yaml:"-,omitempty" bson:"-,omitempty"`
	AppId     string      `json:"app_id,omitempty" yaml:"app_id,omitempty" bson:"app_id,omitempty"`
	Target    string      `json:"target,omitempty" yaml:"target,omitempty" bson:"target,omitempty"`
	TargetIDs []string    `json:"target_ids,omitempty" yaml:"target_ids,omitempty" bson:"target_ids,omitempty"`

	// layout
	Headers                []CustomHeader                `json:"headers,omitempty" yaml:"headers,omitempty" bson:"headers,omitempty"`
	OptionalTargetTitles   map[string]string             `json:"optional_target_titles,omitempty" yaml:"optional_target_titles,omitempty" bson:"optional_target_titles,omitempty"`
	OutputCurrentSelection bool                          `json:"output_current_selection,omitempty" yaml:"output_current_selection,omitempty" bson:"output_current_selection,omitempty"`
	CurrentSelectionOrder  map[string]int                `json:"current_selection_order" yaml:"current_selection_order" bson:"current_selection_order"`
	ColumnHeaderFormats    map[string]ColumnHeaderFormat `json:"column_header_formats,omitempty" yaml:"column_header_formats,omitempty" bson:"column_header_formats,omitempty"` // only supports stack object

	// output
	Driver       *string       `json:"driver,omitempty" yaml:"driver,omitempty" bson:"driver,omitempty"`
	OutputFormat *ReportFormat `json:"output_format,omitempty" yaml:"output_format,omitempty" bson:"output_format,omitempty"`
	OutputFolder *string       `json:"output_folder,omitempty" yaml:"output_folder,omitempty" bson:"output_folder,omitempty"`
	OutputOffset *enigma.Rect  `json:"output_offset,omitempty" yaml:"output_offset,omitempty" bson:"output_offset,omitempty"`

	// logging
	LogFolder *string         `json:"log_folder,omitempty" yaml:"log_folder,omitempty" bson:"log_folder,omitempty"`
	Logger    *zerolog.Logger `json:"-" yaml:"-" bson:"-"`
}

func (r Report) IsValid() bool {
	if r.Doc == nil {
		return false
	}

	if r.ID == nil || r.OutputFormat == nil || r.OutputFolder == nil {
		return false
	}

	if !r.OutputFormat.IsValid() {
		return false
	}

	if r.AppId == "" {
		return false
	}

	if r.Target == "sheet" && len(r.TargetIDs) != 1 {
		return false
	}

	if r.Target == "objects" && len(r.TargetIDs) < 1 {
		return false
	}

	return true
}

func (r *Report) Validate() *util.Result {
	if r.Doc == nil {
		return util.MsgError("ValidateReport", "invalid engine connection")
	}

	if r.ID == nil {
		r.ID = new(string)
		*r.ID = fmt.Sprintf("%s-%s", r.AppId, time.Now().Format("20060102150405"))
	}

	if r.AppId == "" {
		return util.MsgError("ValidateReport", "No app id")
	}

	switch r.Target = strings.ToLower(r.Target); r.Target {
	case "sheet":
		if len(r.TargetIDs) != 1 {
			return util.MsgError("ValidateReport", "supports only 1 sheet per Report")
		}
	case "objects":
		if len(r.TargetIDs) < 1 {
			return util.MsgError("ValidateReport", "no object in Report")
		}
	default:
		return util.MsgError("ValidateReport", "invalid target, support only `sheet` and `objects`")
	}

	if r.OutputFormat == nil {
		r.OutputFormat = new(ReportFormat)
		r.OutputFormat.MaybeDefault()
	}

	if r.OutputFolder == nil {
		r.OutputFolder = new(string)
	}

	if r.LogFolder == nil {
		r.LogFolder = new(string)
	}

	return nil
}

type ReportResult struct {
	ID          string          `json:"id,omitempty" yaml:"id,omitempty"`
	Result      *util.Result    `json:"result,omitempty" yaml:"result,omitempty" bson:"result,omitempty"`
	ReportFile  *string         `json:"report_file,omitempty" yaml:"report_file,omitempty" bson:"report_file,omitempty"`
	LogFile     *string         `json:"log_file,omitempty" yaml:"log_file,omitempty" bson:"log_file,omitempty"`
	Logger      *zerolog.Logger `json:"-,omitempty" yaml:"-,omitempty" bson:"-"`
	PrintedRows int             `json:"printed_rows,omitempty" yaml:"printed_rows,omitempty"`
}

func NewReportResult(r Report) (*ReportResult, *util.Result) {
	if !r.IsValid() {
		return nil, util.MsgError("Check", "invalid report")
	}

	rr := ReportResult{}

	var rf string
	if r.Name != nil && len(*r.Name) > 0 {
		rn := strings.ReplaceAll(*r.Name, "/", "_")
		rn = strings.ReplaceAll(rn, "\\", "_")
		rf = filepath.Join(util.MaybeNil(r.OutputFolder), fmt.Sprintf("%s.%s", rn, *r.OutputFormat))
	} else {
		rf = filepath.Join(util.MaybeNil(r.OutputFolder), fmt.Sprintf("%s.%s", util.MaybeNil(r.ID), *r.OutputFormat))
	}
	rr.ReportFile = &rf

	if r.Logger != nil {
		rr.Logger = r.Logger
	} else {
		lf := filepath.Join(util.MaybeNil(r.LogFolder), fmt.Sprintf("log-%s.%s", util.MaybeNil(r.ID), "log"))
		rr.LogFile = &lf
		logger, err := loggers.GetLogger(lf)
		if err != nil {
			return nil, util.Error("GetLogger", err)
		}
		rr.Logger = logger
	}

	return &rr, nil
}
