package report

import (
	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/qlik/engine"
	"github.com/soderasen-au/go-qlik/qlik/managed/qrs"
	"os"
)

type SenseReportPrinter struct {
	ReportPrinterBase
	qrsClient *qrs.Client
}

func NewSenseReportPrinter(_qrsClient *qrs.Client) *SenseReportPrinter {
	p := &SenseReportPrinter{}
	p.ReportResults = make(map[string]*ReportResult)
	p.qrsClient = _qrsClient
	return p
}

func (p *SenseReportPrinter) Print(r Report) *util.Result {
	if !r.IsValid() {
		return util.MsgError("Print", "invalid report")
	}

	rResult, res := NewReportResult(r)
	if res != nil {
		return res.With("NewReportResult")
	}
	p.ReportResults[util.MaybeNil(r.ID)] = rResult
	logger := rResult.Logger.With().Str("report", *r.ID).Logger()

	if r.Doc == nil {
		return util.LogMsgError(&logger, "CheckDoc", "doc is not opened")
	}
	if r.Target != TARGET_OBJECTS {
		return util.LogMsgError(&logger, "CheckTarget", r.Target+" is not supported. Sense only supports objects")
	}
	if len(r.TargetIDs) != 1 {
		return util.LogMsgError(&logger, "CheckTarget", "Sense only supports report single object")
	}

	obj, err := r.Doc.GetObject(engine.ConnCtx, r.TargetIDs[0])
	if err != nil {
		return util.Error("GetObject", err)
	}

	downloadUrl, _, err := obj.ExportData(engine.ConnCtx, "OOXML", "", "", "A", false)
	if err != nil {
		return util.Error("ExportData", err)
	}
	logger.Info().Msgf("got download url: %s", downloadUrl)

	buf, res := p.qrsClient.GetAppContent(downloadUrl)
	if res != nil {
		return res.LogWith(&logger, "GetAppContent")
	}

	err = os.WriteFile(*rResult.ReportFile, buf, os.ModePerm)
	if err != nil {
		return util.Error("WriteFile", err)
	}
	logger.Info().Msgf("report is saved as [%s]", *rResult.ReportFile)

	return nil
}

func (p SenseReportPrinter) GetReportResult(id string) (*ReportResult, *util.Result) {
	result, ok := p.ReportResults[id]
	if !ok {
		return nil, util.MsgError("ReportFiles", "report id doesn't exists")
	}
	return result, nil
}
