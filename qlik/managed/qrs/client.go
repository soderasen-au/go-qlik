package qrs

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/qlik/engine"
	"github.com/soderasen-au/go-qlik/qlik/rac"
)

const (
	QLIK_XRF_KEY string = "abcdefghijklmnop"
)

type Client struct {
	Cfg    Config `json:"config" yaml:"config"`
	client *rac.RestApiClient
}

func NewClient(cfg Config) (*Client, *util.Result) {
	cfg.Auth.Xrf = true
	c := &Client{Cfg: cfg}

	client, res := rac.New(cfg.Config)
	if res != nil {
		return nil, res.With("NewRAC")
	}
	c.client = client

	_, res = c.About()
	if res != nil {
		return nil, res.With("About")
	}
	return c, nil
}

func NewFromEngine(cfg engine.Config) (*Client, *util.Result) {
	if !cfg.IsOnPrem() {
		return nil, util.MsgError("new qrs client", "qrs client only supports on-prem Qlik")
	}

	if cfg.AuthMode != engine.AUTH_MODE_CERT {
		return nil, util.Errorf("QRS only support Cert authentication")
	}

	qrsCfg := NewConfigFromEngine(cfg)

	return NewClient(*qrsCfg)
}

func (c *Client) IsUsingCert() bool {
	return c.Cfg.Auth.Method == rac.AuthMethodCert
}

func (c *Client) Logger() *zerolog.Logger {
	return c.client.Logger
}

func (c *Client) SetLogger(l *zerolog.Logger) {
	c.client.Logger = l
}

func (c *Client) Do(method string, endpoint string, params map[string]string, body interface{}) ([]byte, *util.Result) {
	_, resp, res := c.client.Do(method, endpoint, body, rac.WithParams(params))
	return resp, res
}

func (c *Client) Get(endpoint string, opts ...rac.RequestOption) ([]byte, *util.Result) {
	_, buf, res := c.client.Do(http.MethodGet, endpoint, nil, opts...)
	if res != nil {
		return nil, res.With("rac.Do")
	}
	return buf, nil
}

func (c *Client) GetObject(endpoint string, o interface{}) *util.Result {
	buf, res := c.Get(endpoint)
	if res != nil {
		return res.With("Get")
	}
	// resp := string(buf)
	// fmt.Println("resp", resp)
	err := json.Unmarshal(buf, o)
	if err != nil {
		return util.Error("Unmarshal", err)
	}
	return nil
}

func (c *Client) Post(endpoint string, body interface{}, opts ...rac.RequestOption) ([]byte, *util.Result) {
	_, buf, res := c.client.Do(http.MethodPost, endpoint, body, opts...)
	if res != nil {
		return nil, res.With("rac.Do")
	}
	return buf, nil
}
