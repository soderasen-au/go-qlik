package qrs

import (
	"encoding/json"

	"github.com/soderasen-au/go-common/util"
)

type About struct {
	SchemaPath        *string `json:"schemaPath,omitempty"`
	BuildVersion      string  `json:"buildVersion"`
	BuildDate         string  `json:"buildDate"`
	DatabaseProvider  string  `json:"databaseProvider"`
	NodeType          int32   `json:"nodeType"`
	SharedPersistence bool    `json:"sharedPersistence"`
	RequiresBootstrap bool    `json:"requiresBootstrap"`
	SingleNodeOnly    bool    `json:"singleNodeOnly"`
}

func (c *Client) About() (*About, *util.Result) {
	resp, res := c.Get("/about")
	if res != nil {
		return nil, res.With("Get")
	}

	about := About{}
	err := json.Unmarshal(resp, &about)
	if err != nil {
		return nil, util.Error("ParseAbout", err)
	}

	return &about, nil
}
