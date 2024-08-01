package qrs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/soderasen-au/go-common/util"
	"gopkg.in/yaml.v3"

	"github.com/rs/zerolog"
)

var (
	happinessAppID string = "7ec5decd-c6ca-432c-a1dc-9703ebd873a7"
)

func setupTestSuite(conf string, t *testing.T) (*Client, *zerolog.Logger, func(t2 *testing.T)) {
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

	client, res := NewClient(cfg)
	if res != nil {
		t.Errorf("can't parse config file: %s", res.Error())
		return nil, nil, nil
	}
	_logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}.Out).Level(zerolog.TraceLevel)
	client.SetLogger(&_logger)
	_logger.Info().Msgf("test in %s using %s", wd, conf)

	return client, &_logger, func(t2 *testing.T) {}
}

func TestClient_GetAppList(t *testing.T) {
	client, logger, tearDown := setupTestSuite("../../../test/qrs/soderasen-au.com.yaml", t)
	if t.Failed() {
		return
	}
	defer tearDown(t)

	apps, res := client.GetAppList()
	if res != nil {
		logger.Error().Msg(res.Error())
		t.Errorf("failed: %s", res.Error())
		return
	}

	for _, app := range apps {
		logger.Info().Msgf("[%s]: `%s`, owned by `%s\\%s`", app.ID, app.Name, app.Owner.UserDirectory, app.Owner.UserID)
	}
	logger.Info().Msgf("%d in total", len(apps))
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

func TestClient_Copy(t *testing.T) {
	client, logger, tearDown := setupTestSuite("../../../test/qrs/soderasen-au.com.yaml", t)
	if t.Failed() {
		return
	}
	defer tearDown(t)

	app, res := client.Copy(happinessAppID, "new_happiness")
	if res != nil {
		logger.Error().Msg(res.Error())
		t.Errorf("failed: %s", res.Error())
		return
	}

	logger.Info().Msgf("duplicated app [%s]: `%s`, owned by `%s\\%s`", app.ID, app.Name, app.Owner.UserDirectory, app.Owner.UserID)
}

func TestClient_GetAppHubList(t *testing.T) {
	client, logger, tearDown := setupTestSuite("../../../test/qrs/soderasen-au.com.yaml", t)
	if t.Failed() {
		return
	}
	defer tearDown(t)

	apps, res := client.GetAppHubList()
	if res != nil {
		logger.Error().Msg(res.Error())
		t.Errorf("failed: %s", res.Error())
		return
	}

	fmt.Println(util.JsonStr(apps))
}
