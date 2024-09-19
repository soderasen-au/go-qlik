package ss

import (
	"fmt"
	"os"

	"github.com/soderasen-au/go-common/util"
)

const CMD_NAME_DEL_FILE = "del_file"

func init() {
	taskRunnerCreators[CMD_NAME_DEL_FILE] = NewDel_FileTask
}

type Del_FileTask struct {
	*CmdTaskBase
	FilePath string
}

func (t *Del_FileTask) Run() *util.Result {
	t.Logger.Info().Msgf("deleting file: %s", t.FilePath)
	if exists, _ := util.Exists(t.FilePath); !exists {
		t.Logger.Warn().Msgf("file: %s doesn't exist, ignore task.", t.FilePath)
		return util.OK(t.Name)
	}

	err := os.Remove(t.FilePath)
	if err != nil {
		return util.Error("Remove", err)
	}

	return util.OK(t.Name)
}

func NewDel_FileTask(s *Script, d *FuncCmdDef, n string) (TaskRunner, *util.Result) {
	t := &Del_FileTask{}
	t.CmdTaskBase = NewCmdTaskBase(s, d, fmt.Sprintf("%s::%s", n, CMD_NAME_DEL_FILE))

	if res := t.CmdTaskBase.Validate(); res != nil {
		return nil, res.With(t.Name + "::Validate")
	}

	if d.Cmd != CMD_NAME_DEL_FILE {
		return nil, util.MsgError(t.Name+"::Validate", "wrong action name")
	}

	t.FilePath = d.Target

	return t, nil
}
