package qcs

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"

	"github.com/soderasen-au/go-qlik/qlik/engine"
	"github.com/soderasen-au/go-qlik/qlik/rac"
)

func setupTestSuite(conf string, t *testing.T) (*Client, *zerolog.Logger, func(t2 *testing.T)) {
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

	client, res := NewClient(cfg)
	if res != nil {
		t.Errorf("can't parse config file: %s", res.Error())
	}
	_logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}.Out).Level(zerolog.DebugLevel)
	client.SetLogger(&_logger)
	_logger.Info().Msgf("test in %s using %s", wd, conf)

	return client, &_logger, func(t2 *testing.T) {}
}

func TestClient_NewEngineConn(t *testing.T) {
	client, logger, tearDown := setupTestSuite("../../test/qcs/psdemo.yaml", t)
	defer tearDown(t)

	appId := "3eb6de59-d7d4-4c92-9908-db9009790ba3"
	conn, res := client.NewEngineConn(appId)
	if res != nil {
		logger.Error().Msgf("NewEngineConn: %s", res.Error())
		t.Errorf("failed: %s", res.Error())
		return
	}
	defer conn.Global.DisconnectFromServer()

	user, _ := conn.Global.GetAuthenticatedUser(engine.ConnCtx)
	logger.Info().Msgf("current user %s.\n", user)
	_, err := conn.Global.OpenDoc(engine.ConnCtx, appId, "", "", "", false)
	if err != nil {
		logger.Error().Msgf("OpenDoc: %s", err.Error())
		t.Errorf("failed: %s", err.Error())
		return
	}
}
