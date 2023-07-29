package qcs

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/soderasen-au/go-common/util"
)

type ActionName string

const (
	CREATE ActionName = "create"
	READ   ActionName = "read"
	UPDATE ActionName = "update"
	DELETE ActionName = "delete"

	MANAGED_SPACE string = "managed"
	SHARED_SPACE  string = "shared"
	DATA_SPACE    string = "data"
)

var (
	SPACE_TYPES map[string]string = map[string]string{
		MANAGED_SPACE: MANAGED_SPACE,
		SHARED_SPACE:  SHARED_SPACE,
		DATA_SPACE:    DATA_SPACE,
	}
)

type SpaceMeta struct {
	Actions         []ActionName `json:"actions"`
	Roles           []RoleType   `json:"roles"`
	AssignableRoles []RoleType   `json:"assignableRoles"`
}

type SpaceLinks struct {
	Self        Link `json:"self"`
	Assignments Link `json:"assignments"`
}

type Space struct {
	Id          string     `json:"id"`
	Type        *string    `json:"type,omitempty"`
	OwnerId     *string    `json:"ownerId,omitempty"`
	TenantId    string     `json:"tenantId"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	CreatedBy   *string    `json:"createdBy,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
	Meta        *SpaceMeta `json:"meta,omitempty"`
	Links       SpaceLinks `json:"links"`
}

type SpacesMeta struct {
	Count int32 `json:"count"`
}

type PagesLinks struct {
	Self Link  `json:"self"`
	Prev *Link `json:"prev,omitempty"`
	Next *Link `json:"next,omitempty"`
}

type Spaces struct {
	Data  []Space     `json:"data,omitempty"`
	Meta  *SpacesMeta `json:"meta,omitempty"`
	Links *PagesLinks `json:"links,omitempty"`
}

type SpaceGetParam struct {
	Name    string `json:"name,omitempty"`
	Action  string `json:"action,omitempty"`
	Type    string `json:"type,omitempty"`
	OwnerId string `json:"ownerId,omitempty"`
}

func (p SpaceGetParam) GetRequestParams() map[string]string {
	params := make(map[string]string)
	if p.Type != "" {
		params["type"] = p.Type
	}
	if p.Name != "" {
		params["name"] = p.Name
	}
	if p.OwnerId != "" {
		params["ownerId"] = p.OwnerId
	}
	if p.Action != "" {
		params["action"] = p.Action
	}
	return params
}

type SpaceCreate struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
}

type SpaceUpdate struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	OwnerId     string `json:"ownerId,omitempty"`
}

type SpaceAssignmentCreate struct {
	Type       string     `json:"type,omitempty"`
	AssigneeId string     `json:"assigneeId,omitempty"`
	Roles      []RoleType `json:"roles,omitempty"`
}

type SpaceAssignment struct {
	ID            string     `json:"id,omitempty"`
	Type          string     `json:"type,omitempty"`
	AssigneeId    string     `json:"assigneeId,omitempty"`
	Roles         []RoleType `json:"roles,omitempty"`
	SpaceID       string     `json:"spaceId,omitempty"`
	TenantID      string     `json:"tenantId,omitempty"`
	CreatedAt     *time.Time `json:"createdAt,omitempty"`
	CreatedBy     *string    `json:"createdBy,omitempty"`
	LastUpdatedAt *time.Time `json:"lastUpdatedAt,omitempty"`
}

func (c *Client) GetSpaceList(option SpaceGetParam) ([]Space, *util.Result) {
	params := option.GetRequestParams()
	_, resp, res := c.Get("/spaces", params)
	if res != nil {
		return nil, res.With("get spaces")
	}

	ret := make([]Space, 0)

	var spaces Spaces
	err := json.Unmarshal(resp, &spaces)
	if err != nil {
		return nil, util.Error("parse spaces", err)
	}
	ret = append(ret, spaces.Data...)

	for spaces.Links != nil && spaces.Links.Next != nil && spaces.Links.Next.Href != nil {
		nextUrl := *spaces.Links.Next.Href
		if nextUrl == "" {
			break
		}
		_, resp, res := c.GetRawUrl(nextUrl)
		if res != nil {
			return nil, res.With("get spaces")
		}
		spaces = Spaces{}
		err = json.Unmarshal(resp, &spaces)
		if err != nil {
			return nil, util.Error("parse spaces", err)
		}
		ret = append(ret, spaces.Data...)
	}
	return ret, nil
}

func (c *Client) GetSpaces(action string) (*Spaces, *util.Result) {
	params := make(map[string]string)
	params["limit"] = "100"
	params["action"] = action

	_, resp, res := c.Get("/spaces", params)
	if res != nil {
		return nil, res.With("get spaces")
	}

	var spaces Spaces
	err := json.Unmarshal(resp, &spaces)
	if err != nil {
		return nil, util.Error("parse spaces", err)
	}

	return &spaces, nil
}

func (c *Client) GetSpace(spaceId string) (*Space, *util.Result) {
	_, resp, res := c.Get("/spaces/"+spaceId, nil)
	if res != nil {
		return nil, res.With("get space")
	}

	var space Space
	err := json.Unmarshal(resp, &space)
	if err != nil {
		return nil, util.Error("parse space", err)
	}

	return &space, nil
}

func (c *Client) CreateSpace(param SpaceCreate) (*Space, *util.Result) {
	_, resp, res := c.client.Do(http.MethodPost, "/spaces", &param)
	if res != nil {
		return nil, res.With("DoRequest")
	}

	var space Space
	if err := json.Unmarshal(resp, &space); err != nil {
		return nil, util.Error("ParseResponse", err)
	}

	return &space, nil
}

func (c *Client) UpdateSpace(spaceId string, param SpaceUpdate) (*Space, *util.Result) {
	_, resp, res := c.client.Do(http.MethodPut, "/spaces/"+spaceId, &param)
	if res != nil {
		return nil, res.With("DoRequest")
	}

	var space Space
	if err := json.Unmarshal(resp, &space); err != nil {
		return nil, util.Error("ParseResponse", err)
	}

	return &space, nil
}

func (c *Client) AssignSpace(spaceId string, assignment SpaceAssignmentCreate) (*SpaceAssignment, *util.Result) {
	_, resp, res := c.client.Do(http.MethodPost, "/spaces/"+spaceId+"/assignments", &assignment)
	if res != nil {
		return nil, res.With("DoRequest")
	}

	var spaceAssignment SpaceAssignment
	if err := json.Unmarshal(resp, &spaceAssignment); err != nil {
		return nil, util.Error("ParseResponse", err)
	}

	return &spaceAssignment, nil
}
