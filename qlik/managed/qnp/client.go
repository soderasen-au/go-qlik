package qnp

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"time"

	"github.com/soderasen-au/go-common/loggers"

	"github.com/soderasen-au/go-qlik/qlik"

	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"
)

const (
	QLIK_XRF_KEY string = "abcdefghijklmnop"
)

type Client struct {
	Cfg     Config          `json:"config" yaml:"config"`
	client  *http.Client    `json:"-" yaml:"-"`
	headers http.Header     `json:"-" yaml:"-"`
	Logger  *zerolog.Logger `json:"-" yaml:"-"`
}

func NewClient(cfg Config) (*Client, *util.Result) {
	c := &Client{Cfg: cfg}
	if res := c.Cfg.ParseKeyPair(); res != nil {
		return nil, res.With("ParseKeyPair")
	}

	c.headers = make(http.Header)
	c.headers.Add("Content-type", "application/json")
	c.headers.Add("Accept", "application/json")

	options := cookiejar.Options{}

	jar, err := cookiejar.New(&options)
	if err != nil {
		return nil, util.Error("NewCookieJar", err)
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	c.client = &http.Client{
		Jar:       jar,
		Transport: transport,
		Timeout:   time.Duration(300 * time.Second),
	}

	c.Logger = loggers.NullLogger

	return c, nil
}

func (c *Client) doNewRequest(token *qlik.JwtClaim, method string, url string, extraHeaders map[string]string, params map[string]string, body interface{}) (*http.Request, *util.Result) {
	var err error
	var req *http.Request
	if body != nil {
		marshaledBody, _ := json.Marshal(body)
		bodyReader := bytes.NewBuffer(marshaledBody)
		req, err = http.NewRequest(method, url, bodyReader)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, util.Error("NewRequest", err)
	}

	q := req.URL.Query()
	for pk, pv := range params {
		q.Add(pk, pv)
	}
	req.URL.RawQuery = q.Encode()

	req.Header = c.headers
	if extraHeaders != nil {
		for hk, hv := range extraHeaders {
			req.Header.Set(hk, hv)
		}
	}

	usr := token
	if usr == nil {
		usr = c.Cfg.User
	}
	if usr != nil {
		if c.Cfg.RsaKeyPair.Key == nil {
			return nil, util.MsgError("GetPrivateKey", "no private key")
		}
		jwt, res := usr.GetJWT(c.Cfg.RsaKeyPair.Key)
		if res != nil {
			return nil, res.With("GetJWT")
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	}

	return req, nil
}

func (c *Client) newRequest(token *qlik.JwtClaim, method string, endpoint string, params map[string]string, body interface{}) (*http.Request, *util.Result) {
	u, res := util.GetAPIUrl(c.Cfg.BaseURI, endpoint)
	if res != nil {
		return nil, res.With("GetAPIUrl")
	}

	return c.doNewRequest(token, method, u, nil, params, body)
}

func (c *Client) newNewsStandRequest(token *qlik.JwtClaim, method string, endpoint string, params map[string]string, body interface{}) (*http.Request, *util.Result) {
	u, res := util.GetAPIUrl(c.Cfg.NewsStandURI, endpoint)
	if res != nil {
		return nil, res.With("GetAPIUrl")
	}

	return c.doNewRequest(token, method, u, nil, params, body)
}

func (c *Client) DoRequest(req *http.Request) ([]byte, *util.Result) {
	c.Logger.Debug().Msgf("QNP ===> %v", req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, util.Error("HttpDo", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, util.Error("ReponseError", fmt.Errorf("%d: %s", resp.StatusCode, resp.Status))
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, util.Error("ReadResponse", err)
	}
	c.Logger.Debug().Msgf("QNP <=== %v", string(buf))

	return buf, nil
}

func (c *Client) DownloadFile(url string, folder string) (downloadedFile string, res *util.Result) {
	c.Logger.Debug().Msgf("DownloadFile(%v): %s to %s ...", c.Cfg.User, url, folder)

	req, res := c.doNewRequest(c.Cfg.User, http.MethodGet, url, nil, nil, nil)
	if res != nil {
		return "", res.With("doNewRequest")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", util.Error("Do", err)
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var filename string
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			filename = params["filename"]
		} else {
			return "", util.Error("ParseMediaType", err)
		}
	}

	downloadedFile = filepath.Join(folder, filename)
	out, err := os.Create(downloadedFile)
	if err != nil {
		return "", util.Error("CreateFile", err)
	}
	defer out.Close()

	bodyBuffer := bytes.NewBuffer(data)
	io.Copy(out, bodyBuffer)

	fi, err := os.Stat(downloadedFile)
	size := fi.Size()
	if err != nil || size == 0 {
		return "", util.Errorf("Error downloading file. 0k or does not exist on disk: %+v", downloadedFile)
	}

	return downloadedFile, nil
}

func (c *Client) Do(method string, endpoint string, params map[string]string, body interface{}) ([]byte, *util.Result) {
	req, res := c.newRequest(nil, method, endpoint, params, body)
	if res != nil {
		return nil, res.With("newRequest")
	}

	return c.DoRequest(req)
}

func (c *Client) DoFor(usr *qlik.JwtClaim, method string, endpoint string, params map[string]string, body interface{}) ([]byte, *util.Result) {
	req, res := c.newRequest(usr, method, endpoint, params, body)
	if res != nil {
		return nil, res.With("newRequest")
	}

	return c.DoRequest(req)
}

func (c *Client) Get(endpoint string, params map[string]string) ([]byte, *util.Result) {
	return c.Do("GET", endpoint, params, nil)
}

func (c *Client) GetFor(usr *qlik.JwtClaim, endpoint string, params map[string]string) ([]byte, *util.Result) {
	return c.DoFor(usr, "GET", endpoint, params, nil)
}

func (c *Client) Post(endpoint string, params map[string]string, body interface{}) ([]byte, *util.Result) {
	return c.Do("POST", endpoint, params, body)
}

func (c *Client) PostFor(usr *qlik.JwtClaim, endpoint string, params map[string]string, body interface{}) ([]byte, *util.Result) {
	return c.DoFor(usr, "POST", endpoint, params, body)
}

func (c *Client) NSDo(method string, endpoint string, params map[string]string, body interface{}) ([]byte, *util.Result) {
	req, res := c.newNewsStandRequest(nil, method, endpoint, params, body)
	if res != nil {
		return nil, res
	}

	return c.DoRequest(req)
}

func (c *Client) NSDoFor(usr *qlik.JwtClaim, method string, endpoint string, params map[string]string, body interface{}) ([]byte, *util.Result) {
	req, res := c.newNewsStandRequest(usr, method, endpoint, params, body)
	if res != nil {
		return nil, res
	}

	return c.DoRequest(req)
}

func (c *Client) NSGet(endpoint string, params map[string]string) ([]byte, *util.Result) {
	return c.NSDo("GET", endpoint, params, nil)
}

func (c *Client) NSGetFor(usr *qlik.JwtClaim, endpoint string, params map[string]string) ([]byte, *util.Result) {
	return c.NSDoFor(usr, "GET", endpoint, params, nil)
}

func (c *Client) NSPost(endpoint string, params map[string]string, body interface{}) ([]byte, *util.Result) {
	return c.NSDo("POST", endpoint, params, body)
}

func (c *Client) NSPostFor(usr *qlik.JwtClaim, endpoint string, params map[string]string, body interface{}) ([]byte, *util.Result) {
	return c.NSDoFor(usr, "POST", endpoint, params, body)
}

func (c *Client) DoRaw(usr *qlik.JwtClaim, method, endpoint string, extraHeaders, params map[string]string, body interface{}) ([]byte, *util.Result) {
	u, res := util.GetUrl(c.Cfg.BaseURI, endpoint)
	if res != nil {
		return nil, res.With("GetUrl")
	}

	req, res := c.doNewRequest(usr, method, u, extraHeaders, params, body)
	if res != nil {
		return nil, res.With("doNewRequest")
	}

	return c.DoRequest(req)
}

func (c *Client) NSDoRaw(usr *qlik.JwtClaim, method, endpoint string, extraHeaders, params map[string]string, body interface{}) ([]byte, *util.Result) {
	u, res := util.GetUrl(c.Cfg.NewsStandURI, endpoint)
	if res != nil {
		return nil, res.With("GetUrl")
	}

	req, res := c.doNewRequest(usr, method, u, extraHeaders, params, body)
	if res != nil {
		return nil, res.With("doNewRequest")
	}

	return c.DoRequest(req)
}
