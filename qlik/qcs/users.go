package qcs

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/soderasen-au/go-common/util"
)

type RoleType string

const (
	CONSUMER      RoleType = "consumer"
	DATA_CONSUMER RoleType = "dataconsumer"
	FACILITATOR   RoleType = "facilitator"
	PRODUCER      RoleType = "producer"
	CONTRIBUTOR   RoleType = "contributor"
	PUBLISHER     RoleType = "publisher"
	OPERATOR      RoleType = "operator"
)

func (r *RoleType) Normalize() {
	rs := strings.ToLower(string(*r))
	*r = RoleType(rs)
}

func (r RoleType) IsValid() bool {
	r.Normalize()
	switch r {
	case CONSUMER, DATA_CONSUMER, FACILITATOR, PRODUCER, CONTRIBUTOR, PUBLISHER, OPERATOR:
		return true
	default:
		return false
	}
}

type Role struct {
	Id    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Level string `json:"level,omitempty"`
}

type User struct {
	Id            string     `json:"id"`
	TenantId      string     `json:"tenantId"`
	Subject       *string    `json:"subject,omitempty"`
	Status        *string    `json:"status,omitempty"`
	InviteExpiry  *int       `json:"inviteExpiry,omitempty"`
	Name          string     `json:"name"`
	Email         *string    `json:"email,omitempty"`
	CreatedAt     *time.Time `json:"createdAt,omitempty"`
	LastUpdatedAt *time.Time `json:"lastUpdatedAt,omitempty"`
	AssignedRoles []Role     `json:"assignedRoles,omitempty"`
	Links         Link       `json:"links"`
}

type Users struct {
	Data  []User `json:"data,omitempty"`
	Links *Links `json:"links,omitempty"`
}

func (c *Client) GetUser(userId string) (*User, *util.Result) {
	_, resp, res := c.Get("/users/"+userId, nil)
	if res != nil {
		return nil, res.With("get user")
	}

	var user User
	err := json.Unmarshal(resp, &user)
	if err != nil {
		return nil, util.Error("parse user", err)
	}

	return &user, nil
}

func (c *Client) GetUsers() (*Users, *util.Result) {
	_, resp, res := c.Get("/users", nil)
	if res != nil {
		return nil, res.With("get users")
	}

	var users Users
	err := json.Unmarshal(resp, &users)
	if err != nil {
		return nil, util.Error("parse users", err)
	}

	return &users, nil
}
