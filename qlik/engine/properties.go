package engine

import (
	"encoding/json"
	"github.com/qlik-oss/enigma-go/v4"

	"github.com/soderasen-au/go-common/util"

	"github.com/rs/zerolog"
)

type AppProperties struct {
	CreatedDate  *string `json:"createdDate,omitempty"`
	ModifiedDate *string `json:"modifiedDate,omitempty"`
	enigma.NxAppProperties
}

func (p *AppProperties) RemoveVolatileFields() {
	p.CreatedDate = nil
	p.ModifiedDate = nil
}

type ObjectPropeties struct {
	Info       *enigma.NxInfo     `json:"qInfo"`
	Meta       *NxMeta            `json:"qMeta"`
	Properties json.RawMessage    `json:"qProperties"`
	ChildInfos []*ObjectPropeties `json:"qChildInfos"`
}

func (p *ObjectPropeties) RemoveVolatileFields() {
	p.Meta.CreatedDate = nil
	p.Meta.ModifiedDate = nil
}

func (p *ObjectPropeties) GetChild(id string) (*ObjectPropeties, bool) {
	if p.Info != nil && p.Info.Id == id {
		return p, true
	}
	for _, c := range p.ChildInfos {
		if c != nil {
			if prop, ok := c.GetChild(id); ok {
				return prop, true
			}
		}
	}
	return nil, false
}

// return title, description
func GetTitle(parentInfo *enigma.NxInfo, obj *ObjectPropeties, logger *zerolog.Logger) (*string, *string) {
	var properties map[string]json.RawMessage
	err := json.Unmarshal(obj.Properties, &properties)
	if err != nil {
		if logger != nil {
			logger.Err(err).Send()
		}
		return nil, nil
	}

	var title, desc string

	if obj.Info.Type == "variable" {
		if v, ok := properties["qName"]; ok {
			err = json.Unmarshal(v, &title)
			if err != nil {
				if logger != nil {
					logger.Err(err).Msgf("can't get variable name")
				}
				return nil, nil
			}
			return &title, nil
		}
	}

	if v, ok := properties["qMetaDef"]; ok {
		var def map[string]interface{}
		err = json.Unmarshal(v, &def)
		if err != nil {
			if logger != nil {
				logger.Err(err).Msgf("can't get meta def")
			}
			return nil, nil
		}
		if t, ok := def["title"]; ok {
			title = t.(string)
		}
		if d, ok := def["description"]; ok {
			desc = d.(string)
		}
	}

	if len(title) == 0 {
		if v, ok := properties["title"]; ok {
			err = json.Unmarshal(v, &title)
			if err != nil {
				expr := &struct {
					StringExpression struct {
						Expr string `json:"qExpr,omitempty"`
					} `json:"qStringExpression,omitempty"`
				}{}
				if err = json.Unmarshal(v, expr); err == nil {
					title = expr.StringExpression.Expr
				} else {
					if logger != nil {
						logger.Err(err).Msgf("can't get object title")
					}
				}
				return nil, nil
			}
		}
	}

	if len(desc) == 0 {
		if v, ok := properties["description"]; ok {
			err = json.Unmarshal(v, &desc)
			if err != nil {
				if logger != nil {
					logger.Err(err).Msgf("can't get object description")
				}
				return nil, nil
			}
		}
	}

	return &title, &desc
}

func GetTitleEx(obj enigma.GenericObject, objLayout ObjectLayoutEx) (*string, *string, *util.Result) {
	if objLayout.Title != "" {
		return &objLayout.Title, nil, nil
	}

	prop := ObjectPropeties{
		Info: objLayout.Info,
	}
	rawProp, err := obj.GetPropertiesRaw(ConnCtx)
	if err != nil {
		return nil, nil, util.Error("GetPropertiesRaw", err)
	}
	prop.Properties = rawProp
	title, desc := GetTitle(objLayout.Info, &prop, nil)
	return title, desc, nil
}
