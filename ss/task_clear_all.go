package ss

import (
	"fmt"

	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/qlik/engine"
)

const CMD_NAME_CLEAR_ALL = "clear_all"

func init() {
	taskRunnerCreators[CMD_NAME_CLEAR_ALL] = NewClearAllTask
}

type ClearAllTask struct {
	*CmdTaskBase
}

func (t *ClearAllTask) Run() *util.Result {
	err := t.Script.Env.Doc.ClearAll(engine.ConnCtx, false, "$")
	if err != nil {
		return util.Error(t.Name+"::ClearAll", err)
	}

	return util.OK(t.Name)
}

func NewClearAllTask(s *Script, d *FuncCmdDef, n string) (TaskRunner, *util.Result) {
	t := &ClearAllTask{}
	t.CmdTaskBase = NewCmdTaskBase(s, d, fmt.Sprintf("%s::%s", n, CMD_NAME_CLEAR_ALL))

	if res := t.CmdTaskBase.Validate(); res != nil {
		return nil, res.With(t.Name + "::Validate")
	}

	if d.Cmd != CMD_NAME_CLEAR_ALL {
		return nil, util.MsgError(t.Name+"::Validate", "wrong action name")
	}

	t.Script.Env.csOrder = make(map[string]int)

	return t, nil
}
