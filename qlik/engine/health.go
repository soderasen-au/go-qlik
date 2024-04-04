package engine

import (
	"encoding/json"

	"github.com/soderasen-au/go-common/util"
)

type MemHealthInfo struct {
	Committed float64 `json:"committed" yaml:"committed"`
	Allocated float64 `json:"allocated" yaml:"allocated"`
	Free      float64 `json:"free" yaml:"free"`
}

type CPUHealthInfo struct {
	Total int `json:"total" yaml:"total"`
}

type SessionHealthInfo struct {
	Active int `json:"active" yaml:"active"`
	Total  int `json:"total" yaml:"total"`
}

type AppsHealthInfo struct {
	ActiveDocs   []string `json:"active_docs,omitempty" yaml:"active_docs,omitempty"`
	LoadedDocs   []string `json:"loaded_docs,omitempty" yaml:"loaded_docs,omitempty"`
	InMemoryDocs []string `json:"in_memory_docs,omitempty" yaml:"in_memory_docs,omitempty"`
	Calls        int64    `json:"calls" yaml:"calls"`
	Selections   int64    `json:"selections" yaml:"selections"`
}

type UsersHealthInfo struct {
	Active int `json:"active" yaml:"active"`
	Total  int `json:"total" yaml:"total"`
}

type CacheHealthInfo struct {
	Hits       int64 `json:"hits" yaml:"hits"`
	Lookups    int64 `json:"lookups" yaml:"lookups"`
	Added      int64 `json:"added" yaml:"added"`
	Replaced   int64 `json:"replaced" yaml:"replaced"`
	BytesAdded int64 `json:"bytes_added" yaml:"bytes_added"`
}

type HealthInfo struct {
	Version   string            `json:"version" yaml:"version"`
	Started   string            `json:"started" yaml:"started"`
	Mem       MemHealthInfo     `json:"mem" yaml:"mem"`
	Cpu       CPUHealthInfo     `json:"cpu" yaml:"cpu"`
	Session   SessionHealthInfo `json:"session" yaml:"session"`
	Apps      AppsHealthInfo    `json:"apps" yaml:"apps"`
	Users     UsersHealthInfo   `json:"users" yaml:"users"`
	Cache     CacheHealthInfo   `json:"cache" yaml:"cache"`
	Saturated bool              `json:"saturated" yaml:"saturated"`
}

func (c *HttpClient) GetHealthInfo() (*HealthInfo, *util.Result) {
	buf, res := c.Get("healthcheck", nil)
	if res != nil {
		return nil, res.With("Get")
	}

	var info HealthInfo
	err := json.Unmarshal(buf, &info)
	if err != nil {
		return nil, util.Error("ParseHealthInfo", err)
	}

	return &info, nil
}
