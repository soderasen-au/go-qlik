package qcs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Click-CI/common/util"
)

const (
	RELOAD_CREATED   string = "CREATED"
	RELOAD_QUEUED    string = "QUEUED"
	RELOAD_RELOADING string = "RELOADING"
	RELOAD_FAILED    string = "FAILED"
	RELOAD_SUCCEEDED string = "SUCCEEDED"
)

type ReloadRequest struct {
	AppId   string `json:"appId,omitempty"`
	Partial bool   `json:"partial,omitempty"`
}

type Reload struct {
	ResponseBase

	Status string  `json:"status,omitempty"`
	Log    *string `json:"log,omitempty"`
}

func (r Reload) Finished() bool {
	return r.Status == RELOAD_FAILED || r.Status == RELOAD_SUCCEEDED
}

func (r Reload) Succeeded() bool {
	return r.Status == RELOAD_SUCCEEDED
}

func (r Reload) Failed() bool {
	return r.Status == RELOAD_FAILED
}

func (c *Client) NewReloadTask(appid string, partial bool) (*Reload, *util.Result) {
	reqBody := &ReloadRequest{
		AppId:   appid,
		Partial: partial,
	}
	_, buf, res := c.client.Do(http.MethodPost, "/reloads", reqBody)
	if res != nil {
		return nil, res.With("DoRequest")
	}

	var reload Reload
	err := json.Unmarshal(buf, &reload)
	if err != nil {
		return nil, util.Error("parse response", err)
	}
	return &reload, nil
}

func (c *Client) WaitForReloadResult(reloadId string, timeout *time.Duration) (*Reload, *util.Result) {
	start := time.Now()
	duration := time.Minute
	if timeout != nil {
		duration = *timeout
	}

	for time.Since(start) < duration {
		_, buf, res := c.Get(fmt.Sprintf("/reloads/%s", reloadId), nil)
		if res != nil {
			return nil, res.With("GetReloadRecord")
		}

		var reload Reload
		err := json.Unmarshal(buf, &reload)
		if err != nil {
			return nil, util.Error("parse response", err)
		}

		if reload.Finished() {
			return &reload, nil
		}
		time.Sleep(time.Second)
	}

	return nil, util.MsgError("WaitForReloadResult", "time out")
}

func (c *Client) Reload(appid string, partial bool, timeout *time.Duration) (*Reload, *util.Result) {
	reloadTask, res := c.NewReloadTask(appid, partial)
	if res != nil {
		c.Logger().Error().Msg(res.Error())
		return nil, res.With("new reload task")
	}

	if reloadTask.ID == nil {
		res = util.MsgError("reload task", "no task id")
		c.Logger().Error().Msg(res.Error())
		return nil, res
	}

	reloadResult, res := c.WaitForReloadResult(*reloadTask.ID, timeout)
	if res != nil {
		c.Logger().Error().Msg(res.Error())
		return nil, res.With("wait for reload task")
	}

	if reloadResult.Failed() {
		res = util.MsgError("reload log", util.MaybeNil(reloadResult.Log))
		c.Logger().Error().Msg(res.Error())
		return reloadResult, res.With(fmt.Sprintf("task: %s failed", *reloadTask.ID))
	}

	return reloadResult, nil
}
