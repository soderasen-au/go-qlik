package qnp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/soderasen-au/go-common/util"
)

type (
	FilterRequest struct {
		ID          string `json:"id" yaml:"id" bson:"id"`
		AppID       string `json:"appId" yaml:"appId" bson:"appId"`
		Name        string `json:"name" yaml:"name" bson:"name"`
		Description string `json:"description" yaml:"description" bson:"description"`
		Enabled     bool   `json:"enabled" yaml:"enabled" bson:"enabled"`
		Created     string `json:"created" yaml:"created" bson:"created"`
		LastUpdate  string `json:"lastUpdate" yaml:"lastUpdate" bson:"lastUpdate"`
	}

	ReportPreviewFilterRequest struct {
		Page    int
		Count   int
		OrderBy string
		Filter  map[string]interface{}
		Group   map[string]interface{}
		Sorting map[string]interface{}
	}

	ReportPreview struct {
		ID           string `json:"id" yaml:"id" bson:"id"`
		Title        string `json:"title" yaml:"title" bson:"title"`
		Description  string `json:"description" yaml:"description" bson:"description"`
		Starred      bool   `json:"starred" yaml:"starred" bson:"starred"`
		Status       string `json:"status" yaml:"status" bson:"status"`
		Created      string `json:"created" yaml:"created" bson:"created"`
		LastUpdate   string `json:"lastUpdate" yaml:"lastUpdate" bson:"lastUpdate"`
		OutputFormat string `json:"outputFormat" yaml:"outputFormat" bson:"outputFormat"`
	}

	ReportPreviewListResult struct {
		Code  int             `json:"code" yaml:"code" bson:"code"`
		Total int             `json:"total" yaml:"total" bson:"total"`
		List  []ReportPreview `json:"list" yaml:"list" bson:"list"`
	}

	ReportPreviewListResponse struct {
		Code   int                      `json:"code" yaml:"code" bson:"code"`
		Result *ReportPreviewListResult `json:"result" yaml:"result" bson:"result"`
	}
)

func (c *Client) GetReportPreviewList() ([]ReportPreview, *util.Result) {
	payload := ReportPreviewFilterRequest{
		Page:    1,
		Count:   1000,
		OrderBy: "-lastUpdate",
		Filter:  map[string]interface{}{},
		Group:   map[string]interface{}{},
		Sorting: map[string]interface{}{"lastUpdate": "desc"},
	}
	resp, res := c.NSDoRaw(nil, http.MethodPost, "/npe/reportpreview/filter", nil, nil, &payload)
	if res != nil {
		return nil, res.With("NSDoRaw")
	}
	var reportPreviewList ReportPreviewListResponse
	err := json.Unmarshal(resp, &reportPreviewList)
	if err != nil {
		return nil, util.Error("ParseResponse", err)
	}
	if reportPreviewList.Result == nil {
		return nil, util.MsgError("ParseResponse", "No result")
	}

	return reportPreviewList.Result.List, nil
}

func (c *Client) GetReportPreviewThumbnail(id, pn, width, height string) ([]byte, *util.Result) {
	endpoint := fmt.Sprintf("/content-npe/reportpreview/%s/page/%s/preview", id, pn)
	params := map[string]string{
		"width":  width,
		"height": height,
	}
	data, res := c.NSDoRaw(nil, http.MethodGet, endpoint, nil, params, nil)
	if res != nil {
		return nil, res.With("NSDoRaw")
	}
	return data, nil
}
