package qrs

import (
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
	"time"
)

var (
	happinessAppID string = "08df6142-ab4e-4e91-88e8-e32682d3bfcd"
)

func setupTestSuite(conf string, t *testing.T) (*Client, *zerolog.Logger, func(t2 *testing.T)) {
	wd, _ := os.Getwd()
	buf, err := os.ReadFile(conf)
	if err != nil {
		t.Errorf("can't load config file: %s", err.Error())
	}

	var cfg Config
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

func TestClient_GetAppList(t *testing.T) {
	client, logger, tearDown := setupTestSuite("../../../test/qrs/localhost.yaml", t)
	defer tearDown(t)
	_, res := client.GetAppList()
	if res != nil {
		logger.Error().Msg(res.Error())
		t.Errorf("failed: %s", res.Error())
		return
	}

	//newApp := AppPtr{
	//	CustomProperties: []CustomPropertyValue{
	//		{
	//			Definition: CustomPropertyDefinitionCondensed{
	//				ID: "faf04d21-feb1-41c6-b9c6-a263ec84601d",
	//			},
	//			Value: "value2",
	//		},
	//	},
	//}
	//newApp.CustomProperties = append(newApp.CustomProperties)
	//
	//resApp, res := client.UpdateApp(app.ID, &newApp)
	//if res != nil {
	//	t.Errorf("failed UpdateApp: %s", res.Error())
	//	return
	//}
}
