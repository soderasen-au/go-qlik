package qrs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/soderasen-au/go-qlik/qlik/rac"

	"github.com/Click-CI/common/util"
)

type ExecutionResultDetailCondensed struct {
	Id                *string    `json:"id,omitempty"`
	Privileges        []string   `json:"privileges,omitempty"`
	DetailsType       *int32     `json:"detailsType,omitempty"`
	Message           *string    `json:"message,omitempty"`
	DetailCreatedDate *time.Time `json:"detailCreatedDate,omitempty"`
}

type ExecutionResultCondensed struct {
	Id                 *string                          `json:"id,omitempty"`
	Privileges         []string                         `json:"privileges,omitempty"`
	ExecutingNodeName  *string                          `json:"executingNodeName,omitempty"`
	Status             *int32                           `json:"status,omitempty"`
	StartTime          *time.Time                       `json:"startTime,omitempty"`
	StopTime           *time.Time                       `json:"stopTime,omitempty"`
	Duration           *int32                           `json:"duration,omitempty"`
	FileReferenceID    *string                          `json:"fileReferenceID,omitempty"`
	ScriptLogAvailable *bool                            `json:"scriptLogAvailable,omitempty"`
	Details            []ExecutionResultDetailCondensed `json:"details,omitempty"`
	ScriptLogLocation  *string                          `json:"scriptLogLocation,omitempty"`
	ScriptLogSize      *int64                           `json:"scriptLogSize,omitempty"`
}

type ReloadTaskOperationalCondensed struct {
	Id                  *string                   `json:"id,omitempty"`
	Privileges          []string                  `json:"privileges,omitempty"`
	LastExecutionResult *ExecutionResultCondensed `json:"lastExecutionResult,omitempty"`
	NextExecution       *time.Time                `json:"nextExecution,omitempty"`
}

type ReloadTask struct {
	Id                  *string                         `json:"id,omitempty"`
	CreatedDate         *time.Time                      `json:"createdDate,omitempty"`
	ModifiedDate        *time.Time                      `json:"modifiedDate,omitempty"`
	ModifiedByUserName  *string                         `json:"modifiedByUserName,omitempty"`
	SchemaPath          *string                         `json:"schemaPath,omitempty"`
	Privileges          []string                        `json:"privileges,omitempty"`
	CustomProperties    []CustomPropertyValue           `json:"customProperties,omitempty"`
	Name                string                          `json:"name"`
	TaskType            *int32                          `json:"taskType,omitempty"`
	Enabled             *bool                           `json:"enabled,omitempty"`
	TaskSessionTimeout  int32                           `json:"taskSessionTimeout"`
	MaxRetries          int32                           `json:"maxRetries"`
	Tags                []TagCondensed                  `json:"tags,omitempty"`
	App                 AppCondensed                    `json:"app"`
	IsManuallyTriggered *bool                           `json:"isManuallyTriggered,omitempty"`
	Operational         *ReloadTaskOperationalCondensed `json:"operational,omitempty"`
}

type ReloadTaskCondensed struct {
	Id                 *string                         `json:"id,omitempty"`
	Privileges         []string                        `json:"privileges,omitempty"`
	Name               string                          `json:"name"`
	TaskType           *int32                          `json:"taskType,omitempty"`
	Enabled            *bool                           `json:"enabled,omitempty"`
	TaskSessionTimeout int32                           `json:"taskSessionTimeout"`
	MaxRetries         int32                           `json:"maxRetries"`
	Operational        *ReloadTaskOperationalCondensed `json:"operational,omitempty"`
}

func (c *Client) NewReloadTask(appid string, partial bool) (*ReloadTask, *util.Result) {
	_, _, res := c.client.Do(http.MethodPost, path.Join("/app", appid, "reload"), nil)
	if res != nil {
		return nil, res.With("PostReloadRequest")
	}

	return c.GetAppReloadTask(appid)
}

func (c *Client) GetAppReloadTask(appid string) (*ReloadTask, *util.Result) {
	filter := fmt.Sprintf("(app.id eq %s) and (isManuallyTriggered eq true)", appid)
	buf, res := c.Get("/reloadtask/full", rac.WithParam("filter", filter))
	if res != nil {
		return nil, res.With("GetReloadTaskFull")
	}

	var ret []ReloadTask
	err := json.Unmarshal(buf, &ret)
	if err != nil {
		return nil, util.Error("ParseReloadTask", err)
	}

	return &ret[0], nil
}

func (c *Client) GetReloadTask(taskId string) (*ReloadTask, *util.Result) {
	var ret ReloadTask
	res := c.GetObject("/reloadtask/"+taskId, &ret)
	return &ret, res
}

func (c *Client) WaitForReloadResult(taskId string, timeout *time.Duration) (*ReloadTask, *util.Result) {
	start := time.Now()
	duration := time.Minute
	if timeout != nil {
		duration = *timeout
	}

	var status int32
	var reloadTask *ReloadTask
	var res *util.Result
	for time.Since(start) < duration {
		reloadTask, res = c.GetReloadTask(taskId)
		if res != nil {
			return nil, res.With("GetReloadTask")
		}

		status = util.MaybeNil(reloadTask.Operational.LastExecutionResult.Status)
		if status >= 7 {
			break
		}

		time.Sleep(3 * time.Second)
	}

	if status > 7 {
		return nil, util.MsgError("ReloadTask", fmt.Sprintf("task failed with %v", status))
	}

	return reloadTask, nil
}

func (c *Client) Reload(appid string, partial bool, timeout *time.Duration) (*ReloadTask, *util.Result) {
	reloadTask, res := c.NewReloadTask(appid, partial)
	if res != nil {
		c.Logger().Error().Msg(res.Error())
		return nil, res.With("new reload task")
	}

	if reloadTask.Id == nil {
		res = util.MsgError("reload task", "no task id")
		c.Logger().Error().Msg(res.Error())
		return nil, res
	}

	reloadResult, res := c.WaitForReloadResult(*reloadTask.Id, timeout)
	if res != nil {
		c.Logger().Error().Msg(res.Error())
		return nil, res.With("wait for reload task")
	}

	if res != nil {
		res = res.With("WaitForReloadResult")
		c.Logger().Error().Msg(res.Error())
		return reloadResult, res
	}

	return reloadResult, nil
}
