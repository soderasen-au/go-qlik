package rac

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"time"

	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/crypto"
	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/qlik"
)

const (
	QlikXrfKey string = "abcdefghijklmnop"
)

type RestApiClient struct {
	Config  Config             `json:"config" yaml:"config" bson:"config"`
	Logger  *zerolog.Logger    `json:"-" yaml:"-" bson:"-"`
	baseUrl *url.URL           `json:"-" yaml:"-" bson:"-"`
	certs   *crypto.RsaKeyPair `json:"-" yaml:"-" bson:"-"`
	client  *http.Client       `json:"-" yaml:"-" bson:"-"`

	csrfToken string `json:"-" yaml:"-" bson:"-"` //only Cloud uses this csrf
}

func New(cfg Config) (*RestApiClient, *util.Result) {
	rac := &RestApiClient{Config: cfg}

	//rac.Config
	if rac.Config.Auth == nil {
		return nil, util.MsgError("Config", "no authentication info")
	}
	if rac.Config.ExtraHeaders == nil {
		rac.Config.ExtraHeaders = make(map[string]string)
	}
	if rac.Config.Auth.Xrf {
		rac.Config.ExtraHeaders["X-Qlik-Xrfkey"] = QlikXrfKey
	}
	rac.Config.ExtraHeaders["Content-type"] = "application/json"
	rac.Config.ExtraHeaders["Accept"] = "image/avif,image/webp,image/apng,image/svg+xml,image/*,application/json,*/*;q=0.8"

	if rac.Config.Cookie == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, util.Error("createCookieJar", err)
		}
		rac.Config.Cookie = jar
	}
	if rac.Config.TimeoutSec == nil {
		rac.Config.TimeoutSec = util.Ptr(300)
	}

	//rac.Logger
	if cfg.LogFileName != nil && *cfg.LogFileName != "" {
		var err error
		rac.Logger, err = loggers.GetLogger(*cfg.LogFileName)
		if err != nil {
			return nil, util.Error("NewLogger", err)
		}
	} else {
		rac.Logger = loggers.NullLogger
	}

	//baseURL
	u, err := url.Parse(cfg.BaseUrl)
	if err != nil {
		return nil, util.Error("ParseBaseUrl", err)
	}
	u.Path = path.Join(util.MaybeNil(cfg.VirtualProxy), util.MaybeNil(cfg.APIPrefix))
	rac.baseUrl = u

	//rac.certs
	if cfg.Auth.Certs != nil {
		kp, res := cfg.Auth.Certs.NewRsaKeyPair()
		if res != nil {
			return nil, res.With("NewRsaKeyPair")
		}
		rac.certs = kp
	}

	//rac.client
	tlsConfig := &tls.Config{}
	if rac.Config.ExtraTlsConfig != nil {
		rac.Config.ExtraTlsConfig.Apply(tlsConfig)
	}
	var transport *http.Transport
	transport = &http.Transport{TLSClientConfig: tlsConfig}
	rac.client = &http.Client{
		Transport: transport,
		Timeout:   time.Duration(time.Duration(*rac.Config.TimeoutSec) * time.Second),
		Jar:       rac.Config.Cookie,
	}

	//init.
	switch cfg.Auth.Method {
	case AuthMethodAPIKey:
		return rac.initApiKeyClient()
	case AuthMethodCert:
		return rac.initCertClient()
	case AuthMethodJWT:
		return rac.initJwtClient()
	case AuthMethodWebTicket:
		return rac.initWebTicketClient()
	}

	return nil, util.MsgError("NewRestApiClient", fmt.Sprintf("auth method %v is not supported", cfg.Auth.Method))
}

func (c *RestApiClient) initApiKeyClient() (*RestApiClient, *util.Result) {
	c.Config.ExtraHeaders["Authorization"] = fmt.Sprintf("Bearer %s", *c.Config.Auth.Token)

	return c, nil
}

func (c *RestApiClient) initCertClient() (*RestApiClient, *util.Result) {
	if c.Config.Auth.Certs == nil {
		return nil, util.MsgError("ParseCerts", "no certs in config")
	}

	dir := c.Config.Auth.User.Directory
	id := c.Config.Auth.User.Id
	c.Config.ExtraHeaders["X-Qlik-User"] = fmt.Sprintf("UserDirectory=%s; UserId=%s", dir, id)

	tlsCertsConfig, res := c.Config.Auth.Certs.NewTlsConfig()
	if res != nil {
		return nil, res.With("Certs.NewTlsConfig")
	}
	transport, ok := c.client.Transport.(*http.Transport)
	if !ok {
		return nil, util.MsgError("http.Transport", "can't convert to transport")
	}
	transport.TLSClientConfig.Certificates = tlsCertsConfig.Certificates
	transport.TLSClientConfig.RootCAs = tlsCertsConfig.RootCAs
	c.client.Transport = transport

	return c, nil
}

