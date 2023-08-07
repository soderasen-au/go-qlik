package ss

import (
	"fmt"
	"github.com/soderasen-au/go-common/util"
	"strings"
)

type TaskRunner interface {
	Run() *util.Result
}

type TaskRunnerCreator func(s *Script, d *FuncCmdDef, n string) (TaskRunner, *util.Result)

var (
	taskRunnerCreators = map[string]TaskRunnerCreator{}
)

func NewTaskRunner(s *Script, d *FuncCmdDef, n string) (TaskRunner, *util.Result) {
	d.Cmd = strings.ToLower(d.Cmd)
	if creator, ok := taskRunnerCreators[d.Cmd]; ok {
		return creator(s, d, n)
	}
	return nil, util.MsgError("NewTaskRunner", fmt.Sprintf("no task creator for action: %s", d.Cmd))
}

func RegisterNewTask(cmdName string, creator TaskRunnerCreator) {
	taskRunnerCreators[cmdName] = creator
}
