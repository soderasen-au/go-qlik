package qps

import (
	"encoding/json"
	"io"

	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/qlik/rac"
)

const (
	QLIK_XRF_KEY string = "ponmlkjihgfedcba"
)

type Client struct {
	impl *rac.RestApiClient
}

func NewClient(cfg rac.Config) (*Client, *util.Result) {
	cfg.Auth.Xrf = true
	restClient, res := rac.New(cfg)
	if res != nil {
		return nil, res.With("NewRAC")
	}
	c := &Client{impl: restClient}

	return c, nil
}

func (c *Client) Logger() *zerolog.Logger {
	return c.impl.Logger
}

func (c *Client) SetLogger(l *zerolog.Logger) {
	c.impl.Logger = l
}

func (c *Client) Do(method string, endpoint string, params map[string]string, body interface{}) ([]byte, *util.Result) {
	_, buf, res := c.impl.Do(method, endpoint, body, rac.WithParams(params))
	if res != nil {
		return nil, res.With("Do")
	}

	return buf, res
}

func (c *Client) DoRaw(method string, endpoint string, extraHeaders map[string]string, params map[string]string, body io.Reader) ([]byte, *util.Result) {
	_, buf, res := c.impl.Do(method, endpoint, body, rac.WithParams(params), rac.WithHeaders(extraHeaders))
	if res != nil {
		return nil, res.With("Do")
	}

	return buf, res
}

func (c *Client) Get(endpoint string, params map[string]string) ([]byte, *util.Result) {
	return c.Do("GET", endpoint, params, nil)
}

func (c *Client) GetRaw(endpoint string, extraHeaders map[string]string, params map[string]string) ([]byte, *util.Result) {
	return c.DoRaw("GET", endpoint, extraHeaders, params, nil)
}

func (c *Client) GetObject(endpoint string, o interface{}) *util.Result {
	buf, res := c.Do("GET", endpoint, nil, nil)
	if res != nil {
		return res.With("Do")
	}

	err := json.Unmarshal(buf, o)
	if err != nil {
		return util.Error("Unmarshal", err)
	}
	return nil
}

func (c *Client) Post(endpoint string, params map[string]string, body interface{}) ([]byte, *util.Result) {
	return c.Do("POST", endpoint, params, body)
}

func (c *Client) PostRaw(endpoint string, extraHeaders map[string]string, params map[string]string, body io.Reader) ([]byte, *util.Result) {
	return c.DoRaw("POST", endpoint, extraHeaders, params, body)
}
