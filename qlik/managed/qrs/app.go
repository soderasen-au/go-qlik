package qrs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/soderasen-au/go-qlik/qlik/rac"

	"github.com/soderasen-au/go-common/util"
)

type AppExportResponse struct {
	ExportToken  string `json:"exportToken,omitempty"`
	DownloadPath string `json:"downloadPath,omitempty"`
	SchemaPath   string `json:"schemaPath,omitempty"`
	AppID        string `json:"appId,omitempty"`
	Cancelled    bool   `json:"cancelled,omitempty"`
}

type App struct {
	Owner                 UserCondensed         `json:"owner,omitempty"`
	Privileges            []string              `json:"privileges,omitempty"`
	PublishTime           time.Time             `json:"publishTime,omitempty"`
	MigrationHash         string                `json:"migrationHash,omitempty"`
	Thumbnail             string                `json:"thumbnail,omitempty"`
	SchemaPath            string                `json:"schemaPath,omitempty"`
	Description           string                `json:"description,omitempty"`
	SavedInProductVersion string                `json:"savedInProductVersion,omitempty"`
	SourceAppID           string                `json:"sourceAppId,omitempty"`
	Published             bool                  `json:"published"`
	Tags                  []Tag                 `json:"tags,omitempty"`
	LastReloadTime        time.Time             `json:"lastReloadTime,omitempty"`
	CreatedDate           time.Time             `json:"createdDate,omitempty"`
	CustomProperties      []CustomPropertyValue `json:"customProperties,omitempty"`
	ModifiedByUserName    string                `json:"modifiedByUserName,omitempty"`
	Stream                *StreamCondensed      `json:"stream,omitempty"`
	FileSize              int                   `json:"fileSize,omitempty"`
	AppID                 string                `json:"appId,omitempty"`
	ModifiedDate          time.Time             `json:"modifiedDate,omitempty"`
	Name                  string                `json:"name,omitempty"`
	DynamicColor          string                `json:"dynamicColor,omitempty"`
	ID                    string                `json:"id,omitempty"`
	AvailabilityStatus    int                   `json:"availabilityStatus,omitempty"`
	TargetAppID           string                `json:"targetAppId,omitempty"`
}

type AppCondensed struct {
	Id                    *string                `json:"id,omitempty"`
	Privileges            []string               `json:"privileges,omitempty"`
	Name                  string                 `json:"name"`
	AppId                 *string                `json:"appId,omitempty"`
	PublishTime           *time.Time             `json:"publishTime,omitempty"`
	Published             *bool                  `json:"published,omitempty"`
	Stream                *StreamCondensed       `json:"stream,omitempty"`
	SavedInProductVersion *string                `json:"savedInProductVersion,omitempty"`
	MigrationHash         *string                `json:"migrationHash,omitempty"`
	AvailabilityStatus    *AppAvailabilityStatus `json:"availabilityStatus,omitempty"`
}

type AppAvailabilityStatus int32

//type AppPtr struct {
//	Id                    *string               `json:"id,omitempty"`
//	CreatedDate           *time.Time            `json:"createdDate,omitempty"`
//	ModifiedDate          *time.Time            `json:"modifiedDate,omitempty"`
//	ModifiedByUserName    *string               `json:"modifiedByUserName,omitempty"`
//	SchemaPath            *string               `json:"schemaPath,omitempty"`
//	Privileges            []string              `json:"privileges,omitempty"`
//	CustomProperties      []CustomPropertyValue `json:"customProperties,omitempty"`
//	Owner                 *UserCondensed        `json:"owner,omitempty"`
//	Name                  *string               `json:"name,omitempty"`
//	AppId                 *string               `json:"appId,omitempty"`
//	SourceAppId           *string               `json:"sourceAppId,omitempty"`
//	TargetAppId           *string               `json:"targetAppId,omitempty"`
//	PublishTime           *time.Time            `json:"publishTime,omitempty"`
//	Published             *bool                 `json:"published,omitempty"`
//	Tags                  []TagCondensed        `json:"tags,omitempty"`
//	Description           *string               `json:"description,omitempty"`
//	Stream                *StreamCondensed      `json:"stream,omitempty"`
//	FileSize              *int32                `json:"fileSize,omitempty"`
//	LastReloadTime        *time.Time            `json:"lastReloadTime,omitempty"`
//	Thumbnail             *string               `json:"thumbnail,omitempty"`
//	SavedInProductVersion *string               `json:"savedInProductVersion,omitempty"`
//	MigrationHash         *string               `json:"migrationHash,omitempty"`
//	DynamicColor          *string               `json:"dynamicColor,omitempty"`
//	AvailabilityStatus    *int32                `json:"availabilityStatus,omitempty"`
//}

