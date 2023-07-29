package qcs

import (
	"encoding/json"

	"github.com/soderasen-au/go-common/util"
)

type ItemLink struct {
	Self       *Link `json:"self,omitempty"`
	Next       *Link `json:"next,omitempty"`
	Prev       *Link `json:"prev,omitempty"`
	Collection *Link `json:"collection,omitempty"`
}

type ItemsCreateItemRequestBody struct {
	Description              *string         `json:"description,omitempty"`
	Name                     string          `json:"name"`
	ResourceAttributes       json.RawMessage `json:"resourceAttributes,omitempty"`
	ResourceCreatedAt        string          `json:"resourceCreatedAt"`
	ResourceCustomAttributes json.RawMessage `json:"resourceCustomAttributes,omitempty"`
	ResourceId               *string         `json:"resourceId,omitempty"`
	ResourceLink             *string         `json:"resourceLink,omitempty"`
	ResourceType             string          `json:"resourceType"`
	ResourceSubType          *string         `json:"resourceSubType,omitempty"`
	ResourceUpdatedAt        *string         `json:"resourceUpdatedAt,omitempty"`
	SpaceId                  *string         `json:"spaceId,omitempty"`
	ThumbnailId              *string         `json:"thumbnailId,omitempty"`
}

type ItemLinks struct {
	Collections *Link `json:"collections,omitempty"`
	Open        *Link `json:"open,omitempty"`
	Self        *Link `json:"self,omitempty"`
	Thumbnail   *Link `json:"thumbnail,omitempty"`
}

type ItemTag struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ItemMeta struct {
	Actions     *[]string  `json:"actions,omitempty"`
	Collections *[]ItemTag `json:"collections,omitempty"`
	IsFavorited *bool      `json:"isFavorited,omitempty"`
	Tags        *[]ItemTag `json:"tags,omitempty"`
}

type ItemResourceSize struct {
	AppFile   *float32 `json:"appFile,omitempty"`
	AppMemory *float32 `json:"appMemory,omitempty"`
}

type ItemResultResponseBody struct {
	Actions                  []string          `json:"actions"`
	CollectionIds            []string          `json:"collectionIds"`
	CreatedAt                string            `json:"createdAt"`
	CreatorId                *string           `json:"creatorId,omitempty"`
	Description              *string           `json:"description,omitempty"`
	Id                       string            `json:"id"`
	IsFavorited              bool              `json:"isFavorited"`
	Links                    ItemLinks         `json:"links"`
	Meta                     ItemMeta          `json:"meta"`
	OwnerId                  *string           `json:"ownerId,omitempty"`
	Name                     string            `json:"name"`
	ResourceAttributes       json.RawMessage   `json:"resourceAttributes"`
	ResourceCreatedAt        string            `json:"resourceCreatedAt"`
	ResourceCustomAttributes json.RawMessage   `json:"resourceCustomAttributes"`
	ResourceId               *string           `json:"resourceId,omitempty"`
	ResourceLink             *string           `json:"resourceLink,omitempty"`
	ResourceReloadEndTime    *string           `json:"resourceReloadEndTime,omitempty"`
	ResourceReloadStatus     *string           `json:"resourceReloadStatus,omitempty"`
	ResourceSize             *ItemResourceSize `json:"resourceSize,omitempty"`
	ResourceSubType          *string           `json:"resourceSubType,omitempty"`
	ResourceType             string            `json:"resourceType"`
	ResourceUpdatedAt        string            `json:"resourceUpdatedAt"`
	SpaceId                  *string           `json:"spaceId,omitempty"`
	TenantId                 string            `json:"tenantId"`
	ThumbnailId              *string           `json:"thumbnailId,omitempty"`
	UpdatedAt                string            `json:"updatedAt"`
	UpdaterId                *string           `json:"updaterId,omitempty"`
}

type ItemCollectionLinks struct {
	Items *Link `json:"items,omitempty"`
	Self  *Link `json:"self,omitempty"`
}

type ItemCollection struct {
	CreatedAt string              `json:"createdAt"`
	CreatorId *string             `json:"creatorId,omitempty"`
	Id        string              `json:"id"`
	ItemCount int64               `json:"itemCount"`
	Links     ItemCollectionLinks `json:"links"`
	TenantId  string              `json:"tenantId"`
	Type      string              `json:"type"`
	UpdatedAt string              `json:"updatedAt"`
	UpdaterId *string             `json:"updaterId,omitempty"`

	ItemsCreateItemRequestBody
}

type ItemsListResponseBody struct {
	Data []ItemCollection `json:"data"`
}

func (c *Client) GetItems(params map[string]string) ([]ItemCollection, *util.Result) {
	req := c.NewRequest("GET", "/items", params)
	_, buf, res := c.Do(req)
	if res != nil {
		return nil, res.With("DoRequest")
	}

	var items ItemsListResponseBody
	err := json.Unmarshal(buf, &items)
	if err != nil {
		return nil, util.Error("parse response", err)
	}

	return items.Data, nil
}

func (c *Client) GetPersonalApps() ([]ItemCollection, *util.Result) {
	params := make(map[string]string)
	params["spaceId"] = "personal"
	params["resourceType"] = "app,qvapp,qlikview"
	params["limit"] = "100"

	return c.GetItems(params)
}

func (c *Client) GetAppItem(appId string) ([]ItemCollection, *util.Result) {
	params := make(map[string]string)
	params["resourceId"] = appId
	params["resourceType"] = "app"
	params["limit"] = "100"

	return c.GetItems(params)
}
