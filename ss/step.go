package ss

import (
	"fmt"

	"github.com/soderasen-au/go-common/util"
)

type ScriptStep struct {
	FuncCmdBase `json:",inline" yaml:",inline"`

	Function []*FuncCmdDef `json:"function,omitempty" yaml:"function,omitempty"`
}

func (tc *ScriptStep) GenerateTaskRunners(ts *Script) ([]TaskRunner, *util.Result) {
	tasks := make([]TaskRunner, 0)
	for i, a := range tc.Function {
		group := fmt.Sprintf("Step[%s]::Function[%d]", tc.Name, i)
		task, res := NewTaskRunner(ts, a, group)
		if res != nil {
			return nil, res.With(group)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
