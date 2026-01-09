//go:build !windows

package report

import (
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"
)

// ExcelToPDFWinConfig is a stub for non-Windows platforms
// This type exists only for API compatibility - actual functionality requires Windows
type ExcelToPDFWinConfig struct {
	InputExcelPath     string          `json:"input_excel_path" yaml:"input_excel_path" bson:"input_excel_path"`
	OutputPDFPath      string          `json:"output_pdf_path" yaml:"output_pdf_path" bson:"output_pdf_path"`
	Password           string          `json:"password,omitempty" yaml:"password,omitempty" bson:"password,omitempty"`
	SheetNames         []string        `json:"sheet_names,omitempty" yaml:"sheet_names,omitempty" bson:"sheet_names,omitempty"`
	SheetIndices       []int           `json:"sheet_indices,omitempty" yaml:"sheet_indices,omitempty" bson:"sheet_indices,omitempty"`
	PaperSize          int             `json:"paper_size,omitempty" yaml:"paper_size,omitempty" bson:"paper_size,omitempty"`
	Orientation        int             `json:"orientation,omitempty" yaml:"orientation,omitempty" bson:"orientation,omitempty"`
	FitToWidth         int             `json:"fit_to_width,omitempty" yaml:"fit_to_width,omitempty" bson:"fit_to_width,omitempty"`
	FitToHeight        int             `json:"fit_to_height,omitempty" yaml:"fit_to_height,omitempty" bson:"fit_to_height,omitempty"`
	LeftMargin         float64         `json:"left_margin,omitempty" yaml:"left_margin,omitempty" bson:"left_margin,omitempty"`
	RightMargin        float64         `json:"right_margin,omitempty" yaml:"right_margin,omitempty" bson:"right_margin,omitempty"`
	TopMargin          float64         `json:"top_margin,omitempty" yaml:"top_margin,omitempty" bson:"top_margin,omitempty"`
	BottomMargin       float64         `json:"bottom_margin,omitempty" yaml:"bottom_margin,omitempty" bson:"bottom_margin,omitempty"`
	PrintArea          string          `json:"print_area,omitempty" yaml:"print_area,omitempty" bson:"print_area,omitempty"`
	ExportMultiplePDFs bool            `json:"export_multiple_pdfs,omitempty" yaml:"export_multiple_pdfs,omitempty" bson:"export_multiple_pdfs,omitempty"`
	IncludeDocProperties bool          `json:"include_doc_properties,omitempty" yaml:"include_doc_properties,omitempty" bson:"include_doc_properties,omitempty"`
	OpenAfterPublish   bool            `json:"open_after_publish,omitempty" yaml:"open_after_publish,omitempty" bson:"open_after_publish,omitempty"`
	Logger             *zerolog.Logger `json:"-" yaml:"-" bson:"-"`
}

// ExcelToPDFWin is a stub for non-Windows platforms
type ExcelToPDFWin struct{}

// NewExcelToPDFWin returns an error on non-Windows platforms
func NewExcelToPDFWin(config ExcelToPDFWinConfig) (*ExcelToPDFWin, *util.Result) {
	return nil, util.MsgError("NewExcelToPDFWin", "Excel to PDF conversion using COM automation is only available on Windows. Use LibreExcel2PDF for cross-platform conversion.")
}

// Convert returns an error on non-Windows platforms
func (e *ExcelToPDFWin) Convert() *util.Result {
	return util.MsgError("Convert", "Excel to PDF conversion using COM automation is only available on Windows. Use LibreExcel2PDF for cross-platform conversion.")
}
