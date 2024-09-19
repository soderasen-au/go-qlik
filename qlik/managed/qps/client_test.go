package qps

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"
	"gopkg.in/yaml.v3"

	"github.com/soderasen-au/go-qlik/qlik/rac"
)

var (
	happinessAppID string = "08df6142-ab4e-4e91-88e8-e32682d3bfcd"
)

func setupTestSuite(conf string, internal bool, t *testing.T) (*Client, *zerolog.Logger, func(t2 *testing.T)) {
	wd, _ := os.Getwd()
	buf, err := os.ReadFile(conf)
	if err != nil {
		t.Errorf("can't load config file: %s", err.Error())
	}

	var cfg rac.Config
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		t.Errorf("can't parse config file: %s", err.Error())
	}

	if internal {
		base, err := url.Parse(cfg.BaseUrl)
		if err != nil {
			t.Errorf("invalid base url: %s", err.Error())
		}
		base.Host = base.Hostname() + ":4243"
		cfg.BaseUrl = base.String()
		cfg.VirtualProxy = util.Ptr("/")
	}

	client, res := NewClient(cfg)
	if res != nil {
		t.Errorf("can't parse config file: %s", res.Error())
	}
	_logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}.Out).Level(zerolog.TraceLevel)
	client.SetLogger(&_logger)
	_logger.Info().Msgf("test in %s using %s", wd, conf)

	return client, &_logger, func(t2 *testing.T) {}
}

func TestNewClient(t *testing.T) {
	client, _, tearDown := setupTestSuite("../../../test/qps/localhost.yaml", true, t)
	defer tearDown(t)
	t.Run("GetOpenApiSpec", func(t *testing.T) {
		openapi, res := client.Get("/about/openapi/main", nil)
		if res != nil {
			t.Errorf("GetOpenApiSpec failed: %s", res.Error())
		}
		var out bytes.Buffer
		_ = json.Indent(&out, openapi, "", "  ")
		_ = os.WriteFile("../../../test/qps/proxy-api.json", out.Bytes(), fs.ModePerm)

		personal, res := client.Get("/about/openapi/personal", nil)
		if res != nil {
			t.Errorf("GetOpenApiSpec failed: %s", res.Error())
		}
		_ = json.Indent(&out, personal, "", "  ")
		_ = os.WriteFile("../../../test/qps/personal-api.json", out.Bytes(), fs.ModePerm)
	})
}
