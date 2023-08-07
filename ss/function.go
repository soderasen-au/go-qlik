package ss

import (
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/soderasen-au/go-common/util"
)

type IFuncCmd interface {
	Exec() *util.Result
}

type FuncCmdDef struct {
	Cmd         string               `json:"cmd,omitempty" yaml:"cmd,omitempty"`
	Target      string               `json:"target,omitempty" yaml:"target,omitempty"`
	Args        []string             `json:"args,omitempty" yaml:"args,omitempty"`
	FieldValues []*enigma.FieldValue `json:"field_values,omitempty" yaml:"field_values,omitempty"`
}

type MetaInfo struct {
	ID          string `json:"id,omitempty" yaml:"id,omitempty"`
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type FuncCmdBase struct {
	MetaInfo `json:",inline" yaml:",inline"`

	Def *FuncCmdDef `json:"def,omitempty" yaml:"def,omitempty"`
	Env *ExecEnv    `json:"env,omitempty" yaml:"env,omitempty"`
}
