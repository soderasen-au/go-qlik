package qrs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/soderasen-au/go-qlik/qlik/rac"

	"github.com/soderasen-au/go-common/util"
)

type Subscription struct {
	TypeName     *string `json:"type_name,omitempty" yaml:"type_name,omitempty"`
	ObjectId     *string `json:"object_id,omitempty" yaml:"object_id,omitempty"`
	Filter       *string `json:"filter,omitempty" yaml:"filter,omitempty"`
	Condition    *string `json:"condition,omitempty" yaml:"condition,omitempty"`
	ChangeType   *string `json:"change_type,omitempty" yaml:"change_type,omitempty"`
	PropertyName *string `json:"property_name,omitempty" yaml:"property_name,omitempty"`
	CallbackURL  string  `json:"callback_url,omitempty" yaml:"callback_url,omitempty"`
}

type SubscriptionResponse struct {
	Value string
}

type ChangeType int

func (t ChangeType) String() string {
	switch int(t) {
	case 1:
		return "Add"
	case 2:
		return "Update"
	case 3:
		return "Delete"
	default:
		return "Undefined"
	}
}

func (t *ChangeType) MarshalJSON() ([]byte, error) {
	v := t.String()
	return json.Marshal(&v)
}

type ChangeEvent struct {
	ChangeType          ChangeType `json:"changeType,omitempty"`
	ObjectType          string     `json:"objectType,omitempty"`
	ObjectID            string     `json:"objectID,omitempty"`
	ChangedProperties   []string   `json:"changedProperties,omitempty"`
	EngineID            string     `json:"engineID,omitempty"`
	EngineType          string     `json:"engineType,omitempty"`
	OriginatorNodeID    string     `json:"originatorNodeID,omitempty"`
	OriginatorHostName  string     `json:"originatorHostName,omitempty"`
	OriginatorContextID string     `json:"originatorContextID,omitempty"`
	CreatedDate         time.Time  `json:"createdDate,omitempty"`
	ModifiedDate        time.Time  `json:"modifiedDate,omitempty"`
	SchemaPath          string     `json:"schemaPath,omitempty"`
}

type ChangeEvents []ChangeEvent

func (s Subscription) GetParams() map[string]string {
	ret := make(map[string]string)
	if s.TypeName != nil {
		ret["name"] = *s.TypeName
	}
	if s.ObjectId != nil {
		ret["id"] = *s.ObjectId
	}
	if s.Filter != nil {
		ret["filter"] = *s.Filter
	}
	if s.Condition != nil {
		ret["condition"] = *s.Condition
	}
	if s.ChangeType != nil {
		ret["changetype"] = *s.ChangeType
	}
	if s.PropertyName != nil {
		ret["propertyname"] = *s.PropertyName
	}
	return ret
}

func (c *Client) Subscribe(sub Subscription) (string, *util.Result) {
	cbUrl := strings.Trim(sub.CallbackURL, `"`)
	if cbUrl == "" {
		return "", util.MsgError("Check", "no callback url")
	}
	cbUrl = fmt.Sprintf(`"%s"`, cbUrl)
	body := strings.NewReader(cbUrl)
	_, resp, res := c.client.Do(http.MethodPost, "/notification", body, rac.WithParams(sub.GetParams()))
	if res != nil {
		return "", res.With("Do")
	}

	var subId SubscriptionResponse
	err := json.Unmarshal(resp, &subId)
	if err != nil {
		return "", util.Error("ParseId", err)
	}

	return subId.Value, nil
}