func (c *Client) GetTempContent(url *url.URL) ([]byte, *util.Result) {
	endpoint := url.Path
	data, res := c.Get(endpoint)
	if res != nil {
		return nil, res.With("Get")
	}
	c.Logger().Debug().Msgf("get temp content from url: %s", endpoint)

	return data, nil
}

func (c *Client) DownloadApp(id, dstPath string, skipData bool) *util.Result {
	endpoint := fmt.Sprintf("app/%s/export/%s", id, id)
	s := "true"
	if !skipData {
		s = "false"
	}
	_, resp, res := c.client.Do(http.MethodPost, endpoint, nil, rac.WithParam("skipData", s))
	if res != nil {
		return res.With("PostAppExport")
	}
	expResp := &AppExportResponse{}
	err := json.Unmarshal(resp, expResp)
	if err != nil {
		return util.Error("Parse export response", err)
	}

	reqUrl, _ := url.Parse("/../" + expResp.DownloadPath)
	fileData, res := c.GetTempContent(reqUrl)
	if res != nil {
		return res.With("GetTempContent")
	}

	file, err := os.Create(dstPath)
	if err != nil {
		return util.Error("create local file", err)
	}
	defer file.Close()

	_, err = file.Write(fileData)
	if err != nil {
		return util.Error("Write file", err)
	}

	return nil
}

func (c *Client) Import(qvfPath, fallbackName string, skipData bool) (*App, *util.Result) {
	keepData := "false"
	if !skipData {
		keepData = "true"
	}

	file, err := os.Open(qvfPath)
	if err != nil {
		return nil, util.Error("Can't open file", err)
	}
	defer file.Close()

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, util.Error("Can't read file", err)
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, util.Error("Can't stat file", err)
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fi.Name())
	if err != nil {
		return nil, util.Error("Can't create multipart writer", err)
	}
	part.Write(fileContents)
	err = writer.Close()
	if err != nil {
		return nil, util.Error("Can't close multipart writer", err)
	}

	//Send Http request
	_, resp, err := c.client.Do(http.MethodPost, "app/upload", body,
		rac.WithHeader("Content-Type", writer.FormDataContentType()),
		rac.WithParam("name", fallbackName), rac.WithParam("keepData", keepData))
	if err != nil {
		return nil, util.Error("QRS request failed", err)
	}

	//Cast Body into struct
	app := &App{}
	err = json.Unmarshal(resp, app)
	if err != nil {
		return nil, util.Error("can't parse response", err)
	}
	return app, nil
}

func (c *Client) ImportReplace(qvfPath, targetAppID string, skipData bool) (*App, *util.Result) {
	keepData := "false"
	if !skipData {
		keepData = "true"
	}

	file, err := os.Open(qvfPath)
	if err != nil {
		return nil, util.Error("Can't open file", err)
	}
	defer file.Close()

	fileContents, err := io.ReadAll(file)
	if err != nil {
		return nil, util.Error("Can't read file", err)
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, util.Error("Can't stat file", err)
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fi.Name())
	if err != nil {
		return nil, util.Error("Can't create multipart writer", err)
	}
	part.Write(fileContents)
	err = writer.Close()
	if err != nil {
		return nil, util.Error("Can't close multipart writer", err)
	}

	params := map[string]string{
		"targetappid": targetAppID,
		"keepData":    keepData,
	}

	resp, res := c.Post("app/upload/replace", body, rac.WithParams(params), rac.WithHeader("Content-Type", writer.FormDataContentType()))
	if res != nil {
		return nil, res.With("QRS request failed")
	}
	app := &App{}
	err = json.Unmarshal(resp, app)
	if err != nil {
		return nil, util.Error("can't parse response", err)
	}
	return app, nil
}

func (c *Client) GetApp(id string) (*App, *util.Result) {
	resp, res := c.Get("/app/" + id)
	if res != nil {
		return nil, res.With("GetApp")
	}

	app := App{}
	err := json.Unmarshal(resp, &app)
	if err != nil {
		return nil, util.Error("ParseApp", err)
	}

	return &app, nil
}

func (c *Client) GetAppInfo(appid string) (*App, *util.Result) {
	app, res := c.GetApp(appid)
	if res != nil {
		return nil, res.With("GetApp")
	}

	app.Thumbnail = c.client.GetUrl(rac.GetHostRootPath(app.Thumbnail))

	return app, nil
}

func (c *Client) DeleteApp(id string) (*App, *util.Result) {
	resp, res := c.Do(http.MethodDelete, "/app/"+id, nil, nil)
	if res != nil {
		return nil, res.With("DeleteApp")
	}

	app := App{}
	err := json.Unmarshal(resp, &app)
	if err != nil {
		return nil, util.Error("ParseApp", err)
	}

	return &app, nil
}

