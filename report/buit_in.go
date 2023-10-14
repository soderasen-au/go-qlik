package report

import "github.com/soderasen-au/go-common/util"

type BuiltInReportPrinter struct {
	ExcelPrinter *ExcelReportPrinter
	CsvPrinter   *CsvReportPrinter
}

func NewBuiltInReportPrinter() *BuiltInReportPrinter {
	p := &BuiltInReportPrinter{
		ExcelPrinter: NewExcelReportPrinter(),
		CsvPrinter:   NewCsvReportPrinter(),
	}
	return p
}

func (p BuiltInReportPrinter) GetReportResult(id string) (*ReportResult, *util.Result) {
	if result, res := p.ExcelPrinter.GetReportResult(id); res == nil {
		return result, nil
	}
	if result, res := p.CsvPrinter.GetReportResult(id); res == nil {
		return result, nil
	}
	return nil, util.MsgError("ReportFiles", "report id doesn't exists")
}

func (p *BuiltInReportPrinter) Print(r Report) *util.Result {
	if r.OutputFormat == nil {
		r.OutputFormat = util.Ptr(REPORT_FORMAT_XLSX)
	}

	if r.OutputFormat.IsExcel() {
		return p.ExcelPrinter.Print(r)
	} else if r.OutputFormat.IsCsv() {
		return p.CsvPrinter.Print(r)
	} else {
		return util.MsgError("Print", "built_in printer doesn't support output format: "+string(*r.OutputFormat))
	}
}
