package ss

import (
	"fmt"
	"github.com/google/uuid"

	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/qlik/engine"
)

type Script struct {
	FuncCmdBase `json:",inline" yaml:",inline"`
	AppID       *string       `json:"app_id,omitempty" yaml:"app_id,omitempty"`
	Setup       []*FuncCmdDef `json:"setup,omitempty" yaml:"setup,omitempty"`
	Steps       []*ScriptStep `json:"steps,omitempty" yaml:"steps,omitempty"`
	Cleanup     []*FuncCmdDef `json:"cleanup,omitempty" yaml:"cleanup,omitempty"`
}

func (ts *Script) CreateExecEnv(cfg *engine.Config, logger *zerolog.Logger, opts ...ExecEnvOption) *util.Result {
	env, res := NewExecEnv(cfg, util.MaybeNil(ts.AppID), logger, opts...)
	if res != nil {
		return res.With("NewExecEnv")
	}

	ts.Env = env
	return nil
}

func (ts *Script) GenerateTaskRunners() ([]TaskRunner, *util.Result) {
	tasks := make([]TaskRunner, 0)

	for i, a := range ts.Setup {
		group := fmt.Sprintf("Setup[%d]", i)
		task, res := NewTaskRunner(ts, a, group)
		if res != nil {
			return nil, res.With(group)
		}
		tasks = append(tasks, task)
	}

	for i, t := range ts.Steps {
		subtasks, res := t.GenerateTaskRunners(ts)
		if res != nil {
			return nil, res.With(fmt.Sprintf("GenerateTaskRunners for Test[%d]", i))
		}
		tasks = append(tasks, subtasks...)
	}

	for i, a := range ts.Cleanup {
		group := fmt.Sprintf("Cleanup[%d]", i)
		task, res := NewTaskRunner(ts, a, group)
		if res != nil {
			return nil, res.With(group)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s *Script) NewRequest() (*Request, *util.Result) {
	req := &Request{
		Id:     uuid.NewString(),
		Script: s,
	}
	if req.Script == nil {
		return nil, util.MsgError("CheckScriptRequest", "no script")
	}

	tasks, res := s.GenerateTaskRunners()
	if res != nil {
		return nil, res.With("GenerateTaskRunners")
	}

	req.Tasks = tasks
	return req, nil
}
