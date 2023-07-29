package qrs

import (
	"encoding/json"
	"time"

	"github.com/soderasen-au/go-common/util"
)

type DataConnection struct {
	Id                 *string               `json:"id,omitempty"`
	CreatedDate        *time.Time            `json:"createdDate,omitempty"`
	ModifiedDate       *time.Time            `json:"modifiedDate,omitempty"`
	ModifiedByUserName *string               `json:"modifiedByUserName,omitempty"`
	SchemaPath         *string               `json:"schemaPath,omitempty"`
	Privileges         []string              `json:"privileges,omitempty"`
	CustomProperties   []CustomPropertyValue `json:"customProperties,omitempty"`
	Owner              UserCondensed         `json:"owner"`
	Name               string                `json:"name"`
	Connectionstring   string                `json:"connectionstring"`
	Type               *string               `json:"type,omitempty"`
	EngineObjectId     *string               `json:"engineObjectId,omitempty"`
	Username           *string               `json:"username,omitempty"`
	Password           *string               `json:"password,omitempty"`
	LogOn              *int32                `json:"logOn,omitempty"`
	Architecture       *int32                `json:"architecture,omitempty"`
	Tags               []Tag                 `json:"tags,omitempty"`
}

type DataConnectionCondensed struct {
	Id               *string  `json:"id,omitempty"`
	Privileges       []string `json:"privileges,omitempty"`
	Name             string   `json:"name"`
	Connectionstring string   `json:"connectionstring"`
	Type             *string  `json:"type,omitempty"`
	EngineObjectId   *string  `json:"engineObjectId,omitempty"`
	Username         *string  `json:"username,omitempty"`
	Password         *string  `json:"password,omitempty"`
	LogOn            *int32   `json:"logOn,omitempty"`
	Architecture     *int32   `json:"architecture,omitempty"`
}

func (c *Client) GetDataConnectionList() ([]DataConnectionCondensed, *util.Result) {
	resp, res := c.Get("/dataconnection", nil)
	if res != nil {
		return nil, res.With("get dataconnection list")
	}

	var contents []DataConnectionCondensed
	err := json.Unmarshal(resp, &contents)
	if err != nil {
		return nil, util.Error("parse dataconnection list", err)
	}

	return contents, nil
}

func (c *Client) GetDataConnections() ([]DataConnection, *util.Result) {
	resp, res := c.Get("/dataconnection/full", nil)
	if res != nil {
		return nil, res.With("get dataconnection list")
	}

	var contents []DataConnection
	err := json.Unmarshal(resp, &contents)
	if err != nil {
		return nil, util.Error("parse dataconnection list", err)
	}

	return contents, nil
}

func (c *Client) GetDataConnection(id string) (*DataConnection, *util.Result) {
	resp, res := c.Get("/dataconnection/"+id, nil)
	if res != nil {
		return nil, res.With("GetDataConnection")
	}

	dc := DataConnection{}
	err := json.Unmarshal(resp, &dc)
	if err != nil {
		return nil, util.Error("ParseDataConnection", err)
	}

	return &dc, nil
}
