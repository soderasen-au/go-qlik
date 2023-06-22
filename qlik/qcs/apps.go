package qcs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/soderasen-au/go-qlik/qlik/rac"

	"github.com/Click-CI/common/util"
	"github.com/rs/zerolog/log"
	"github.com/soderasen-au/go-qlik/qlik/engine"
)

type NxAppAttributes struct {
	//shared attributes with on-prem qrs
	ID             string `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	Thumbnail      string `json:"thumbnail,omitempty"`
	LastReloadTime string `json:"lastReloadTime,omitempty"`
	CreatedDate    string `json:"createdDate,omitempty"`
	ModifiedDate   string `json:"modifiedDate,omitempty"`
	DynamicColor   string `json:"dynamicColor,omitempty"`
	Published      bool   `json:"published"`
	PublishTime    string `json:"publishTime,omitempty"`

	//special attributes in cloud
	Owner            string          `json:"owner,omitempty"`
	OwnerID          string          `json:"ownerId,omitempty"`
	Custom           json.RawMessage `json:"custom,omitempty"`
	HasSectionAccess bool            `json:"hasSectionAccess,omitempty"`
	Encrypted        bool            `json:"encrypted,omitempty"`
	OriginAppId      string          `json:"originAppId,omitempty"`
	ResourceType     string          `json:"_resourcetype,omitempty"`
}

type NxAppCreatePrivileges struct {
	Resource  *string `json:"resource,omitempty"`
	CanCreate *bool   `json:"canCreate,omitempty"`
}
type NxApp struct {
	Attributes NxAppAttributes          `json:"attributes,omitempty"`
	Privileges []string                 `json:"privileges,omitempty"`
	Create     *[]NxAppCreatePrivileges `json:"create,omitempty"`
}

type AppUpdateAttributes struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type PublishApp struct {
	SpaceId    string              `json:"spaceId,omitempty"`
	Attributes AppUpdateAttributes `json:"attributes,omitempty"`
	Data       string              `json:"data,omitempty"`
}

type RepublishApp struct {
	TargetId         string              `json:"targetId,omitempty"`
	Data             string              `json:"data,omitempty"` // The republished app will have data from source or target app. The default is source.  * source: Publish with source data * target: Publish with target data
	Attributes       AppUpdateAttributes `json:"attributes,omitempty"`
	CheckOriginAppId bool                `json:"checkOriginAppId,omitempty"`
}

// appId is not used at the moment
func (c *Client) Import(binPath, appId, appName, spaceId string, skipData bool) (*NxApp, *util.Result) {
	fileID, res := c.UploadToTCS(binPath)
	if res != nil {
		return nil, res.With("UploadToTCS")
	}

	params := make(map[string]string)
	params["fileId"] = fileID
	if len(appName) > 0 {
		params["name"] = appName
	}
	if len(spaceId) > 0 {
		params["spaceId"] = spaceId
	}
	//params["mode"] = "autoreplace"

	_, buf, res := c.client.Do(http.MethodPost, "/apps/import", nil, rac.WithParams(params))
	if res != nil {
		return nil, res.With("DoImportRequest")
	}

	var app NxApp
	err := json.Unmarshal(buf, &app)
	if err != nil {
		return nil, util.Error("parse import response", err)
	}

	c.Logger().Info().Msgf("%s is imported to cloud as %s", binPath, app.Attributes.ID)
	return &app, nil
}

func (c *Client) Export(id, dstFolder string, skipData bool) (string, *util.Result) {
	endpoint := fmt.Sprintf("apps/%s/export", id)
	s := "true"
	if !skipData {
		s = "false"
	}
	resp, _, res := c.client.Do(http.MethodPost, endpoint, nil, rac.WithParam("NoData", s))
	if res != nil {
		return "", res.With("PostAppExport")
	}
	downloadPath := resp.Header.Get("Location")
	if downloadPath == "" {
		return "", util.MsgError("ParseExportResponse", "No download location")
	}

	endpoint = rac.GetRootPath(downloadPath)
	resp, fileData, res := c.client.Do(http.MethodGet, endpoint, nil)
	if res != nil {
		return "", res.With("DownloadFile: " + downloadPath)
	}
	disposition := resp.Header.Get("Content-Disposition")
	dispositions := strings.Split(disposition, ";")
	filename := ""
	for _, d := range dispositions {
		d = strings.TrimSpace(d)
		if strings.HasPrefix(strings.ToLower(d), "filename=") {
			filename = strings.Trim(d[9:], "\"")
		}
	}
	if filename == "" {
		filename = "downloadApp.qvf"
	}

	_ = util.MaybeCreate(dstFolder)
	dstPath := path.Join(dstFolder, filename)
	file, err := os.Create(dstPath)
	if err != nil {
		return "", util.Error("create local file", err)
	}
	defer file.Close()

	_, err = file.Write(fileData)
	if err != nil {
		return "", util.Error("Write file", err)
	}

	return filename, nil
}

func (c *Client) MakeSheetsPublic(appid string) *util.Result {
	conn, res := c.NewEngineConn(appid)
	if res != nil {
		c.Logger().Err(res).Msg("failed to connect to cloud engine")
		return util.Error("NewEngineConn", res)
	}
	defer conn.Global.DisconnectFromServer()

	doc, err := conn.Global.OpenDoc(engine.ConnCtx, appid, "", "", "", false)
	if err != nil {
		c.Logger().Err(err).Msg("can't open app in engine")
		return util.Error("OpenDoc", err)
	}
	sessionLayout, res := engine.GetSessionObjectLayout(doc)
	if res != nil {
		c.Logger().Err(err).Msg("can't get app's session object")
		return res.With("GetSessionObjectLayout")
	}

	if sessionLayout.AppObjectList != nil && sessionLayout.AppObjectList.Items != nil {
		for _, item := range sessionLayout.AppObjectList.Items {
			c.Logger().Info().Msgf("MakeSheetsPublic: %s(%s)", item.Info.Id, item.Info.Type)
			if strings.ToLower(item.Info.Type) == "sheet" {
				c.Logger().Info().Msgf("publishing sheet %s ...", item.Info.Id)
				obj, err := doc.GetObject(engine.ConnCtx, item.Info.Id)
				if err != nil {
					c.Logger().Err(err).Msg("can't get sheet object")
					return util.Error("GetObject", err)
				}

				err = obj.Publish(engine.ConnCtx)
				if err != nil {
					c.Logger().Err(err).Msg("can't get sheet object")
					return util.Error("GetObject", err)
				}
				c.Logger().Info().Msgf(" - sheet %s is published", item.Info.Id)
			} else {
				c.Logger().Info().Msg("skip non-sheet object")
			}
		}
	} else {
		c.Logger().Warn().Msg("there's no AppObjectList")
	}

	return nil
}

func (c *Client) GetPublishedApps(spaceId string) ([]NxAppAttributes, *util.Result) {
	params := make(map[string]string)
	params["limit"] = "100"
	params["sort"] = "-updatedAt"

	_, buf, res := c.Get(fmt.Sprintf("/items/%s/publisheditems", spaceId), params)
	if res != nil {
		return nil, res.With("get published items")
	}

	var itemList ItemsListResponseBody
	err := json.Unmarshal(buf, &itemList)
	if err != nil {
		return nil, util.Error("parse spaces", err)
	}

	apps := make([]NxAppAttributes, 0)
	for i, item := range itemList.Data {
		if item.ResourceAttributes != nil {
			var app NxAppAttributes
			err = json.Unmarshal(item.ResourceAttributes, &app)
			if err != nil {
				return nil, util.Error(fmt.Sprintf("parse app[%d]", i), err)
			}
			apps = append(apps, app)
		}
	}

	return apps, nil
}

func (c *Client) Publish(appId, appName, spaceId string, alwaysNew bool) (*ItemResultResponseBody, *util.Result) {
	logger := c.Logger().With().Str("func", "Publish").Str("srcApp", appId).Str("space", spaceId).Bool("alwaysNew", alwaysNew).Logger()
	logger.Info().Msg("start")

	appItems, res := c.GetAppItem(appId)
	if res != nil {
		log.Err(res).Msg("GetAppItem failed")
		return nil, res.With("get app item")
	}
	if len(appItems) < 1 {
		return nil, util.MsgError("get app item", "no item")
	}

	publishedApps, res := c.GetPublishedApps(appItems[0].Id)
	if res != nil {
		log.Err(res).Msg("GetPublishedApps failed")
		return nil, res.With("get published apps")
	}
	logger.Info().Msgf("got %d published apps", len(publishedApps))

	published := false
	targetAppID := ""
	for _, app := range publishedApps {
		// TODO: publish to multi destination?
		if app.OriginAppId == appId {
			published = true
			targetAppID = app.ID
			break
		}
	}
	logger.Info().Msgf("published(%v),  target app id(%s)", published, targetAppID)

	var publishBody []byte
	var err error
	var method string
	if published && !alwaysNew {
		logger.Info().Msgf("re-publish ...")
		method = "PUT"
		repubPayload := RepublishApp{
			TargetId:   targetAppID,
			Data:       "target",
			Attributes: AppUpdateAttributes{Name: appName},
		}
		publishBody, err = json.Marshal(&repubPayload)
	} else {
		logger.Info().Msgf("   publish ...")
		method = "POST"
		publishPayload := PublishApp{
			SpaceId:    spaceId,
			Attributes: AppUpdateAttributes{Name: appName},
			Data:       "source",
		}
		publishBody, err = json.Marshal(&publishPayload)
	}
	if err != nil {
		log.Err(err).Msg("encode publish request body failed")
		return nil, util.Error("encode publish request body", err)
	}

	logger.Info().Msgf("post publish request")
	_, publishResp, res := c.client.Do(method, fmt.Sprintf("/apps/%s/publish", appId), publishBody)
	if res != nil {
		log.Err(res).Msg("Do")
		return nil, res.With("do publish request")
	}
	var newApp NxApp
	err = json.Unmarshal(publishResp, &newApp)
	if err != nil {
		log.Err(res).Msg("Unmarshal resp")
		return nil, util.Error("decode publish response body", err)
	}

	attr, err := json.Marshal(&newApp.Attributes)
	if err != nil {
		log.Err(res).Msg("Marshal attr")
		return nil, util.Error("encode create item request attribute", err)
	}
	itemPayload := ItemsCreateItemRequestBody{
		Name:               newApp.Attributes.Name,
		ResourceAttributes: attr,
		ResourceCreatedAt:  newApp.Attributes.CreatedDate,
		ResourceId:         &newApp.Attributes.ID,
		ResourceType:       newApp.Attributes.ResourceType,
		SpaceId:            &spaceId,
	}
	itemBody, err := json.Marshal(&itemPayload)
	if err != nil {
		log.Err(res).Msg("Marshal item")
		return nil, util.Error("encode create item request body", err)
	}
	logger.Info().Msgf("post new item")
	_, itemResp, res := c.client.Do(http.MethodPost, "/items", itemBody)
	if res != nil {
		log.Err(res).Msg("post new item")
		return nil, res.With("do create item request")
	}
	var newItem ItemResultResponseBody
	err = json.Unmarshal(itemResp, &newItem)
	if err != nil {
		log.Err(res).Msg("Unmarshal new item")
		return nil, util.Error("decode create item response body", err)
	}

	var appAttr NxAppAttributes
	err = json.Unmarshal(newItem.ResourceAttributes, &appAttr)
	if err != nil {
		log.Err(res).Msg("Unmarshal NxAppAttributes")
		return nil, util.Error("Unmarshal NxAppAttributes", err)
	}
	logger.Info().Msgf("newly published app id: %s, with original app id: %s", appAttr.ID, appAttr.OriginAppId)

	logger.Info().Msgf("get new item")
	_, _, res = c.Get("/items/"+newItem.Id, nil)
	if res != nil {
		log.Err(res).Msg("get new item")
		return nil, util.Error("can't get newly created item", err)
	}

	logger.Info().Msg("end")
	return &newItem, nil
}

func (c *Client) Delete(appId string) *util.Result {
	logger := c.Logger().With().Str("func", "Delete").Str("srcApp", appId).Logger()
	logger.Info().Msg("start")

	_, _, res := c.client.Do(http.MethodDelete, "/apps/"+appId, nil)
	if res != nil {
		log.Err(res).Msg("delete app")
		return res.With("Do request")
	}

	logger.Info().Msg("end")
	return nil
}
