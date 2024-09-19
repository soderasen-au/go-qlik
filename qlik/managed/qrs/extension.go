package qrs

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/qlik/rac"
)

type Extension struct {
	Id                 *string                           `json:"id,omitempty"`
	CreatedDate        *time.Time                        `json:"createdDate,omitempty"`
	ModifiedDate       *time.Time                        `json:"modifiedDate,omitempty"`
	ModifiedByUserName *string                           `json:"modifiedByUserName,omitempty"`
	SchemaPath         *string                           `json:"schemaPath,omitempty"`
	Privileges         []string                          `json:"privileges,omitempty"`
	CustomProperties   []CustomPropertyValue             `json:"customProperties,omitempty"`
	Owner              UserCondensed                     `json:"owner"`
	Name               string                            `json:"name"`
	Tags               []TagCondensed                    `json:"tags,omitempty"`
	WhiteList          FileExtensionWhiteListCondensed   `json:"whiteList"`
	References         []StaticContentReferenceCondensed `json:"references,omitempty"`
}

type FileExtensionWhiteListCondensed struct {
	Id          *string            `json:"id,omitempty"`
	Privileges  []string           `json:"privileges,omitempty"`
	LibraryType ContentLibraryType `json:"libraryType"`
}

type ContentLibraryType int32

type ImportExtensionOpt func(map[string]string) map[string]string

func ImportExtWithPwd(pwd string) ImportExtensionOpt {
	return func(params map[string]string) map[string]string {
		params["pwd"] = pwd
		return params
	}
}

func ImportExtWithoutPrivileges() ImportExtensionOpt {
	return func(params map[string]string) map[string]string {
		params["privileges"] = "false"
		return params
	}
}

func ImportExtNoReplace() ImportExtensionOpt {
	return func(params map[string]string) map[string]string {
		params["replace"] = "false"
		return params
	}
}

func (c *Client) ImportExtension(fileContents []byte, opts ...ImportExtensionOpt) ([]Extension, *util.Result) {
	params := make(map[string]string)
	params["pwd"] = ""
	params["privileges"] = "true"
	params["replace"] = "true"
	for _, opt := range opts {
		opt(params)
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", uuid.NewString())
	if err != nil {
		return nil, util.Error("Can't create multipart writer", err)
	}
	_, err = part.Write(fileContents)
	if err != nil {
		return nil, util.Error("WriteFileContents", err)
	}
	err = writer.Close()
	if err != nil {
		return nil, util.Error("Can't close multipart writer", err)
	}

	_, resp, res := c.client.Do(http.MethodPost, "extension/upload", body,
		rac.WithHeader("Content-Type", writer.FormDataContentType()),
		rac.WithParams(params))
	if res != nil {
		return nil, res.With("rac.Client.Do")
	}

	var ret []Extension
	if err = json.Unmarshal(resp, &ret); err != nil {
		return nil, util.Error("ParseExtension", err)
	}

	return ret, nil
}
