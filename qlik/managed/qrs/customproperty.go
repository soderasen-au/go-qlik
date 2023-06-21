package qrs

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Click-CI/common/util"
)

// CustomPropertyDefinitionCondensed is used to decribe Qlik Sense CustomPropertyDefinition in Condensed format
type CustomPropertyDefinitionCondensed struct {
	Privileges   []string `json:"privileges"`
	ValueType    string   `json:"valueType"`
	Name         string   `json:"name"`
	ChoiceValues []string `json:"choiceValues"`
	ID           string   `json:"id"`
}

type CustomPropertyValue struct {
	CreatedDate        time.Time                         `json:"createdDate"`
	ModifiedByUserName string                            `json:"modifiedByUserName"`
	SchemaPath         string                            `json:"schemaPath"`
	ModifiedDate       time.Time                         `json:"modifiedDate"`
	Definition         CustomPropertyDefinitionCondensed `json:"definition"`
	ID                 string                            `json:"id"`
	Value              string                            `json:"value"`
}

type CustomPropertySyntheticValueItem struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

type CustomPropertySyntheticValue struct {
	Values  []CustomPropertySyntheticValueItem `json:"values,omitempty"`
	Removed []string                           `json:"removed,omitempty"`
	Added   []string                           `json:"added,omitempty"`
}

// CustomPropertyDefinition is used to describe Qlik Sense CustomPropertyDefinition
type CustomPropertyDefinition struct {
	Privileges         []string  `json:"privileges"`
	CreatedDate        time.Time `json:"createdDate"`
	ModifiedByUserName string    `json:"modifiedByUserName"`
	SchemaPath         string    `json:"schemaPath"`
	ValueType          string    `json:"valueType"`
	ModifiedDate       time.Time `json:"modifiedDate"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	ChoiceValues       []string  `json:"choiceValues"`
	ID                 string    `json:"id"`
	ObjectTypes        []string  `json:"objectTypes"`
}

// CreateCustomProperty is used to create Qlik Sense CustomProperty
type CreateCustomProperty struct {
	Value      string                   `json:"value"`
	Definition CustomPropertyDefinition `json:"definition"`
}

type CPValueMap map[string]*CustomPropertyValue

func NewCPMap(cps []CustomPropertyValue) CPValueMap {
	cpmap := make(CPValueMap)
	for ci, cp := range cps {
		cpmap[cp.ID] = &cps[ci]
	}
	return cpmap
}

type CPValuesMap map[string]map[string]int // map[PropertyName][PropertyValue] = PropertyValueCount;

func NewCPValuesMap(cps []CustomPropertyValue) CPValuesMap {
	valuesMap := make(CPValuesMap)
	for _, cp := range cps {
		if _, ok := valuesMap[cp.Definition.Name]; !ok {
			valuesMap[cp.Definition.Name] = make(map[string]int)
		}
		if _, ok := valuesMap[cp.Definition.Name][cp.Value]; !ok {
			valuesMap[cp.Definition.Name][cp.Value] = 0
		}
		valuesMap[cp.Definition.Name][cp.Value] = valuesMap[cp.Definition.Name][cp.Value] + 1
	}
	return valuesMap
}

type MapDiffType int

func (t MapDiffType) String() string {
	switch int(t) {
	case -1:
		return "Delete"
	case 0:
		return "Update"
	case 1:
		return "Add"
	default:
		return "Unknown"
	}
}

type MapDiffValue struct {
	Type  MapDiffType
	Key   string
	Left  *CustomPropertyValue
	Right *CustomPropertyValue
}

func (left CPValueMap) Diff(right CPValueMap) []MapDiffValue {
	diffs := make([]MapDiffValue, 0)
	for lk, lv := range left {
		if rv, ok := right[lk]; ok {
			//intersection; update?
			if lv.Value != rv.Value {
				diffs = append(diffs, MapDiffValue{
					Type:  0,
					Key:   lk,
					Left:  lv,
					Right: rv,
				})
			}
		} else {
			//left only, no right; delete?
			diffs = append(diffs, MapDiffValue{
				Type: -1,
				Key:  lk,
				Left: lv,
			})
		}
	}

	//check right-only keys; add?
	for rk, rv := range right {
		if _, ok := left[rk]; !ok {
			diffs = append(diffs, MapDiffValue{
				Type:  1,
				Key:   rk,
				Right: rv,
			})
		}
	}

	return diffs
}

func (c *Client) GetCustomPropertyList() ([]CustomPropertyDefinitionCondensed, *util.Result) {
	resp, res := c.Get("/custompropertydefinition", nil)
	if res != nil {
		return nil, res.With("GetCustomPropertyList")
	}

	cps := make([]CustomPropertyDefinitionCondensed, 0)
	err := json.Unmarshal(resp, &cps)
	if err != nil {
		return nil, util.Error("ParseCustomPropertyList", err)
	}

	return cps, nil
}

func (c *Client) GetCustomProperty(id string) (*CustomPropertyDefinition, *util.Result) {
	resp, res := c.Get(fmt.Sprintf("/custompropertydefinition/%s", id), nil)
	if res != nil {
		return nil, res.With("GetCustomDefinition")
	}

	cps := CustomPropertyDefinition{}
	err := json.Unmarshal(resp, &cps)
	if err != nil {
		return nil, util.Error("ParseCustomPropertyDefinition", err)
	}

	return &cps, nil
}

func (c *Client) AddAppCustomProperty(app *App, name, value string) *util.Result {
	if len(app.CustomProperties) > 0 {
		for _, cp := range app.CustomProperties {
			if cp.Definition.Name == name && cp.Value == value {
				return nil
			}
		}
	}

	cps, res := c.GetCustomPropertyList()
	if res != nil {
		return res.With("GetCustomPropertyList")
	}
	cpId := ""
	for _, cp := range cps {
		if cp.Name == name {
			cpId = cp.ID
			break
		}
	}
	if cpId == "" {
		return util.MsgError("GetCustomPropertyId", fmt.Sprintf("custom property %s doesn't exist", name))
	}
	cpKey := "@" + cpId

	sel, res := c.SelectApp(app.ID)
	if res != nil {
		return res.With("SelectApp")
	}
	defer func() {
		_ = c.DeleteSelection(*sel.Id)
	}()
	sync, res := c.GetAppSynthetic(*sel.Id)
	if res != nil {
		return res.With("GetAppSynthetic")
	}

	cpIdx := -1
	for si, sp := range sync.Properties {
		if util.MaybeNil(sp.Name) == cpKey {
			cpIdx = si
			break
		}
	}
	if cpIdx < 0 {
		return util.MsgError("GetCustomPropertyId", fmt.Sprintf("can't find custom property %s in App's synthetic info", cpKey))
	}

	cpValue := CustomPropertySyntheticValue{
		Removed: []string{},
		Added:   []string{value},
	}
	cpValueMessage, err := json.Marshal(&cpValue)
	if err != nil {
		return util.Error("MarshalCPValue", err)
	}
	//c.client.Logger.Info().Msgf("last mod time: %v", *sync.LatestModifiedDate)
	sync.Properties[cpIdx].Value = cpValueMessage
	*sync.Properties[cpIdx].ValueIsModified = true
	now := sync.LatestModifiedDate.Add(1 * time.Millisecond)
	sync.LatestModifiedDate = &now
	//c.Logger().Info().Msgf("last mod time: %v", *sync.LatestModifiedDate)

	res = c.UpdateAppSynthetic(*sel.Id, *sync)
	if res != nil {
		return res.With("UpdateAppSynthetic")
	}

	return nil
}
