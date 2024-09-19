package rac

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

func setupTestSuite(conf string, t *testing.T) (*RestApiClient, *zerolog.Logger, func(t2 *testing.T)) {
	wd, _ := os.Getwd()
	buf, err := os.ReadFile(conf)
	if err != nil {
		t.Errorf("can't load config file: %s", err.Error())
		return nil, nil, nil
	}

	var cfg Config
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		t.Errorf("can't parse config file: %s", err.Error())
		return nil, nil, nil
	}

	client, res := New(cfg)
	if res != nil {
		t.Errorf("can't parse config file: %s", res.Error())
		return nil, nil, nil
	}
	_logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}.Out).Level(zerolog.DebugLevel)
	client.Logger = &_logger
	_logger.Info().Msgf("test in %s using %s", wd, conf)

	return client, &_logger, func(t2 *testing.T) {}
}

func TestRAC_Cloud_Idp_Jwt(t *testing.T) {
	client, logger, _ := setupTestSuite("../../test/qcs/psdemo_idp_jwt.yaml", t)

	req, res := client.NewRequest(http.MethodGet, "/users", nil)
	if res != nil {
		t.Error(res)
		return
	}

	resp, buf, res := client.DoRequest(req)
	if res != nil {
		t.Error(res)
		return
	}

	logger.Info().Msgf("resp: %s, %s", resp.Status, string(buf))
}
