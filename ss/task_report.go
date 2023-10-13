package ss

import (
	"fmt"
	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/qlik/managed/qrs"
	"github.com/soderasen-au/go-qlik/report"
)

const (
	CMD_NAME_REPORT = "report"

	StashKeyReportFolder = "_report_folder"
	StashKeyReportName   = "_report_name"
	StashKeyReportResult = "_report_result"
)

func init() {
	taskRunnerCreators[CMD_NAME_REPORT] = NewReportTask
}

type ReportTask struct {
	*CmdTaskBase

	Report report.Report
}

func (t *ReportTask) Run() *util.Result {
	if t.Report.AppId == "" {
		dupAppId, ok := t.Script.Env.Unstash(StashKeyTmpDupApp)
		if !ok {
			return util.MsgError(t.Name+"::Run", "there's no duplicated app to "+CMD_NAME_REPORT)
		}
		dupApp, ok := dupAppId.(*qrs.App)
		if !ok {
			return util.MsgError(t.Name+"::Run", "duplicated app has wrong type in stash ")
		}
		t.Logger.Info().Msgf("use duplicated app: %s to %s", dupApp.ID, CMD_NAME_REPORT)
		t.Report.AppId = dupApp.ID
	}

	if t.Script.Env.Doc == nil {
		res := t.Script.Env.OpenDoc(t.Report.AppId)
		if res != nil {
			return res.LogWith(t.Logger, "Script.Env.OpenDoc")
		}
	}
	t.Report.Doc = t.Script.Env.Doc

	if res := t.Report.Validate(); res != nil {
		return res.With("Report.Validate")
	}

	var printer report.IReportPrinter
	if t.Report.Driver != nil {
		if *t.Report.Driver == report.DRIVER_SENSE {
			printer = report.NewSenseReportPrinter(t.Script.Env.QrsClient)
		} else if *t.Report.Driver == report.DRIVER_BUILT_IN {
			printer = report.NewExcelReportPrinter()
		} else {
			return util.LogMsgError(t.Logger, "load driver", "unsupported driver: "+*t.Report.Driver)
		}
	} else {
		printer = report.NewExcelReportPrinter()
	}

	if res := printer.Print(t.Report); res != nil {
		return res.With("printer.Print")
	}
	result, rr := printer.GetReportResult(*t.Report.ID)
	if rr != nil {
		return rr.With("GetReportResult")
	}

	t.Script.Env.Stash(StashKeyReportResult, result)
	return util.NewResult(t.Name, result)
}

func NewReportTask(s *Script, d *FuncCmdDef, n string) (TaskRunner, *util.Result) {
	if d.Cmd != CMD_NAME_REPORT {
		return nil, util.MsgError("::Validate", "wrong action name")
	}
	if d.Report == nil {
		return nil, util.MsgError("::Validate", "report is not defined")
	}

	t := &ReportTask{
		Report: *d.Report,
	}

	if t.Report.OutputFolder == nil {
		if rf, ok := s.Env.UnstashString(StashKeyReportFolder); ok {
			t.Report.OutputFolder = util.Ptr(rf)
		}
	}
	if t.Report.Name == nil {
		if rn, ok := s.Env.UnstashString(StashKeyReportName); ok {
			t.Report.Name = util.Ptr(rn)
		}
	}

	t.CmdTaskBase = NewCmdTaskBase(s, d, fmt.Sprintf("%s::%s", n, CMD_NAME_REPORT))
	if res := t.CmdTaskBase.Validate(); res != nil {
		return nil, res.With(t.Name + "::ValidateTaskBase")
	}

	t.Report.CurrentSelectionOrder = make(map[string]int)
	for k, v := range s.Env.csOrder {
		t.Report.CurrentSelectionOrder[k] = v
	}

	if t.Report.LogFolder == nil {
		t.Report.Logger = s.Env.Logger()
	}

	return t, nil
}
