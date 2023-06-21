package config

import (
	"path/filepath"

	"github.com/soderasen-au/go-qlik/qlik/managed/qnp"
	"github.com/soderasen-au/go-qlik/qlik/managed/qrs"
	"github.com/soderasen-au/go-qlik/qlik/rac"

	"github.com/Click-CI/common/util"
	"github.com/soderasen-au/go-qlik/qlik/engine"
)

type QSFoldersConfig struct {
	RootFolder          *string `json:"root_folder,omitempty" yaml:"root_folder,omitempty" bson:"root_folder,omitempty"`
	AppFolder           *string `json:"app_folder,omitempty" yaml:"app_folder,omitempty" bson:"app_folder,omitempty"`
	StaticContentFolder *string `json:"static_content_folder,omitempty" yaml:"static_content_folder,omitempty" bson:"static_content_folder,omitempty"`
	ArchivedLogsFolder  *string `json:"archived_logs_folder,omitempty" yaml:"archived_logs_folder,omitempty" bson:"archived_logs_folder,omitempty"`
	LogFolder           *string `json:"log_folder,omitempty" yaml:"log_folder,omitempty" bson:"log_folder,omitempty"`
}

func (c *QSFoldersConfig) Validate() {
	util.MaybeAssignStr(&c.RootFolder, "C:/shared")
	util.MaybeAssignStr(&c.AppFolder, filepath.Join(*c.RootFolder, "Apps"))
	util.MaybeAssignStr(&c.StaticContentFolder, filepath.Join(*c.RootFolder, "StaticContent"))
	util.MaybeAssignStr(&c.ArchivedLogsFolder, filepath.Join(*c.RootFolder, "ArchivedLogs"))
	util.MaybeAssignStr(&c.LogFolder, "C:/ProgramData/Qlik/Sense/Log")
}

type QSConfig struct {
	Engine  *engine.Config   `json:"engine,omitempty" yaml:"engine,omitempty" bson:"engine,omitempty"`
	QRS     *qrs.Config      `json:"qrs,omitempty" yaml:"qrs,omitempty" bson:"qrs,omitempty"`
	QPS     *rac.Config      `json:"qps,omitempty" yaml:"qps,omitempty" bson:"qps,omitempty"`
	Hub     *HubConfig       `json:"hub,omitempty" yaml:"hub,omitempty" bson:"hub,omitempty"`
	Folders *QSFoldersConfig `json:"folders,omitempty" yaml:"folders,omitempty" bson:"folders,omitempty"`
}

type QVConfig struct {
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty" bson:"hostname,omitempty"`
}

type Config struct {
	Sense     *QSConfig   `json:"sense,omitempty" yaml:"sense,omitempty" bson:"sense,omitempty"`
	Qlikview  *QVConfig   `json:"qlikview,omitempty" yaml:"qlikview,omitempty" bson:"qlikview,omitempty"`
	NPrinting *qnp.Config `json:"nprinting,omitempty" yaml:"nprinting,omitempty" bson:"nprinting,omitempty"`
}