func (c *RestApiClient) initJwtClient() (*RestApiClient, *util.Result) {
	var res *util.Result
	if util.MaybeNil(c.Config.IsCloud) {
		res = c.GenerateCloudJWT()
		if res != nil {
			return nil, res.With("GenerateCloudJWT")
		}
		c.Config.ExtraHeaders["Authorization"] = fmt.Sprintf("Bearer %s", *c.Config.Auth.Token)
		res = c.StartCloudJWTSession()
		if res != nil {
			return nil, res.With("StartCloudJWTSession")
		}
	} else {
		res = c.GenerateManagedJWT()
		c.Config.ExtraHeaders["Authorization"] = fmt.Sprintf("Bearer %s", *c.Config.Auth.Token)
		if res != nil {
			return nil, res.With("GenerateManagedJWT")
		}
	}

	return c, nil
}

func (c *RestApiClient) initWebTicketClient() (*RestApiClient, *util.Result) {
	return c, nil
}

func (c *RestApiClient) GetHttpClient() *http.Client {
	return c.client
}

func (c RestApiClient) GetJWT() string {
	return util.MaybeNil(c.Config.Auth.Token)
}

func (c RestApiClient) GetCookieJar() http.CookieJar {
	if c.client != nil {
		return c.client.Jar
	}
	return nil
}

func (c *RestApiClient) SetCookieJar(jar http.CookieJar) {
	if c.client != nil {
		c.client.Jar = jar
	}
}

func (c RestApiClient) GetCloudCsrfToken() string {
	return c.csrfToken
}

func (c *RestApiClient) GenerateCloudJWT() *util.Result {
	if !util.MaybeNil(c.Config.IsCloud) {
		return util.MsgError("CheckCloudJWTConfig", "config is not for cloud")
	}
	if c.Config.Auth.Method != AuthMethodJWT {
		return util.MsgError("CheckCloudJWTConfig", "not jwt auth method")
	}
	if c.Config.Auth.CloudJwt == nil {
		return util.MsgError("CheckCloudJWTConfig", "there's no cloud jwt config")
	}
	if !c.Config.Auth.CloudJwt.IsValid() {
		return util.MsgError("CheckCloudJWTConfig", "invalid cloud jwt config")
	}
	if c.certs == nil || c.certs.Key == nil {
		return util.MsgError("CheckCloudJWTConfig", "invalid certs config")
	}

	jwt, res := c.Config.Auth.CloudJwt.GenerateJWT(c.certs.Key)
	if res != nil {
		return res.With("GenerateManagedJWT")
	}
	c.Config.Auth.Token = &jwt
	return nil
}

func (c *RestApiClient) GenerateManagedJWT() *util.Result {
	if c.Config.Auth.Method != AuthMethodJWT {
		return util.MsgError("CheckJWTConfig", "not jwt auth method")
	}
	if c.certs == nil || c.certs.Key == nil {
		return util.MsgError("CheckJWTConfig", "invalid certs config")
	}
	if c.Config.Auth.User == nil {
		return util.MsgError("CheckJWTConfig", "no user config")
	}

	jwtPayload := qlik.JwtClaim{
		UserID:        &c.Config.Auth.User.Id,
		UserDirectory: &c.Config.Auth.User.Directory,
	}

	jwt, res := jwtPayload.GetJWT(c.certs.Key)
	if res != nil {
		return res.With("GenerateManagedJWT")
	}
	c.Config.Auth.Token = &jwt
	return nil
}

func (c *RestApiClient) StartCloudJWTSession() *util.Result {
	//Login
	_, _, res := c.Do(http.MethodPost, "*^/login/jwt-session", nil,
		WithParam("qlik-web-integration-id", c.Config.Auth.CloudJwt.WebIntegrationID),
		WithHeader("Qlik-Web-Integration-ID", c.Config.Auth.CloudJwt.WebIntegrationID))
	if res != nil {
		return res.With("DoLogin")
	}

	//GetCSRFToken
	resp, _, res := c.Do(http.MethodGet, "*^/api/v1/csrf-token", nil,
		WithParam("qlik-web-integration-id", c.Config.Auth.CloudJwt.WebIntegrationID),
		WithHeader("Qlik-Web-Integration-ID", c.Config.Auth.CloudJwt.WebIntegrationID))
	if res != nil {
		return res.With("DoGetCSRF")
	}
	if resp.StatusCode > 220 {
		return util.MsgError("DoCSRFRequest", resp.Status)
	}

	c.csrfToken = resp.Header.Get("qlik-csrf-token")
	c.Config.Auth.CloudJwt.CsrfToken = c.csrfToken

	return nil
}

func (c *RestApiClient) IsCloud() bool {
	return c.Config.IsForCloud()
}
