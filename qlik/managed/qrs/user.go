package qrs

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/soderasen-au/go-common/util"
)

type User struct {
	ID                 string                `json:"id"`
	CreatedDate        time.Time             `json:"createdDate"`
	ModifiedDate       time.Time             `json:"modifiedDate"`
	ModifiedByUserName string                `json:"modifiedByUserName"`
	CustomProperties   []CustomPropertyValue `json:"customProperties"`
	UserID             string                `json:"userId"`
	UserDirectory      string                `json:"userDirectory"`
	Name               string                `json:"name"`
	Roles              []string              `json:"roles"`
	Attributes         []UserAttribute       `json:"attributes"`
	Inactive           bool                  `json:"inactive"`
	RemovedExternally  bool                  `json:"removedExternally"`
	Blacklisted        bool                  `json:"blacklisted"`
	DeleteProhibited   bool                  `json:"deleteProhibited"`
	Tags               []Tag                 `json:"tags"`
	Privileges         []string              `json:"privileges"`
	SchemaPath         string                `json:"schemaPath"`
}

// UserCondensed is used to describe Qlik User in condensed mode
type UserCondensed struct {
	Privileges    []string `json:"privileges"`
	UserDirectory string   `json:"userDirectory"`
	Name          string   `json:"name"`
	ID            string   `json:"id"`
	UserID        string   `json:"userId"`
}

func (c *Client) GetUser(userId string) (*User, *util.Result) {
	resp, res := c.Get("/user/" + userId)
	if res != nil {
		return nil, res.With("get user")
	}

	var contents User
	err := json.Unmarshal(resp, &contents)
	if err != nil {
		return nil, util.Error("parse user", err)
	}

	return &contents, nil
}

func (c *Client) GetUserList() ([]UserCondensed, *util.Result) {
	resp, res := c.Get("/user")
	if res != nil {
		return nil, res.With("get user list")
	}

	var contents []UserCondensed
	err := json.Unmarshal(resp, &contents)
	if err != nil {
		return nil, util.Error("parse user list", err)
	}

	return contents, nil
}

func (c *Client) GetUsers() ([]User, *util.Result) {
	resp, res := c.Get("/user/full")
	if res != nil {
		return nil, res.With("get users")
	}

	var contents []User
	err := json.Unmarshal(resp, &contents)
	if err != nil {
		return nil, util.Error("parse users", err)
	}

	return contents, nil
}

func (c *Client) GetUserByName(username string) (*UserCondensed, *util.Result) {
	users, res := c.GetUserList()
	if res != nil {
		return nil, res.With("GetUserList")
	}
	for _, user := range users {
		un := strings.ToLower(user.UserDirectory + "\\" + user.UserID)
		if un == strings.ToLower(username) {
			return &user, nil
		}
	}
	return nil, util.MsgError("GetUserByName", "Not found")
}
