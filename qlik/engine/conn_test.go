package engine

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"
	"gopkg.in/yaml.v3"

	"github.com/soderasen-au/go-qlik/qlik/rac"
)

func TestNewConn(t *testing.T) {
	conn, logger, _ := setupTestSuiteConn("../../test/engine/soderasen-au-qs.yaml", t)
	if conn == nil {
		logger.Fatal().Msgf("conn is nil")
		return
	}
	defer conn.Global.DisconnectFromServer()

	ver, err := conn.Global.EngineVersion(ConnCtx)
	if err != nil {
		logger.Fatal().Msgf("engine version error: %v", err)
		t.Errorf("error %v", err)
		return
	}
	logger.Info().Msgf("engine version: %v", ver)

	_, err = conn.Global.OpenDoc(ConnCtx, "2dea34ed-10db-4463-991c-3f8ba6176476", "", "", "", false)
	if err != nil {
		t.Errorf("%s error: %s", "OpenDoc", err.Error())
	}
}

func setupRACClient(conf string, t *testing.T) (*rac.RestApiClient, *zerolog.Logger, func(t2 *testing.T)) {
	wd, _ := os.Getwd()
	buf, err := os.ReadFile(conf)
	if err != nil {
		t.Errorf("can't load config file: %s", err.Error())
		return nil, nil, nil
	}

	var cfg rac.Config
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		t.Errorf("can't parse config file: %s", err.Error())
		return nil, nil, nil
	}

	client, res := rac.New(cfg)
	if res != nil {
		t.Errorf("can't parse config file: %s", res.Error())
		return nil, nil, nil
	}
	_logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}.Out).Level(zerolog.DebugLevel)
	client.Logger = &_logger
	_logger.Info().Msgf("test in %s using %s", wd, conf)

	return client, &_logger, func(t2 *testing.T) {}
}

func TestNewConnFromRAC(t *testing.T) {
	rac, logger, _ := setupRACClient("../../test/qcs/psdemo_idp_jwt.yaml", t)
	conn, res := NewConnFromRAC(rac, "3eb6de59-d7d4-4c92-9908-db9009790ba3") //happiness
	if res != nil {
		t.Errorf("NewConnFromRAC: %s", res.Error())
		return
	}
	defer conn.Global.DisconnectFromServer()

	ver, err := conn.Global.EngineVersion(ConnCtx)
	if err != nil {
		t.Errorf("EngineVersion: %s", err.Error())
		return
	}
	logger.Info().Msgf("%v", *ver)

	doc, err := conn.Global.OpenDoc(ConnCtx, "3eb6de59-d7d4-4c92-9908-db9009790ba3", "", "", "", false)
	if err != nil {
		t.Errorf("OpenDoc: %s", err.Error())
		return
	}

	layout, res := GetSessionObjectLayout(doc)
	if res != nil {
		t.Errorf("GetSessionObjectLayout: %s", res.Error())
		return
	}
	fmt.Printf("layout: %v", util.JsonStr(layout))
}