//func (c *Client) UpdateApp(id string, app *AppPtr) (*AppPtr, *util.Result) {
//	now := time.Now()
//	app.ModifiedDate = &now
//	resp, err := c.Do(http.MethodPut, "/app/"+id, nil, &app)
//	if err != nil {
//		return nil, util.Error("PutApp", err)
//	}
//
//	resApp := AppPtr{}
//	err = json.Unmarshal(resp, &resApp)
//	if err != nil {
//		return nil, util.Error("ParseApp", err)
//	}
//
//	return &resApp, nil
//}

func (c *Client) GetAppList() ([]App, *util.Result) {
	resp, res := c.Get("/app/full")
	if res != nil {
		return nil, res.With("GetAppList")
	}

	apps := make([]App, 0)
	err := json.Unmarshal(resp, &apps)
	if err != nil {
		return nil, util.Error("ParseAppList", err)
	}

	return apps, nil
}

func (c *Client) GetApps() ([]AppCondensed, *util.Result) {
	resp, res := c.Get("/app")
	if res != nil {
		return nil, res.With("GetApps")
	}

	apps := make([]AppCondensed, 0)
	err := json.Unmarshal(resp, &apps)
	if err != nil {
		return nil, util.Error("ParseAppList", err)
	}

	return apps, nil
}

func (c *Client) GetAppsWithTag(tagName string) ([]App, *util.Result) {
	apps, res := c.GetAppList()
	if res != nil {
		return nil, res.With("GetAppList")
	}

	logger := c.Logger().With().Str("GetAppsWithTag", tagName).Logger()
	ret := make([]App, 0)
	for _, app := range apps {
		logger.Debug().Msgf("filtering app `%s`", app.Name)
		for _, tag := range app.Tags {
			c.Logger().Debug().Msgf("    - app tag `%s`", tag.Name)
			if tag.Name == tagName {
				c.Logger().Info().Msgf("get app `%s`", app.Name)
				ret = append(ret, app)
				break
			}
		}
	}
	return ret, nil
}

func (c *Client) PublishApp(id, streamId, publishAppName string) (*App, *util.Result) {
	params := map[string]string{
		"stream": streamId,
	}
	if publishAppName != "" {
		params["name"] = publishAppName
	}

	resp, res := c.Do(http.MethodPut, "/app/"+id+"/publish", params, nil)
	if res != nil {
		return nil, res.With("PublishApp")
	}

	app := App{}
	err := json.Unmarshal(resp, &app)
	if err != nil {
		return nil, util.Error("ParseApp", err)
	}

	return &app, nil
}

func (c *Client) PublishReplaceApp(srcAppId, targetAppId string) (*App, *util.Result) {
	params := map[string]string{
		"app": targetAppId,
	}

	resp, res := c.Do(http.MethodPut, "/app/"+srcAppId+"/replace", params, nil)
	if res != nil {
		return nil, res.With("PublishReplaceApp")
	}

	app := App{}
	err := json.Unmarshal(resp, &app)
	if err != nil {
		return nil, util.Error("ParseApp", err)
	}

	return &app, nil
}

func (c *Client) ChangeAppOwner(appId, userId string) *util.Result {
	sel, res := c.SelectApp(appId)
	if res != nil {
		return res.With("SelectApp")
	}
	defer func() {
		_ = c.DeleteSelection(*sel.Id)
	}()

	userIdBuf, err := json.Marshal(userId)
	if err != nil {
		return util.Error("MarshalUserId", err)
	}

	synthetic := SyntheticRootEntity{
		Type: util.Ptr("App"),
		Properties: []SyntheticPropertyCondensed{
			{
				Name:            util.Ptr("owner"),
				Value:           json.RawMessage(userIdBuf),
				ValueIsModified: util.Ptr(true),
			},
		},
		LatestModifiedDate: util.Ptr(time.Now().Add(1 * time.Second)),
	}
	res = c.UpdateAppSynthetic(*sel.Id, synthetic)
	if res != nil {
		return res.With("UpdateAppSynthetic")
	}

	return nil
}

func (c *Client) Copy(appId, newName string) (*App, *util.Result) {
	endpoint := fmt.Sprintf("/app/%s/copy", appId)
	buf, res := c.Post(endpoint, nil, rac.WithParam("name", newName))
	if res != nil {
		return nil, res.With("Post")
	}

	var newApp App
	err := json.Unmarshal(buf, &newApp)
	if err != nil {
		return nil, util.Error("ParseNewApp", err)
	}

	return &newApp, nil
}
