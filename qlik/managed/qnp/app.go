package qnp

import (
	"encoding/json"

	"github.com/soderasen-au/go-common/util"
)

type (
	App struct {
		ID          string `json:"id" yaml:"id" bson:"id"`
		Name        string `json:"name" yaml:"name" bson:"name"`
		Description string `json:"description" yaml:"description" bson:"description"`
		Enabled     bool   `json:"enabled" yaml:"enabled" bson:"enabled"`
		Created     string `json:"created" yaml:"created" bson:"created"`
		LastUpdate  string `json:"lastUpdate" yaml:"lastUpdate" bson:"lastUpdate"`
	}

	AppResponse struct {
		Data App `json:"data" yaml:"data" bson:"data"`
	}

	AppList struct {
		Items      []App `json:"items" yaml:"items" bson:"items"`
		TotalItems int   `json:"totalItems" yaml:"totalItems" bson:"totalItems"`
		Offset     int   `json:"offset" yaml:"offset" bson:"offset"`
		Limit      int   `json:"limit" yaml:"limit" bson:"limit"`
	}

	AppListResponse struct {
		Data AppList `json:"data" yaml:"data" bson:"data"`
	}
)

func (c *Client) GetApps() ([]App, *util.Result) {
	params := map[string]string{
		"limit": "9999",
	}
	resp, res := c.Get("/apps", params)
	if res != nil {
		return nil, res.With("Get")
	}
	var appListResp AppListResponse
	err := json.Unmarshal(resp, &appListResp)
	if err != nil {
		return nil, util.Error("ParseResponse", err)
	}

	return appListResp.Data.Items, nil
}

func (c *Client) GetApp(id string) (*App, *util.Result) {
	resp, res := c.Get("/apps/"+id, nil)
	if res != nil {
		return nil, res.With("Get")
	}
	var appResp AppResponse
	err := json.Unmarshal(resp, &appResp)
	if err != nil {
		return nil, util.Error("ParseResponse", err)
	}

	return &appResp.Data, nil
}
