package ss

import (
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/qlik/engine"
)

type CmdTaskBase struct {
	Script *Script
	Def    *FuncCmdDef
	Name   string
	Logger *zerolog.Logger
}

func (b CmdTaskBase) LogCurrentSelection() {
	if b.Logger == nil {
		return
	}
	if b.Script == nil || b.Script.Env == nil || b.Script.Env.Doc == nil {
		b.Logger.Debug().Msg("LogCurrentSelection: no info to log")
		return
	}
	selObj, res := engine.GetCurrentSelection(b.Script.Env.Doc, "$")
	if res != nil {
		b.Logger.Debug().Msg("LogCurrentSelection: can't get current selection")
		return
	}

	for _, sel := range selObj.Selections {
		b.Logger.Debug().Msgf("Field: %s, Selected: %s. Count: %d", sel.Field, sel.Selected, sel.SelectedCount)
	}
}

func NewCmdTaskBase(s *Script, d *FuncCmdDef, n string) *CmdTaskBase {
	b := &CmdTaskBase{
		Script: s,
		Def:    d,
		Name:   n,
	}
	logger := s.Env.Logger().With().Str("id", s.ID).Str("name", s.Name).Str("group", n).Logger()
	b.Logger = &logger
	return b
}

func (b CmdTaskBase) Validate() *util.Result {
	if b.Script == nil {
		return util.MsgError("Suite", "nil ptr")
	}

	if b.Script.Env == nil {
		return util.MsgError("Suite.Env", "nil ptr")
	}

	//if b.Script.Env.EngineConn == nil {
	//	return util.MsgError("Suite.Env", "no Engine connection")
	//}
	//
	//if b.Script.Env.Doc == nil {
	//	return util.MsgError("Suite.Env.doc", "no app is opened")
	//}

	if b.Def == nil {
		return util.MsgError("Def", "nil ptr")
	}

	return nil
}
