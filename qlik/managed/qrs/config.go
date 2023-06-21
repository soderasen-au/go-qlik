package qrs

import (
	"github.com/Click-CI/common/util"
	"github.com/soderasen-au/go-qlik/qlik"
	"github.com/soderasen-au/go-qlik/qlik/engine"
	"github.com/soderasen-au/go-qlik/qlik/rac"
)

type Config struct {
	rac.Config       `yaml:",inline"`
	SharedFolderRoot *string `json:"shared_folder_root,omitempty" yaml:"shared_folder_root,omitempty" bson:"shared_folder_root,omitempty"` //QSEoK shared folder which contains: `Apps`, `StaticContent` etc.
}

func NewConfigFromEngine(cfg engine.Config) *Config {
	return &Config{
		Config: rac.Config{
			BaseUrl: cfg.QRSBaseURI,
			IsCloud: util.Ptr(false),
			//APIPrefix: util.Ptr("qrs"),
			Auth: &rac.AuthConfig{
				Method: rac.AuthMethodCert,
				Xrf:    true,
				User: &qlik.User{
					Id:        cfg.UserName,
					Directory: cfg.UserDirectory,
				},
				Certs:    &cfg.Certs,
				Token:    nil,
				CloudJwt: nil,
			},
			ExtraTlsConfig: &rac.ExtraTLSConfig{InsecureSkipVerify: true},
			TimeoutSec:     util.Ptr(300),
		},
		SharedFolderRoot: cfg.SharedFolderRoot,
	}
}
