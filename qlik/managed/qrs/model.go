package qrs

import (
	"time"
)

type UserAttribute struct {
	ID                 string    `json:"id"`
	CreatedDate        time.Time `json:"createdDate"`
	ModifiedDate       time.Time `json:"modifiedDate"`
	ModifiedByUserName string    `json:"modifiedByUserName"`
	AttributeType      string    `json:"attributeType"`
	AttributeValue     string    `json:"attributeValue"`
	ExternalID         string    `json:"externalId"`
	SchemaPath         string    `json:"schemaPath"`
}

type Tag struct {
	Privileges []string `json:"privileges"`
	Name       string   `json:"name"`
	ID         string   `json:"id"`
}

type TagCondensed struct {
	Id         *string  `json:"id,omitempty"`
	Privileges []string `json:"privileges,omitempty"`
	Name       string   `json:"name"`
}
