package qrs

import (
	"encoding/json"
	"time"

	"github.com/soderasen-au/go-common/util"
)

type Stream struct {
	ID                 string                `json:"id"`
	Name               string                `json:"name"`
	Tags               []Tag                 `json:"tags"`
	CreatedDate        time.Time             `json:"createdDate"`
	ModifiedDate       time.Time             `json:"modifiedDate"`
	ModifiedByUserName string                `json:"modifiedByUserName"`
	SchemaPath         string                `json:"schemaPath"`
	Privileges         []string              `json:"privileges"`
	CustomProperties   []CustomPropertyValue `json:"customProperties"`
	Owner              UserCondensed         `json:"owner"`
}

type StreamCondensed struct {
	Privileges []string `json:"privileges"`
	Name       string   `json:"name"`
	ID         string   `json:"id"`
}

func (c *Client) GetStreamList() ([]StreamCondensed, *util.Result) {
	resp, res := c.Get("/stream")
	if res != nil {
		return nil, res.With("get stream list")
	}

	var contents []StreamCondensed
	err := json.Unmarshal(resp, &contents)
	if err != nil {
		return nil, util.Error("parse stream list", err)
	}

	return contents, nil
}

func (c *Client) GetStreams() ([]Stream, *util.Result) {
	resp, res := c.Get("/stream/full")
	if res != nil {
		return nil, res.With("get stream list")
	}

	var contents []Stream
	err := json.Unmarshal(resp, &contents)
	if err != nil {
		return nil, util.Error("parse stream list", err)
	}

	return contents, nil
}

func (c *Client) GetStream(id string) (*Stream, *util.Result) {
	resp, res := c.Get("/stream/" + id)
	if res != nil {
		return nil, res.With("get stream")
	}

	var stream Stream
	err := json.Unmarshal(resp, &stream)
	if err != nil {
		return nil, util.Error("parse stream", err)
	}

	return &stream, nil
}
