package ss

import (
	"fmt"

	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/qlik/managed/qrs"
)

const (
	CMD_NAME_DEL_APP = "del_app"

	_DeleteStashTmpDupApp bool = true
)

func init() {
	taskRunnerCreators[CMD_NAME_DEL_APP] = NewDel_AppTask
}

type Del_AppTask struct {
	*CmdTaskBase
	AppId string
}

func (t *Del_AppTask) Run() *util.Result {
	if t.AppId == "" && _DeleteStashTmpDupApp {
		t.Logger.Warn().Msg("there's no appid to del, try to del temp duplicated app instead")
		dupAppId, ok := t.Script.Env.Unstash(StashKeyTmpDupApp)
		if !ok {
			return util.MsgError(t.Name+"::Run", "there's no duplicated app to "+CMD_NAME_DEL_APP)
		}
		dupApp, ok := dupAppId.(*qrs.App)
		if !ok {
			return util.MsgError(t.Name+"::Run", "duplicated app has wrong type in stash")
		}
		t.Logger.Info().Msgf("use duplicated app: %s to %s", dupApp.ID, CMD_NAME_DEL_APP)
		t.AppId = dupApp.ID
	}

	if t.AppId == "" {
		t.Logger.Warn().Msg("there's no app id to delete, ignore")
		return util.OK("there's no app id")
	}

	t.Logger.Info().Msgf("deleting app: %s", t.AppId)
	res := t.Script.Env.QrsClient.DeleteApp(t.AppId)
	if res != nil {
		t.Logger.Error().Msgf("QrsClient.DeleteApp failed: %s", res.Error())
		return res.With("QrsClient.DeleteApp")
	}

	t.Logger.Info().Msg("deleting stash app")
	t.Script.Env.DeleteStash(StashKeyTmpDupApp)

	return util.OK(t.Name)
}

func NewDel_AppTask(s *Script, d *FuncCmdDef, n string) (TaskRunner, *util.Result) {
	t := &Del_AppTask{}
	t.CmdTaskBase = NewCmdTaskBase(s, d, fmt.Sprintf("%s::%s", n, CMD_NAME_DEL_APP))

	if s.Env.QrsClient == nil {
		return nil, util.MsgError(t.Name+"::Validate", CMD_NAME_DEL_APP+" task needs qrs client")
	}

	if res := t.CmdTaskBase.Validate(); res != nil {
		return nil, res.With(t.Name + "::Validate")
	}

	if d.Cmd != CMD_NAME_DEL_APP {
		return nil, util.MsgError(t.Name+"::Validate", "wrong action name")
	}

	t.AppId = d.Target

	return t, nil
}
