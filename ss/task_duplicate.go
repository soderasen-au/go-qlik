package ss

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/soderasen-au/go-common/util"
)

const CMD_NAME_DUPLICATE = "duplicate"

func init() {
	taskRunnerCreators[CMD_NAME_DUPLICATE] = NewDuplicateTask
}

type DuplicateTask struct {
	*CmdTaskBase
	AppId      string
	NewAppName string
}

func (t *DuplicateTask) Run() *util.Result {
	app, res := t.Script.Env.QrsClient.Copy(t.AppId, t.NewAppName)
	if res != nil {
		t.Logger.Error().Msgf("QrsClient.Copy failed: ", res.Error())
		return res.With("QrsClient.Copy")
	}
	t.Script.Env.Stash(t.Name, app)
	t.Logger.Info().Msgf("Stash[%s]: %s", t.Name, app.ID)
	t.Script.Env.Stash(CMD_NAME_DUPLICATE, app)
	t.Logger.Info().Msgf("Stash[%s]: %s", CMD_NAME_DUPLICATE, app.ID)

	return util.OK(t.Name)
}

func NewDuplicateTask(s *Script, d *FuncCmdDef, n string) (TaskRunner, *util.Result) {
	t := &DuplicateTask{}
	t.CmdTaskBase = NewCmdTaskBase(s, d, fmt.Sprintf("%s::%s", n, CMD_NAME_DUPLICATE))

	if s.Env.QrsClient == nil {
		return nil, util.MsgError(t.Name+"::Validate", CMD_NAME_DUPLICATE+" task needs qrs client")
	}

	if res := t.CmdTaskBase.Validate(); res != nil {
		return nil, res.With(t.Name + "::Validate")
	}

	if d.Cmd != CMD_NAME_DUPLICATE {
		return nil, util.MsgError(t.Name+"::Validate", "wrong action name")
	}

	t.AppId = d.Target

	if len(d.Args) >= 1 {
		t.NewAppName = d.Args[0]
	} else {
		t.NewAppName = fmt.Sprintf("duplicate-%s", uuid.NewString()[:8])
	}

	return t, nil
}
