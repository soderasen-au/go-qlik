package qcs

import (
	"encoding/json"
	"time"

	"github.com/Click-CI/common/util"
)

type Group struct {
	Id            string     `json:"id"`
	Status        *string    `json:"status,omitempty"`
	TenantId      string     `json:"tenantId"`
	Name          string     `json:"name"`
	CreatedAt     *time.Time `json:"createdAt,omitempty"`
	LastUpdatedAt *time.Time `json:"lastUpdatedAt,omitempty"`
}

type Groups struct {
	Data  []Group     `json:"data,omitempty"`
	Links *PagesLinks `json:"links,omitempty"`
}

func (c *Client) GetGroups() (*Groups, *util.Result) {
	_, resp, res := c.Get("/groups", nil)
	if res != nil {
		return nil, res.With("get groups")
	}

	var groups Groups
	err := json.Unmarshal(resp, &groups)
	if err != nil {
		return nil, util.Error("parse groups", err)
	}

	return &groups, nil
}
