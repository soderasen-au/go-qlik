package ss

import (
	"fmt"
	"github.com/soderasen-au/go-qlik/qlik/managed/qrs"

	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/report"
)

const CMD_NAME_REPORT = "report"

func init() {
	taskRunnerCreators[CMD_NAME_REPORT] = NewReportTask
}

type ReportTask struct {
	*CmdTaskBase

	Report report.Report
}

func (t *ReportTask) Run() *util.Result {
	var printer report.IReportPrinter
	if t.Report.Driver != nil {
		if *t.Report.Driver == report.DRIVER_SENSE {
			printer = report.NewSenseReportPrinter(t.Script.Env.QrsClient)
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

	t.CmdTaskBase = NewCmdTaskBase(s, d, fmt.Sprintf("%s::%s", n, CMD_NAME_REPORT))
	if res := t.CmdTaskBase.Validate(); res != nil {
		return nil, res.With(t.Name + "::ValidateTaskBase")
	}

	if t.Report.AppId == "" {
		dupAppId, ok := t.Script.Env.Unstash(CMD_NAME_DUPLICATE)
		if !ok {
			return nil, util.MsgError(t.Name+"::Validate", "there's no duplicated app to "+CMD_NAME_REPORT)
		}
		dupApp, ok := dupAppId.(*qrs.App)
		if !ok {
			return nil, util.MsgError(t.Name+"::Validate", "duplicated app has wrong type in stash ")
		}
		t.Logger.Info().Msgf("use duplicated app: %s to %s", dupApp.ID, CMD_NAME_REPORT)
		t.Report.AppId = dupApp.ID
	}
	t.Report.Doc = s.Env.Doc

	t.Report.CurrentSelectionOrder = make(map[string]int)
	for k, v := range s.Env.csOrder {
		t.Report.CurrentSelectionOrder[k] = v
	}

	//tempFolder := filepath.Join(*global.Settings.System.DBRootFolder, "temp", time.Now().Format("2006-01-02"))
	//if t.Report.OutputFolder == nil {
	//	if err := util.MaybeCreate(tempFolder); err != nil {
	//		return nil, util.Error(t.Name+"::MaybeCreateFolder", err)
	//	}
	//	t.Report.OutputFolder = &tempFolder
	//}
	if t.Report.LogFolder == nil {
		t.Report.Logger = s.Env.Logger()
	}

	if res := t.Report.Validate(); res != nil {
		return nil, res.With("Report.Validate")
	}

	return t, nil
}
