package qcs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Click-CI/common/util"
)

type Connection struct {
	ID                  string  `json:"qID"`
	Name                string  `json:"qName"`
	ConnectStatement    string  `json:"qConnectStatement"`
	Type                string  `json:"qType"`
	LogOn               float32 `json:"qLogOn"`
	Architecture        int32   `json:"qArchitecture"`
	EngineObjectID      string  `json:"qEngineObjectID"`
	CredentialsID       *string `json:"qCredentialsID,omitempty"`
	CredentialsName     *string `json:"qCredentialsName,omitempty"`
	SeparateCredentials bool    `json:"qSeparateCredentials"`
	ReferenceKey        *string `json:"qReferenceKey,omitempty"`
	ConnectionSecret    *string `json:"qConnectionSecret,omitempty"`
	Space               *string `json:"space,omitempty"`
	User                string  `json:"user"`
	Tenant              *string `json:"tenant,omitempty"`
	Created             string  `json:"created"`
	Updated             string  `json:"updated"`
	Links               *Link   `json:"links,omitempty"`
	// Privileges []Privilege `json:"privileges"`
}

type Connections struct {
	Data   *[]Connection             `json:"data,omitempty"`
	Links  *ItemLink                 `json:"links,omitempty"`
	Errors *[]map[string]interface{} `json:"errors,omitempty"`
	//Meta   *Meta                     `json:"meta,omitempty"`
}

type DataFilesQuota struct {
	AllowedExtensions         []string `json:"allowedExtensions,omitempty"`
	AllowedInternalExtensions []string `json:"allowedInternalExtensions,omitempty"`
	MaxFileSize               *int64   `json:"maxFileSize,omitempty"`
	MaxLargeFileSize          *int64   `json:"maxLargeFileSize,omitempty"`
	MaxSize                   *int64   `json:"maxSize,omitempty"`
	Size                      *int64   `json:"size,omitempty"`
}

type DataFilesInfo struct {
	CreatedDate  string `json:"createdDate,omitempty"`
	ID           string `json:"id,omitempty"`
	ModifiedDate string `json:"modifiedDate,omitempty"`
	Name         string `json:"name,omitempty"`
	OwnerId      string `json:"ownerId,omitempty"`
	Size         *int64 `json:"size,omitempty"`
	SpaceId      string `json:"spaceId,omitempty"`
}

func (c *Client) GetConnections(spaceId string) (*Connections, *util.Result) {
	params := make(map[string]string)
	if len(spaceId) > 0 {
		params["spaceId"] = spaceId
	}

	_, resp, res := c.Get("/data-connections", params)
	if res != nil {
		return nil, res.With("get data-connections")
	}

	var conns Connections
	err := json.Unmarshal(resp, &conns)
	if err != nil {
		return nil, util.Error("parse Connections", err)
	}

	return &conns, nil
}

func (c *Client) GetDataFilesConnection(spaceId string) (*Connection, *util.Result) {
	params := make(map[string]string)
	params["type"] = "connectionname"
	if len(spaceId) > 0 {
		params["space"] = spaceId
	}

	_, resp, res := c.Get("/data-connections/DataFiles", params)
	if res != nil {
		return nil, res.With("get datafiles connection")
	}

	var conn Connection
	err := json.Unmarshal(resp, &conn)
	if err != nil {
		return nil, util.Error("parse Connection", err)
	}

	return &conn, nil
}

func (c *Client) GetDataFilesQuota() (*DataFilesQuota, *util.Result) {
	_, resp, res := c.Get("/qix-datafiles/quota", nil)
	if res != nil {
		return nil, res.With("get datafiles quota")
	}

	var quota DataFilesQuota
	err := json.Unmarshal(resp, &quota)
	if err != nil {
		return nil, util.Error("parse quota", err)
	}

	return &quota, nil
}

func (c *Client) GetConnectionDataFiles(connectionId, fileName string) ([]*DataFilesInfo, *util.Result) {
	params := make(map[string]string)
	params["connectionId"] = connectionId
	if len(fileName) > 0 {
		params["name"] = fileName
	}
	_, resp, res := c.Get("/qix-datafiles", params)
	if res != nil {
		return nil, res.With("get DataFilesInfo")
	}

	var infos []*DataFilesInfo
	err := json.Unmarshal(resp, &infos)
	if err != nil {
		return nil, util.Error("parse DataFilesInfo", err)
	}

	return infos, nil
}

func (c *Client) CheckDataFilesQuota(filePath string, quota DataFilesQuota) *util.Result {
	stat, err := os.Stat(filePath)
	if err != nil {
		return util.Error("get file stat", err)
	}

	if stat.IsDir() {
		return util.MsgError("check quota", "it is a dir")
	}

	if quota.MaxFileSize != nil && stat.Size() > *quota.MaxFileSize {
		return util.MsgError("check quota", fmt.Sprintf("file size %d > quota size %d", stat.Size(), *quota.MaxFileSize))
	}

	if len(quota.AllowedExtensions) > 0 {
		fileExt := strings.ToLower(filepath.Ext(filePath))
		if len(fileExt) < 2 {
			return util.MsgError("check quota", "invalid file extension")
		}
		if fileExt[0] == '.' {
			fileExt = fileExt[1:] //remove leaing dot
		}

		for _, ext := range quota.AllowedExtensions {
			if ext == fileExt {
				return nil
			}
		}
		return util.MsgError("check quota", fmt.Sprintf("file ext %s is now allowed", fileExt))
	}

	return nil
}

func (c *Client) ImportDataFile(filePath, spaceId string, replace bool) (*DataFilesInfo, *util.Result) {
	quota, res := c.GetDataFilesQuota()
	if quota == nil || res != nil {
		return nil, res.With("GetDataFilesQuota")
	}

	res = c.CheckDataFilesQuota(filePath, *quota)
	if res != nil {
		return nil, res.With("CheckDataFilesQuota")
	}

	conn, res := c.GetDataFilesConnection(spaceId)
	if res != nil {
		return nil, res.With("GetDataFilesConnection")
	}
	_, fileName := filepath.Split(filePath)
	infos, res := c.GetConnectionDataFiles(conn.ID, fileName)
	if res != nil {
		return nil, res.With("GetConnectionDataFiles")
	}

	doReplace := false
	if len(infos) > 0 {
		if !replace {
			return nil, util.MsgError("pre-upload check", "file has been uploaded, set `replace` to true if needed")
		}
		doReplace = true
	}

	tempFileId, res := c.UploadToTCS(filePath)
	if res != nil {
		return nil, res.With("UploadToTCS")
	}

	params := make(map[string]string)
	params["name"] = fileName
	params["tempContentFileId"] = tempFileId
	params["connectionId"] = conn.ID
	var req *http.Request
	if doReplace {
		req = c.NewRequest(http.MethodPut, "/qix-datafiles/"+infos[0].ID, params)

	} else {
		req = c.NewRequest(http.MethodPost, "/qix-datafiles", params)
	}

	_, resp, res := c.Do(req)
	if res != nil {
		return nil, res.With("create data file")
	}
	var resInfo DataFilesInfo
	err := json.Unmarshal(resp, &resInfo)
	if err != nil {
		return nil, util.Error("parse `create` response", err)
	}

	return &resInfo, nil
}
