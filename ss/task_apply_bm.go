package ss

import (
	"fmt"

	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/qlik/engine"
)

const CMD_NAME_APPLY_BM = "apply_bm"

func init() {
	taskRunnerCreators[CMD_NAME_APPLY_BM] = NewApplyBMTask
}

type ApplyBMTask struct {
	*CmdTaskBase
	BmTitle string
}

func (t *ApplyBMTask) Run() *util.Result {
	t.Logger.Info().Msgf("applying bookmark: %s", t.BmTitle)
	exists, res := t.Script.Env.SyncBookmark(t.BmTitle)
	if res != nil {
		return res.With("SyncBookmark")
	}

	if exists {
		bmid := t.Script.Env.bmMap[t.BmTitle]
		ok, err := t.Script.Env.Doc.ApplyBookmark(engine.ConnCtx, bmid)
		if err != nil {
			return util.Error(t.Name+"::ApplyBookmark", err)
		}

		if !ok {
			return util.MsgError(t.Name+"::ApplyBookmark", "engine returned `Fail`")
		}
	} else {
		return util.MsgError(t.Name+"::ApplyBookmark", "can't find this bookmark "+t.BmTitle)
	}

	t.LogCurrentSelection()

	return util.OK(t.Name)
}

func NewApplyBMTask(s *Script, d *FuncCmdDef, n string) (TaskRunner, *util.Result) {
	t := &ApplyBMTask{}
	t.CmdTaskBase = NewCmdTaskBase(s, d, fmt.Sprintf("%s::%s", n, CMD_NAME_APPLY_BM))

	if res := t.CmdTaskBase.Validate(); res != nil {
		return nil, res.With(t.Name + "::Validate")
	}

	if d.Cmd != CMD_NAME_APPLY_BM {
		return nil, util.MsgError(t.Name+"::Validate", "wrong action name")
	}

	t.BmTitle = d.Target

	return t, nil
}
