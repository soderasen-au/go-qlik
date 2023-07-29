package qrs

import (
	"encoding/json"
	"time"

	"github.com/soderasen-au/go-common/util"
)

type AppObject struct {
	Id                 *string        `json:"id,omitempty"`
	CreatedDate        *time.Time     `json:"createdDate,omitempty"`
	ModifiedDate       *time.Time     `json:"modifiedDate,omitempty"`
	ModifiedByUserName *string        `json:"modifiedByUserName,omitempty"`
	SchemaPath         *string        `json:"schemaPath,omitempty"`
	Privileges         []string       `json:"privileges,omitempty"`
	Owner              UserCondensed  `json:"owner"`
	Name               *string        `json:"name,omitempty"`
	EngineObjectId     *string        `json:"engineObjectId,omitempty"`
	App                AppCondensed   `json:"app"`
	ContentHash        *string        `json:"contentHash,omitempty"`
	Size               *int64         `json:"size,omitempty"`
	EngineObjectType   *string        `json:"engineObjectType,omitempty"`
	Description        *string        `json:"description,omitempty"`
	Attributes         *string        `json:"attributes,omitempty"`
	ObjectType         string         `json:"objectType"`
	PublishTime        *time.Time     `json:"publishTime,omitempty"`
	Published          *bool          `json:"published,omitempty"`
	Approved           *bool          `json:"approved,omitempty"`
	Tags               []TagCondensed `json:"tags,omitempty"`
	SourceObject       *string        `json:"sourceObject,omitempty"`
	DraftObject        *string        `json:"draftObject,omitempty"`
	AppObjectBlobId    *string        `json:"appObjectBlobId,omitempty"`
}

type AppObjectCondensed struct {
	Id               *string    `json:"id,omitempty"`
	Privileges       []string   `json:"privileges,omitempty"`
	Name             *string    `json:"name,omitempty"`
	EngineObjectId   *string    `json:"engineObjectId,omitempty"`
	ContentHash      *string    `json:"contentHash,omitempty"`
	EngineObjectType *string    `json:"engineObjectType,omitempty"`
	Description      *string    `json:"description,omitempty"`
	ObjectType       string     `json:"objectType"`
	PublishTime      *time.Time `json:"publishTime,omitempty"`
	Published        *bool      `json:"published,omitempty"`
}

func (c *Client) GetAppObject(id string) (*AppObject, *util.Result) {
	resp, res := c.Get("/app/object/" + id)
	if res != nil {
		return nil, res.With("GetAppObject")
	}

	appObject := AppObject{}
	err := json.Unmarshal(resp, &appObject)
	if err != nil {
		return nil, util.Error("ParseAppObject", err)
	}

	return &appObject, nil
}
