package engine

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/qlik/rac"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/soderasen-au/go-qlik/qlik"
)

var (
	// ConnCtx is used for Engine requests per session/connection;
	ConnCtx context.Context
)

// Conn creates wss connection to Engine
type Conn struct {
	Global *enigma.Global `json:"-"`
	Cfg    Config         `json:"config"`
}

func newCertConn(cfg Config) (con *Conn, err error) {
	headers := make(http.Header, 1)
	headers.Set("X-Qlik-User", fmt.Sprintf("UserDirectory=%s; UserId=%s", cfg.UserDirectory, cfg.UserName))
	tlsConfig, res := cfg.Certs.NewTlsConfig()
	if res != nil {
		return nil, res.With("NewTlsConfig")
	}
	global, err := enigma.Dialer{TLSClientConfig: tlsConfig}.Dial(ConnCtx, cfg.EngineURI, headers)
	if err != nil {
		return nil, errors.New(fmt.Sprintln("Could not connect", err))
	}
	conn := Conn{
		Global: global,
		Cfg:    cfg,
	}
	return &conn, nil
}

func newJwtConn(cfg Config) (con *Conn, err error) {
	if cfg.JWT == "" {
		jwtPayload := qlik.JwtClaim{
			UserID:        &cfg.UserName,
			UserDirectory: &cfg.UserDirectory,
		}
		keypair, res := cfg.Certs.NewRsaKeyPair()
		if res != nil {
			return nil, res.With("NewRsaKeyPair")
		}
		cfg.JWT, res = jwtPayload.GetJWT(keypair.Key)
		if res != nil {
			return nil, res.With("GetJWT")
		}
	}
	headers := make(http.Header, 1)
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.JWT))

	global, err := enigma.Dialer{Jar: cfg.Cookie, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}.Dial(ConnCtx, cfg.EngineURI, headers)
	if err != nil {
		return nil, errors.New(fmt.Sprintln("Could not connect", err))
	}
	conn := Conn{
		Global: global,
		Cfg:    cfg,
	}
	return &conn, nil
}

func newDesktopConn(cfg Config) (con *Conn, err error) {
	global, err := enigma.Dialer{}.Dial(ConnCtx, cfg.EngineURI, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintln("Could not connect", err))
	}
	conn := Conn{
		Global: global,
		Cfg:    cfg,
	}
	return &conn, nil
}

// NewConn creates a Engine wss according to cfg.
func NewConn(cfg Config) (con *Conn, err error) {
	res := cfg.QCSEngineURIAppendAppID(cfg.AppID)
	if res != nil {
		return nil, res
	}

	switch authMode := strings.ToLower(string(cfg.AuthMode)); authMode {
	case "cert":
		return newCertConn(cfg)
	case "jwt":
		return newJwtConn(cfg)
	case "desktop":
		return newDesktopConn(cfg)
	}

	return nil, errors.New("invalid auth mode in config")
}

func NewConnFromRAC(c *rac.RestApiClient, appId string) (*Conn, *util.Result) {
	config := Config{ServerType: ST_ON_PREM}
	if c.IsCloud() {
		config.ServerType = ST_CLOUD
	}

	baseUrl, err := url.Parse(c.Config.BaseUrl)
	if err != nil {
		return nil, util.Error("url.Parse(rc.BaseUrl)", err)
	}
	baseUrl.Scheme = "wss"
	baseUrl.Host = baseUrl.Hostname() //no port on cloud
	baseUrl.Path = "/app"
	baseUrl.Path = path.Join(baseUrl.Path, appId)

	if c.Config.Auth == nil {
		return nil, util.MsgError("CheckRAC.Config.Auth", "Config.Auth is nil")
	}

	if c.Config.Auth.Method == rac.AuthMethodJWT {
		if c.IsCloud() {
			if c.Config.Auth.CloudJwt == nil {
				return nil, util.MsgError("CheckRAC.Config.Auth", "CloudJwt is nil but it's Cloud JWT config")
			}
			if c.Config.Auth.CloudJwt.CsrfToken == "" {
				return nil, util.MsgError("CheckRAC.Config.Auth", "CloudJwt.CsrfToken is empty, you have to setup cloud jwt session first.(rac.StartCloudJWTSession)")
			}
			queries := baseUrl.Query()
			queries.Add("qlik-web-integration-id", c.Config.Auth.CloudJwt.WebIntegrationID)
			queries.Add("qlik-csrf-token", c.GetCloudCsrfToken())
			baseUrl.RawQuery = queries.Encode()

			config.Cookie = c.GetCookieJar()
		} else {
			if c.Config.Auth.User == nil {
				return nil, util.MsgError("CheckRAC.Config.Auth", "Auth.User is empty, but it's Managed JWT config")
			}
			config.UserName = c.Config.Auth.User.Id
			config.UserDirectory = c.Config.Auth.User.Directory
		}

		config.EngineURI = baseUrl.String()
		config.AppID = appId
		config.AuthMode = AUTH_MODE_JWT
		config.JWT = c.GetJWT()
	} else if c.Config.Auth.Method == rac.AuthMethodAPIKey {
		config.EngineURI = baseUrl.String()
		config.AppID = appId
		config.AuthMode = AUTH_MODE_JWT
		config.JWT = c.GetJWT()
	} else if c.Config.Auth.Method == rac.AuthMethodCert {
		if c.Config.Auth.User == nil {
			return nil, util.MsgError("CheckRAC.Config.Auth", "Auth.User is empty, but it's Cert config")
		}
		if c.Config.Auth.Certs == nil {
			return nil, util.MsgError("CheckRAC.Config.Auth", "Auth.Certs is empty, but it's Cert config")
		}
		config.UserName = c.Config.Auth.User.Id
		config.UserDirectory = c.Config.Auth.User.Directory
	} else {
		return nil, util.MsgError("CheckRAC.Config.Auth", fmt.Sprintf("AuthMethod: %v is not supported in Engine", c.Config.Auth.Method))
	}

	if c.Config.Auth.Certs != nil {
		config.Certs = *c.Config.Auth.Certs
	}

	conn, err := NewConn(config)
	if err != nil {
		return nil, util.Error("NewConn", err)
	}

	return conn, nil

}

func NewConnFromCluster(c *Cluster, appId string) (*Conn, *util.Result) {
	cfg := c.PickOneFor(appId, "")
	conn, err := NewConn(*cfg)
	if err != nil {
		return nil, util.Error("NewConn", err)
	}

	return conn, nil
}

func init() {
	ConnCtx = context.Background()
}
