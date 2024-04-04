package engine

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-common/util"

	"github.com/rs/zerolog"

	"github.com/soderasen-au/go-qlik/qlik"
)

type HttpClient struct {
	client       *http.Client
	BaseUrl      *url.URL
	ExtraHeaders http.Header
	Logger       *zerolog.Logger
}

func NewHttpClient(cfg Config) (*HttpClient, *util.Result) {
	ret := &HttpClient{ExtraHeaders: make(http.Header), Logger: loggers.CoreDebugLogger}

	ret.ExtraHeaders.Set("Content-type", "application/json")
	ret.ExtraHeaders.Set("Accept", "application/json")

	var transport *http.Transport
	if cfg.AuthMode == AUTH_MODE_JWT {
		jwtPayload := qlik.JwtClaim{
			UserID:        &cfg.UserName,
			UserDirectory: &cfg.UserDirectory,
		}
		keypair, res := cfg.Certs.NewRsaKeyPair()
		if res != nil {
			return nil, res.With("NewRsaKeyPair")
		}
		jwt, res := jwtPayload.GetJWT(keypair.Key)
		if res != nil {
			return nil, res.With("GetJWT")
		}
		ret.ExtraHeaders.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))

		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		transport = &http.Transport{TLSClientConfig: tlsConfig}
	} else {
		ret.ExtraHeaders.Add("X-Qlik-User", fmt.Sprintf("UserDirectory=%s; UserId=%s", cfg.UserDirectory, cfg.UserName))

		tlsConfig, err := cfg.Certs.NewTlsConfig()
		if err != nil {
			return nil, util.Error("NewTlsConfig", err)
		}
		transport = &http.Transport{TLSClientConfig: tlsConfig}
	}

	ret.client = &http.Client{
		Timeout:   time.Duration(300 * time.Second),
		Transport: transport,
	}

	var res *util.Result
	ret.BaseUrl, res = cfg.GetHttpsBaseUrl()
	if res != nil {
		return nil, util.Error("GetHttpsBaseUrl", res)
	}

	return ret, nil
}

func (c *HttpClient) GetUrl(endpoint string) string {
	u := *c.BaseUrl
	u.Path = path.Join(u.Path, endpoint)
	return u.String()
}

func (c *HttpClient) NewRequest(method string, endpoint string, params map[string]string, body interface{}) (*http.Request, *util.Result) {
	var bodyReader *bytes.Buffer
	epUrl := c.GetUrl(endpoint)

	var err error
	var req *http.Request
	if body != nil {
		marshaledBody, err := json.Marshal(body)
		if err != nil {
			return nil, util.Error("MarshalRequestBody", err)
		}
		bodyReader = bytes.NewBuffer(marshaledBody)
		req, err = http.NewRequest(method, epUrl, bodyReader)
	} else {
		req, err = http.NewRequest(method, epUrl, nil)
		if err != nil {
			return nil, util.Error("can't create request", err)
		}
	}

	q := req.URL.Query()
	for pk, pv := range params {
		q.Add(pk, pv)
	}
	req.URL.RawQuery = q.Encode()

	req.Header = c.ExtraHeaders
	return req, nil
}

func (c *HttpClient) DoRequest(req *http.Request) ([]byte, *util.Result) {
	c.Logger.Trace().Msgf("engine https client ===> %v", req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, util.Error("client.Do", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, util.MsgError("StatusCode", fmt.Sprintf("%d", resp.StatusCode))
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, util.Error("io.ReadAll", err)
	}
	c.Logger.Trace().Msgf("engine https client <=== %v", string(buf))

	return buf, nil
}

func (c *HttpClient) Do(method string, endpoint string, params map[string]string, body interface{}) ([]byte, *util.Result) {
	req, res := c.NewRequest(method, endpoint, params, body)
	if res != nil {
		return nil, res.With("NewRequest")
	}

	return c.DoRequest(req)
}

func (c *HttpClient) Get(endpoint string, params map[string]string) ([]byte, *util.Result) {
	return c.Do("GET", endpoint, params, nil)
}
