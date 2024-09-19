package ss

import (
	"fmt"
	"strings"

	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/qlik/engine"
)

const CMD_NAME_SET_VAR = "set_var"

func init() {
	taskRunnerCreators[CMD_NAME_SET_VAR] = NewSetVarTask
}

type SetVarTask struct {
	*CmdTaskBase
	VarName  string
	VarValue string
}

func (t *SetVarTask) Run() *util.Result {
	err := engine.SetStringVariable(t.Script.Env.Doc, t.VarName, t.VarValue)
	if err != nil {
		return util.Error(t.Name+"::SetStringVariable", err)
	}
	t.LogCurrentSelection()

	return util.OK(t.Name)
}

func NewSetVarTask(s *Script, d *FuncCmdDef, n string) (TaskRunner, *util.Result) {
	t := &SetVarTask{}
	t.CmdTaskBase = NewCmdTaskBase(s, d, fmt.Sprintf("%s::%s", n, CMD_NAME_SET_VAR))

	if res := t.CmdTaskBase.Validate(); res != nil {
		return nil, res.With(t.Name + "::SetVariable")
	}

	if d.Cmd != CMD_NAME_SET_VAR {
		return nil, util.MsgError(t.Name+"::Validate", "wrong action name")
	}
	if len(d.Args) < 1 {
		return nil, util.MsgError(t.Name+"::Validate", "invalid arg values")
	}

	if text := strings.TrimSpace(d.Args[0]); strings.HasPrefix(text, "=") {
		t.Logger.Info().Msgf("evaluate var value: %s", text)
		dual, err := s.Env.Doc.EvaluateEx(engine.ConnCtx, text)
		if err != nil {
			t.Logger.Err(err).Msg("EvaluateEx")
			return nil, util.Error("EvaluateEx", err)
		}

		text = dual.Text
		if text == "" && dual.IsNumeric {
			text = fmt.Sprintf("%v", dual.Number)
		}
		t.Logger.Debug().Msgf("Evaluate: %s => %v, text: %s", d.Args[0], dual, text)
		d.Args[0] = text
	}

	t.VarName = d.Target
	t.VarValue = d.Args[0]

	return t, nil
}
