package qrs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/soderasen-au/go-common/util"
)

type SelectionItem struct {
	Id                 *string    `json:"id,omitempty"`
	CreatedDate        *time.Time `json:"createdDate,omitempty"`
	ModifiedDate       *time.Time `json:"modifiedDate,omitempty"`
	ModifiedByUserName *string    `json:"modifiedByUserName,omitempty"`
	SchemaPath         *string    `json:"schemaPath,omitempty"`
	Type               string     `json:"type"`
	ObjectID           string     `json:"objectID"`
	ObjectName         *string    `json:"objectName,omitempty"`
}

type SelectionCondensed struct {
	Id         *string  `json:"id,omitempty"`
	Privileges []string `json:"privileges,omitempty"`
}

type Selection struct {
	Id                 *string         `json:"id,omitempty"`
	CreatedDate        *time.Time      `json:"createdDate,omitempty"`
	ModifiedDate       *time.Time      `json:"modifiedDate,omitempty"`
	ModifiedByUserName *string         `json:"modifiedByUserName,omitempty"`
	SchemaPath         *string         `json:"schemaPath,omitempty"`
	Privileges         []string        `json:"privileges,omitempty"`
	Items              []SelectionItem `json:"items,omitempty"`
}

type SyntheticPropertyCondensed struct {
	SchemaPath       *string         `json:"schemaPath,omitempty"`
	Name             *string         `json:"name,omitempty"`
	Value            json.RawMessage `json:"value,omitempty"`
	ValueIsDifferent *bool           `json:"valueIsDifferent,omitempty"`
	ValueIsModified  *bool           `json:"valueIsModified,omitempty"`
}

type SyntheticEntityCondensed struct {
	SchemaPath *string                      `json:"schemaPath,omitempty"`
	Name       *string                      `json:"name,omitempty"`
	Type       *string                      `json:"type,omitempty"`
	Access     []string                     `json:"access,omitempty"`
	Children   []SyntheticEntityCondensed   `json:"children,omitempty"`
	Properties []SyntheticPropertyCondensed `json:"properties,omitempty"`
}

type SyntheticRootEntity struct {
	SchemaPath         *string                      `json:"schemaPath,omitempty"`
	Name               *string                      `json:"name,omitempty"`
	Type               *string                      `json:"type,omitempty"`
	Access             []string                     `json:"access,omitempty"`
	Children           []SyntheticEntityCondensed   `json:"children,omitempty"`
	Properties         []SyntheticPropertyCondensed `json:"properties,omitempty"`
	LatestModifiedDate *time.Time                   `json:"latestModifiedDate,omitempty"`
}

func (c *Client) SelectApp(id string) (*Selection, *util.Result) {
	item := SelectionItem{
		Type:     "App",
		ObjectID: id,
	}
	selection := Selection{
		Items: []SelectionItem{item},
	}

	resp, res := c.Post("/Selection", &selection)
	if res != nil {
		return nil, res.With("PostSelection")
	}

	sel := Selection{}
	err := json.Unmarshal(resp, &sel)
	if err != nil {
		return nil, util.Error("ParseCustomPropertyDefinition", err)
	}

	return &sel, nil
}

func (c *Client) DeleteSelection(id string) *util.Result {
	_, _, res := c.client.Do(http.MethodDelete, "/Selection/"+id, nil)
	if res != nil {
		return res.With("DeleteSelection")
	}

	return nil
}

func (c *Client) GetAppSynthetic(selId string) (*SyntheticRootEntity, *util.Result) {
	resp, res := c.Get(fmt.Sprintf("/selection/%s/app/synthetic", selId))
	if res != nil {
		return nil, res.With("GetAppSynthetic")
	}

	app := SyntheticRootEntity{}
	err := json.Unmarshal(resp, &app)
	if err != nil {
		return nil, util.Error("ParseAppSynthetic", err)
	}

	return &app, nil
}

func (c *Client) UpdateAppSynthetic(selId string, synthetic SyntheticRootEntity) *util.Result {
	_, res := c.Do(http.MethodPut, fmt.Sprintf("/selection/%s/app/synthetic", selId), nil, &synthetic)
	if res != nil {
		return res.With("PutAppSynthetic")
	}

	return nil
}
