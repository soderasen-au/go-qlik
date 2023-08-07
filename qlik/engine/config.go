package engine

import (
	"net/url"
	"path"
	"strings"

	"github.com/google/uuid"

	"github.com/soderasen-au/go-common/crypto"
	"github.com/soderasen-au/go-common/util"
)

type AuthMethod string
type ServerType string

const (
	AUTH_MODE_JWT     AuthMethod = "jwt"
	AUTH_MODE_CERT    AuthMethod = "cert"
	AUTH_MODE_DESKTOP AuthMethod = "desktop"

	ST_ON_PREM ServerType = "on_prem"
	ST_CLOUD   ServerType = "cloud"
)

type HubUri struct {
	Host     string `json:"host,omitempty" yaml:"host,omitempty" bson:"host,omitempty"`
	Port     int    `json:"port,omitempty" yaml:"port,omitempty" bson:"port,omitempty"`
	Prefix   string `json:"prefix,omitempty" yaml:"prefix,omitempty" bson:"prefix,omitempty"`
	IsSecure bool   `json:"isSecure,omitempty" yaml:"isSecure,omitempty" bson:"isSecure,omitempty"`
}

// Config contains connection info used to make wss to Engine.
type Config struct {
	EngineURI          string              `json:"engine_uri,omitempty" yaml:"engine_uri,omitempty" bson:"engine_uri,omitempty"`
	AppID              string              `json:"app_id,omitempty" yaml:"app_id,omitempty" bson:"app_id,omitempty"`
	QRSBaseURI         string              `json:"qrs_base_uri,omitempty" yaml:"qrs_base_uri,omitempty" bson:"qrs_base_uri,omitempty"` //if ServerType is ST_CLOUD, this field is base URL of the QCS tenant e.g. https://tenant.eu.qlikcloud.com
	HubURI             *HubUri             `json:"hub_uri,omitempty" yaml:"hub_uri,omitempty" bson:"hub_uri,omitempty"`
	SharedFolderRoot   *string             `json:"shared_folder_root,omitempty" yaml:"shared_folder_root,omitempty" bson:"shared_folder_root,omitempty"` //QSEoK shared folder which contains: `Apps`, `StaticContent` etc.
	UserName           string              `json:"user_id,omitempty" yaml:"user_id,omitempty" bson:"user_name,omitempty"`
	UserDirectory      string              `json:"user_ad,omitempty" yaml:"user_ad,omitempty" bson:"user_directory,omitempty"`
	AuthMode           AuthMethod          `json:"auth_mode,omitempty" yaml:"auth_mode,omitempty" bson:"auth_mode,omitempty"`
	ServerType         ServerType          `json:"server_type,omitempty" yaml:"server_type,omitempty" bson:"server_type,omitempty"`
	JWT                string              `json:"jwt,omitempty" yaml:"jwt,omitempty" bson:"jwt,omitempty"`
	Certs              crypto.Certificates `json:"certs,omitempty" yaml:"certs,omitempty" bson:"certs,omitempty"`
	RandomProxySession bool                `json:"random_proxy_session" yaml:"random_proxy_session" bson:"random_proxy_session"`
}

func (cfg *Config) QCSEngineURIAppendAppID(appid string) *util.Result {
	cfg.AppID = appid

	uri, err := url.Parse(cfg.EngineURI)
	if err != nil {
		return util.Error("parse engine uri", err)
	}
	if strings.HasSuffix(uri.Path, appid) {
		return nil
	}

	if cfg.AppID == "" && cfg.IsCloud() {
		return util.MsgError("parse engine uri", "no appid for cloud engine")
	}

	_, file := path.Split(uri.Path)
	if file != "app" {
		uri.Path = path.Join(uri.Path, "app")
	}
	if cfg.IsDesktop() {
		appid = url.PathEscape(appid)
	}
	uri.Path = path.Join(uri.Path, appid)

	if cfg.RandomProxySession {
		uri.Path = path.Join(uri.Path, "identity", uuid.NewString())
	}

	cfg.EngineURI = uri.String()

	return nil
}

func (cfg Config) IsCloud() bool {
	return cfg.ServerType == ST_CLOUD
}

func (cfg Config) IsOnPrem() bool {
	return cfg.ServerType == ST_ON_PREM
}

func (cfg Config) IsDesktop() bool {
	return cfg.ServerType == ST_ON_PREM && cfg.AuthMode == AUTH_MODE_DESKTOP
}
