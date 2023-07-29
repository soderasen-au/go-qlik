package qcs

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/qlik-oss/enigma-go/v3"
	"github.com/soderasen-au/go-qlik/qlik/rac"

	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/qlik/engine"
)

func (c Client) GetEngineBaseUrl() (*url.URL, *util.Result) {
	baseUrl, err := url.Parse(c.Config.BaseUrl)
	if err != nil {
		return nil, util.Error("ParseTenantBaseUrl", err)
	}

	baseUrl.Scheme = "wss"
	baseUrl.Host = baseUrl.Hostname() //no port on cloud
	baseUrl.Path = "/app"

	return baseUrl, nil
}

func (c *Client) NewEngineConn(appId string) (conn *engine.Conn, res *util.Result) {
	if util.MaybeNil(c.Config.Auth.Token) == "" {
		return nil, util.MsgError("CheckCloudJWTConfig", "there's no JWT")
	}

	baseUrl, res := c.GetEngineBaseUrl()
	if res != nil {
		return nil, res.With("GetEngineBaseUrl")
	}
	baseUrl.Path = path.Join(baseUrl.Path, appId)
	if c.Config.Auth.Method == rac.AuthMethodJWT {
		queries := baseUrl.Query()
		queries.Add("qlik-web-integration-id", c.Config.Auth.CloudJwt.WebIntegrationID)
		queries.Add("qlik-csrf-token", c.client.GetCloudCsrfToken())
		baseUrl.RawQuery = queries.Encode()
	}
	cfg := engine.Config{
		EngineURI:     baseUrl.String(),
		AppID:         appId,
		QRSBaseURI:    "",
		UserName:      c.Config.Auth.User.Id,
		UserDirectory: c.Config.Auth.User.Directory,
		AuthMode:      engine.AUTH_MODE_JWT,
		ServerType:    engine.ST_CLOUD,
		JWT:           *c.Config.Auth.Token,
	}

	headers := make(http.Header, 1)
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.JWT))

	global, err := enigma.Dialer{Jar: c.client.GetCookieJar()}.Dial(engine.ConnCtx, cfg.EngineURI, headers)
	if err != nil {
		return nil, util.Error("enigma.Dialer", err)
	}
	version, _ := global.EngineVersion(engine.ConnCtx)
	if version == nil {
		return nil, util.MsgError("CheckEngineConn", "connection is invalid. (user may not have access to appid)")
	}
	conn = &engine.Conn{
		Global: global,
		Cfg:    cfg,
	}
	return conn, nil
}
