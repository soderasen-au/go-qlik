package qrs

import (
	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/qlik"
	"github.com/soderasen-au/go-qlik/qlik/engine"
	"github.com/soderasen-au/go-qlik/qlik/rac"
	"net/url"
	"strings"
)

type Config struct {
	rac.Config       `yaml:",inline"`
	SharedFolderRoot *string `json:"shared_folder_root,omitempty" yaml:"shared_folder_root,omitempty" bson:"shared_folder_root,omitempty"` //QSEoK shared folder which contains: `Apps`, `StaticContent` etc.
}

func NewConfigFromEngine(cfg engine.Config) *Config {
	u, _ := url.Parse(cfg.QRSBaseURI)
	var vp *string
	parts := strings.Split(u.Path, "/")
	if len(parts) == 2 && parts[0] != "" {
		vp = util.Ptr(parts[0])
	}
	return &Config{
		Config: rac.Config{
			BaseUrl:      cfg.QRSBaseURI,
			IsCloud:      util.Ptr(false),
			APIPrefix:    util.Ptr("qrs"),
			VirtualProxy: vp,
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
