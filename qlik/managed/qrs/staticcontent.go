package qrs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/soderasen-au/go-qlik/qlik/rac"

	"github.com/soderasen-au/go-common/util"
)

type StaticContentReferenceCondensed struct {
	ID           *string  `json:"id,omitempty"`
	Privileges   []string `json:"privileges,omitempty"`
	DataLocation *string  `json:"dataLocation,omitempty"`
	LogicalPath  string   `json:"logicalPath,omitempty"`
	ExternalPath *string  `json:"externalPath,omitempty"`
	ServeOptions int      `json:"serveOptions,omitempty"`
}

func (c *Client) GetStaticContentList() ([]StaticContentReferenceCondensed, *util.Result) {
	resp, res := c.Get("/staticcontentreference")
	if res != nil {
		return nil, res.With("get static content reference")
	}

	var contents []StaticContentReferenceCondensed
	err := json.Unmarshal(resp, &contents)
	if err != nil {
		return nil, util.Error("parse static content reference", err)
	}

	return contents, nil
}

func (c *Client) GetAppStaticContentList(appid string) ([]StaticContentReferenceCondensed, *util.Result) {
	contents, res := c.GetStaticContentList()
	if res != nil {
		return nil, res.With("GetStaticContentList")
	}

	prefix := fmt.Sprintf("/appcontent/%s/", appid)
	cList := make([]StaticContentReferenceCondensed, 0)
	for _, content := range contents {
		if strings.Index(content.LogicalPath, prefix) == 0 {
			cList = append(cList, content)
		}
	}

	return cList, nil
}

func (c *Client) GetAppContent(downloadPath string) (data []byte, res *util.Result) {
	uri, err := url.ParseRequestURI(downloadPath)
	if err != nil {
		return nil, util.Error("ParseRequestURI", err)
	}
	q := uri.Query()
	fileData, res := c.Get(rac.GetRootPath(uri.Path), rac.WithParam("serverNodeId", q["serverNodeId"][0]))
	if res != nil {
		return nil, res.With("Get")
	}

	return fileData, nil
}

func (c *Client) DownloadAppContent(downloadPath, targetFolder string) (localPath string, res *util.Result) {
	fileData, res := c.GetAppContent(downloadPath)
	if res != nil {
		return "", res.With("GetAppContent")
	}

	_, fileName := filepath.Split(downloadPath)
	targetFile := filepath.Join(targetFolder, fileName)
	file, err := os.Create(targetFile)
	if err != nil {
		return "", util.Error("create local file", err)
	}
	defer file.Close()

	_, err = file.Write(fileData)
	if err != nil {
		return "", util.Error("Write file", err)
	}

	return targetFile, nil
}

// returns list of exported static content, saved as full file path
func (c *Client) ExportAppStaticContent(appid, targetFolder string) ([]string, *util.Result) {
	if c.Cfg.SharedFolderRoot != nil {
		staticContentFolder := path.Join(*c.Cfg.SharedFolderRoot, "StaticContent", "AppContent", appid)
		ok, err := util.Exists(staticContentFolder)
		if err != nil {
			return nil, util.Error("Exists", err)
		}
		if !ok {
			return nil, util.MsgError("Exists", "static content folder doesn't exist: "+*c.Cfg.SharedFolderRoot)
		}

		files, res := util.ListFiles(staticContentFolder)
		if res != nil {
			return nil, res.With("ListFiles")
		}

		return files, nil
	}

	cList, res := c.GetAppStaticContentList(appid)
	if res != nil {
		return nil, res.With("GetAppStaticContentList")
	}

	fileList := make([]string, 0)
	for _, content := range cList {
		c.Logger().Info().Msgf("QRS: downloading app content: %s", content.LogicalPath)
		localPath, res := c.DownloadAppContent(content.LogicalPath, targetFolder)
		if res != nil {
			return fileList, res.With("DownloadAppContent: " + content.LogicalPath)
		}
		c.Logger().Info().Msgf("QRS: saved to: %s", localPath)
		fileList = append(fileList, localPath)
	}

	return fileList, nil
}
