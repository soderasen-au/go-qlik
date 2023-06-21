package rac

import (
	"bytes"
	"encoding/json"
	"github.com/Click-CI/common/util"
	"io"
	"net/http"
	"path"
	"strings"
)

const (
	REQ_ROOT_PATH_PREFIX string = "^"
	REQ_HOST_PATH_PREFIX string = "*"
)

// GetUrl endpoint starts with `*` means host endpoint (ignore any port number)
// endpoint starts with `^` means root endpoint (from / path)
// `*^` is accepted as host root endpoint
// `^*` is invalid
func (c RestApiClient) GetUrl(endpoint string) string {
	u := *c.baseUrl
	if strings.HasPrefix(endpoint, REQ_HOST_PATH_PREFIX) {
		u.Host = u.Hostname()
		endpoint = strings.TrimLeft(endpoint, REQ_HOST_PATH_PREFIX)
	}
	if strings.HasPrefix(endpoint, REQ_ROOT_PATH_PREFIX) {
		u.Path = "/"
		endpoint = strings.TrimLeft(endpoint, REQ_ROOT_PATH_PREFIX)
	}

	u.Path = path.Join(u.Path, endpoint)
	return u.String()
}

func GetRootPath(path string) string {
	return REQ_ROOT_PATH_PREFIX + path
}

func GetHostPath(path string) string {
	return REQ_HOST_PATH_PREFIX + path
}

func GetHostRootPath(path string) string {
	return GetHostPath(GetRootPath(path))
}

type RequestOption func(*http.Request) *http.Request

func WithParam(k, v string) RequestOption {
	return func(req *http.Request) *http.Request {
		query := req.URL.Query()
		query.Set(k, v)
		req.URL.RawQuery = query.Encode()
		return req
	}
}

func WithParams(params map[string]string) RequestOption {
	return func(req *http.Request) *http.Request {
		if params != nil {
			query := req.URL.Query()
			for k, v := range params {
				query.Set(k, v)
			}
			req.URL.RawQuery = query.Encode()
		}
		return req
	}
}

func WithHeader(k, v string) RequestOption {
	return func(req *http.Request) *http.Request {
		req.Header.Set(k, v)
		return req
	}
}

func WithHeaders(headers map[string]string) RequestOption {
	return func(req *http.Request) *http.Request {
		if headers != nil {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		}
		return req
	}
}

func (c RestApiClient) AddQlikOptions(req *http.Request, Opts ...RequestOption) {
	query := req.URL.Query()
	if c.Config.Auth.Xrf {
		query.Set("Xrfkey", QlikXrfKey)
	}
	if c.Config.Auth.Method == AuthMethodJWT && util.MaybeNil(c.Config.IsCloud) {
		if c.csrfToken != "" {
			query.Set("qlik-csrf-token", c.csrfToken)
		}
	}
	req.URL.RawQuery = query.Encode()

	for k, v := range c.Config.ExtraHeaders {
		req.Header.Set(k, v)
	}

	if Opts != nil {
		for _, opt := range Opts {
			if opt != nil {
				_ = opt(req)
			}
		}
	}

	return
}

func (c *RestApiClient) NewRawRequest(method, url string, body interface{}) (*http.Request, *util.Result) {
	var err error
	var req *http.Request
	if readerBody, isIOReader := body.(io.Reader); isIOReader {
		req, err = http.NewRequest(method, url, readerBody)
	} else if body != nil {
		marshaledBody, _err := json.Marshal(body)
		if _err != nil {
			return nil, util.Error("MarshalRequestBody", _err)
		}
		bodyReader := bytes.NewBuffer(marshaledBody)
		req, err = http.NewRequest(method, url, bodyReader)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, util.Error("http.NewRequest", err)
	}

	return req, nil
}

// NewRequest accept object as body since it set json content type header in constructor
func (c *RestApiClient) NewRequest(method, endpoint string, body interface{}, Opts ...RequestOption) (*http.Request, *util.Result) {
	url := c.GetUrl(endpoint)

	req, res := c.NewRawRequest(method, url, body)
	if res != nil {
		return nil, res.With("NewRawRequest")
	}
	c.AddQlikOptions(req, Opts...)

	return req, nil
}

func (c *RestApiClient) DoRequest(req *http.Request) (*http.Response, []byte, *util.Result) {
	c.Logger.Trace().Msgf("Request ===> %v", req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, util.Error("Do", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, nil, util.Errorf("%d: %s", resp.StatusCode, resp.Status)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, util.Error("ReadResp", err)
	}
	c.Logger.Trace().Msgf("Response <=== %v", string(buf))

	return resp, buf, nil
}

func (c *RestApiClient) Do(method, endpoint string, body interface{}, opts ...RequestOption) (*http.Response, []byte, *util.Result) {
	req, res := c.NewRequest(method, endpoint, body, opts...)
	if res != nil {
		return nil, nil, res.With("NewRequest")
	}
	resp, buf, res := c.DoRequest(req)
	if res != nil {
		return nil, nil, res.With("Do")
	}
	return resp, buf, nil
}
