package qcs

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/eventials/go-tus"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-qlik/qlik"
	"github.com/soderasen-au/go-qlik/qlik/engine"
	"github.com/soderasen-au/go-qlik/qlik/rac"

	"github.com/Click-CI/common/util"
	"github.com/eventials/go-tus/memorystore"
)

type Client struct {
	Config    rac.Config
	client    *rac.RestApiClient
	tusClient *tus.Client
}

func NewClient(cfg rac.Config) (*Client, *util.Result) {
	if cfg.IsCloud == nil {
		cfg.IsCloud = util.Ptr(false)
	}
	if !(*cfg.IsCloud) {
		return nil, util.MsgError("Check", "not cloud config")
	}

	if cfg.APIPrefix == nil {
		cfg.APIPrefix = util.Ptr("/api/v1")
	}

	qc, res := rac.New(cfg)
	if res != nil {
		return nil, res.With("NewRAC")
	}

	store, err := memorystore.NewMemoryStore()
	if err != nil {
		return nil, util.Error("NewMemoryStore", err)
	}
	tusConfig := tus.DefaultConfig()
	tusConfig.ChunkSize = 100 * 1024 * 1024
	tusConfig.Resume = true
	tusConfig.Store = store
	tusConfig.Header.Add("Authorization", fmt.Sprintf("Bearer %s", qc.GetJWT()))
	tusConfig.HttpClient = qc.GetHttpClient()
	tusClient, err := tus.NewClient(qc.GetUrl("/temp-contents/files"), tusConfig)
	if err != nil {
		return nil, util.Error("New tus client", err)
	}

	return &Client{Config: cfg, client: qc, tusClient: tusClient}, nil
}

func NewConfigFromEngine(cfg engine.Config) *rac.Config {
	baseURL, _ := url.Parse(cfg.QRSBaseURI)
	baseURL.Path = ""
	return &rac.Config{
		BaseUrl:   baseURL.String(),
		APIPrefix: util.Ptr("/api/v1"),
		Auth: &rac.AuthConfig{
			Method: rac.AuthMethodAPIKey,
			Xrf:    false,
			User: &qlik.User{
				Id:        cfg.UserName,
				Directory: cfg.UserDirectory,
			},
			Token: &cfg.JWT,
		},
		IsCloud:        util.Ptr(true),
		ExtraTlsConfig: &rac.ExtraTLSConfig{InsecureSkipVerify: true},
		TimeoutSec:     util.Ptr(300),
	}
}

func NewFromEngine(cfg engine.Config) (*Client, *util.Result) {
	if !cfg.IsCloud() {
		return nil, util.MsgError("Check", "engine config is not for cloud")
	}
	if cfg.AuthMode == engine.AUTH_MODE_JWT && cfg.JWT == "" {
		return nil, util.MsgError("Check", "no jwt")
	}
	_, err := url.Parse(cfg.QRSBaseURI)
	if err != nil {
		return nil, util.Error("CheckBaseURI", err)
	}

	return NewClient(*NewConfigFromEngine(cfg))
}

func (c *Client) SetCookieJar(jar http.CookieJar) {
	if c.client != nil {
		c.client.SetCookieJar(jar)
	}
}

func (c *Client) NewRequest(method, endpoint string, params map[string]string) *http.Request {
	req, _ := c.client.NewRequest(method, endpoint, nil, rac.WithParams(params))
	return req
}

//
//func (c *Client) NewRawRequest(method, _url string, params, extraHeaders map[string]string, body io.Reader) (*http.Request, *util.Result) {
//	req, err := http.NewRequest(strings.ToUpper(method), _url, body)
//	if err != nil {
//		return nil, util.Error("NewHttpRequest", err)
//	}
//	if params != nil && len(params) > 0 {
//		query := req.URL.Query()
//		for k, v := range params {
//			query.Set(k, v)
//		}
//		req.URL.RawQuery = query.Encode()
//	}
//	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Config.Auth.JWT))
//	for hk, hv := range extraHeaders {
//		req.Header.Set(hk, hv)
//	}
//
//	return req, nil
//}
//
//func (c *Client) NewJsonRequest(method, endpoint string, params map[string]string, body []byte) *http.Request {
//	req := c.NewRequest(method, endpoint, params)
//	req.Header.Set("Content-type", "application/json")
//	req.Body = ioutil.NopCloser(bytes.NewReader(body))
//	return req
//}

func (c *Client) Do(req *http.Request) (*http.Response, []byte, *util.Result) {
	return c.client.DoRequest(req)
}

func (c *Client) Get(endpoint string, params map[string]string) (*http.Response, []byte, *util.Result) {
	return c.client.Do(http.MethodGet, endpoint, rac.WithParams(params))
}

func (c *Client) GetRawUrl(rawUrl string) (*http.Response, []byte, *util.Result) {
	req, res := c.client.NewRawRequest(http.MethodGet, rawUrl, nil)
	if res != nil {
		return nil, nil, res.With("NewRawRequest")
	}
	return c.Do(req)
}

func (c *Client) HostUrl(uri string) (string, *util.Result) {
	return c.client.GetUrl(rac.GetHostPath(uri)), nil
}

func (c Client) Logger() *zerolog.Logger {
	return c.client.Logger
}

func (c *Client) SetLogger(_l *zerolog.Logger) {
	c.client.Logger = _l
}

func (c *Client) UploadToTCS(filePath string) (fileId string, res *util.Result) {
	c.Logger().Info().Msgf("uploading %s to %s ...", filePath, c.tusClient.Url)

	f, err := os.Open(filePath)
	if err != nil {
		return "", util.Error("OpenFile", err)
	}
	defer f.Close()
	uploadTask, err := tus.NewUploadFromFile(f)
	if err != nil {
		return "", util.Error("NewTusUploadTask", err)
	}
	c.Logger().Info().Msgf("uploadTask task fingerprint: %s", uploadTask.Fingerprint)

	uploader, err := c.tusClient.CreateUpload(uploadTask)
	if err != nil {
		return "", util.Error("NewTusUploader", err)
	}
	location, ok := c.tusClient.Config.Store.Get(uploadTask.Fingerprint)
	if !ok {
		return "", util.MsgError("NewTusUploader", "no file location is found")
	}
	c.Logger().Info().Msgf("uploader file location: %s", location)

	err = uploader.Upload()
	if err != nil {
		return "", util.Error("NewTusUploader", err)
	}
	fileParts := strings.Split(location, "/")
	if len(fileParts) < 1 {
		return "", util.MsgError("ParseFileID", "invalid file location")
	}
	fileID := fileParts[len(fileParts)-1]
	c.Logger().Info().Msgf("uploadTask succeeded now start importing %s", fileID)

	return fileID, nil
}

// export
