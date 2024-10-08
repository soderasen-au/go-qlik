package ss

import (
	"fmt"

	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/report"
)

const CMD_NAME_MOVE_FILE = "move_file"

func init() {
	taskRunnerCreators[CMD_NAME_MOVE_FILE] = NewMove_FileTask
}

type Move_FileTask struct {
	*CmdTaskBase
	SrcPath string
	TgtPath string
}

func (t *Move_FileTask) Run() *util.Result {
	if t.SrcPath == "" {
		t.Logger.Warn().Msg("no source file to move, try to use tmp report file ...")
		if rri, ok := t.Script.Env.Unstash(StashKeyReportResult); ok {
			if rr, ok := rri.(*report.ReportResult); ok {
				rf := util.MaybeNil(rr.ReportFile)
				t.Logger.Info().Msgf(" - moving file: %s", rf)
				t.SrcPath = rf
			} else {
				t.Logger.Warn().Msg(" - stashed report has not proper result")
			}
		} else {
			t.Logger.Warn().Msg(" - there's no tmp report file")
		}
	}

	t.Logger.Info().Msgf("moving file: %s => %s", t.SrcPath, t.TgtPath)
	res := util.MoveFile(t.SrcPath, t.TgtPath)
	if res != nil {
		return res.LogWith(t.Logger, "MoveFile")
	}

	return util.OK(t.Name)
}

func NewMove_FileTask(s *Script, d *FuncCmdDef, n string) (TaskRunner, *util.Result) {
	t := &Move_FileTask{}
	t.CmdTaskBase = NewCmdTaskBase(s, d, fmt.Sprintf("%s::%s", n, CMD_NAME_MOVE_FILE))

	if res := t.CmdTaskBase.Validate(); res != nil {
		return nil, res.With(t.Name + "::Validate")
	}

	if d.Cmd != CMD_NAME_MOVE_FILE {
		return nil, util.MsgError(t.Name+"::Validate", "wrong action name")
	}

	t.TgtPath = d.Target
	if t.TgtPath == "" {
		return nil, util.MsgError(t.Name+"::Validate", "no target path")
	}

	if len(d.Args) > 0 {
		t.SrcPath = d.Args[0]
	}

	return t, nil
}
