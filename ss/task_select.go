package ss

import (
	"fmt"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/qlik/engine"
	"strings"
)

const CMD_NAME_SELECT = "select"

func init() {
	taskRunnerCreators[CMD_NAME_SELECT] = NewSelectTask
}

type SelectTask struct {
	*CmdTaskBase
	FieldName   string
	FieldValues []*enigma.FieldValue
}

func (t *SelectTask) Run() *util.Result {
	if len(t.FieldValues) < 1 {
		t.Logger.Warn().Msgf("%s::Validate: doesn't have any value to select, ignore. ", t.Name)
		return util.OK(t.Name)
	}

	field, err := t.Script.Env.Doc.GetField(engine.ConnCtx, t.FieldName, "$")
	if err != nil {
		return util.Error(t.Name+"::GetField", err)
	}

	ok, err := field.SelectValues(engine.ConnCtx, t.FieldValues, false, false)
	if err != nil {
		return util.Error(t.Name+"::SelectValues", err)
	}

	if !ok {
		return util.MsgError(t.Name+"::SelectValues", "engine returned `Fail`")
	}

	t.LogCurrentSelection()

	return util.OK(t.Name)
}

func NewSelectTask(s *Script, d *FuncCmdDef, n string) (TaskRunner, *util.Result) {
	t := &SelectTask{}
	t.CmdTaskBase = NewCmdTaskBase(s, d, fmt.Sprintf("%s::%s", n, CMD_NAME_SELECT))

	if res := t.CmdTaskBase.Validate(); res != nil {
		return nil, res.With(t.Name + "::Validate")
	}

	if d.Cmd != CMD_NAME_SELECT {
		return nil, util.MsgError(t.Name+"::Validate", "wrong action name")
	}
	if len(d.Args) < 1 && len(d.FieldValues) < 1 {
		t.Logger.Warn().Msgf("%s::Validate: doesn't have any value to select, ignore. ", t.Name)
	}

	t.FieldName = d.Target

	if len(d.Args) >= 1 {
		listObj, res := engine.GetListObject(t.Script.Env.Doc, "$", t.FieldName)
		if res != nil {
			return nil, res.With(t.Name + "::GetListObject")
		}

		containsDateTag := false
		for _, tag := range listObj.DimensionInfo.Tags {
			if tag == "$date" {
				containsDateTag = true
				break
			}
		}

		isDateField := (listObj.DimensionInfo != nil && listObj.DimensionInfo.NumFormat != nil && listObj.DimensionInfo.NumFormat.Type == "D") || containsDateTag
		t.Logger.Debug().Msgf("field [%s] is DATE ?: %v", t.FieldName, isDateField)

		t.FieldValues = make([]*enigma.FieldValue, 0)
		for _, fv := range d.Args {
			if isDateField {
				dual, err := t.Script.Env.Doc.EvaluateEx(engine.ConnCtx, fmt.Sprintf("DATE#('%s', '%s')", fv, listObj.DimensionInfo.NumFormat.Fmt))
				if err != nil {
					return nil, util.Error(t.Name+"::EvaluateEx", err)
				}
				t.Logger.Debug().Msgf("DATE: %s => %v", fv, dual)
				t.FieldValues = append(t.FieldValues, dual)
			} else {
				var err error
				fvDuel := &enigma.FieldValue{Text: fv}
				if strings.HasPrefix(fv, "=") {
					t.Logger.Debug().Msgf("[%s]::Value %s is expr, calc it first", t.FieldName, fv)
					fvDuel, err = t.Script.Env.Doc.EvaluateEx(engine.ConnCtx, fmt.Sprintf("DATE#('%s', '%s')", fv, listObj.DimensionInfo.NumFormat.Fmt))
					if err != nil {
						t.Logger.Error().Msgf("[%s]::EvaluateEx err: %s ", t.FieldName, err.Error())
						return nil, util.Error(t.Name+"::EvaluateEx", err)
					}
				}
				t.FieldValues = append(t.FieldValues, fvDuel)
			}
		}
	} else {
		t.FieldValues = d.FieldValues
	}

	s.Env.csOrder[t.FieldName] = len(s.Env.csOrder)
	t.Logger.Debug().Msgf("cs order: %s => %d", t.FieldName, s.Env.csOrder[t.FieldName])

	return t, nil
}
