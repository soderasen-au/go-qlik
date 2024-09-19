package ss

import (
	"fmt"

	"github.com/soderasen-au/go-common/util"
)

const CMD_NAME_ADD_CP = "add_app_cp"

func init() {
	taskRunnerCreators[CMD_NAME_ADD_CP] = NewAddcpTask
}

type AddCpTask struct {
	*CmdTaskBase
	AppId   string
	CpName  string
	CpValue string
}

func (t *AddCpTask) Run() *util.Result {
	t.Logger.Info().Msgf("add (%s=%s) for app %s", t.CpName, t.CpValue, t.AppId)
	qrsClient := t.Script.Env.QrsClient
	if qrsClient == nil {
		return util.MsgError("QrsClient", "there's no qrs client for this task")
	}

	app, res := qrsClient.GetApp(t.AppId)
	if res != nil {
		return res.With("qrs.GetApp")
	}
	res = qrsClient.AddAppCustomProperty(app, t.CpName, t.CpValue)
	if res != nil {
		return res.With("qrs.AddAppCustomProperty")
	}

	return util.OK(t.Name)
}

func NewAddcpTask(s *Script, d *FuncCmdDef, n string) (TaskRunner, *util.Result) {
	t := &AddCpTask{}
	t.CmdTaskBase = NewCmdTaskBase(s, d, fmt.Sprintf("%s::%s", n, CMD_NAME_ADD_CP))
	if d.Cmd != CMD_NAME_ADD_CP {
		return nil, util.MsgError(t.Name+"::Validate", "wrong action name")
	}
	if res := t.CmdTaskBase.Validate(); res != nil {
		return nil, res.With(t.Name + "::Validate")
	}

	t.AppId = d.Target
	if len(d.Args) != 2 {
		return nil, util.MsgError(t.Name+"::Validate", "need cp name in Args[0] and value in Args[1]")
	}
	t.CpName = d.Args[0]
	t.CpValue = d.Args[1]

	return t, nil
}
