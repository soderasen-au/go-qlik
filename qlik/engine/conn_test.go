package engine

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/crypto"
	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/qlik/rac"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
	"time"
)

func TestNewConn(t *testing.T) {
	//logger, err := loggers.GetLogger("../../test/log.json")
	//if err != nil {
	//	t.Errorf("%s error: %s", "GetLogger", err.Error())
	//}

	cfg := Config{
		EngineURI:     "wss://soderasen-au-qs:4747/app",
		AppID:         "7ec5decd-c6ca-432c-a1dc-9703ebd873a7",
		UserName:      "li",
		UserDirectory: "soderasen-au-qs",
		AuthMode:      "cert",
		ServerType:    "on_prem",
		Certs: crypto.Certificates{
			ClientFile:    "\\\\SODERASEN-AU-PC\\certs\\soderasen-au.com\\qlik\\client.pem",
			ClientkeyFile: "\\\\SODERASEN-AU-PC\\certs\\soderasen-au.com\\qlik\\client_key.pem",
			CAFile:        "\\\\SODERASEN-AU-PC\\certs\\soderasen-au.com\\qlik\\root.pem",
		},
		RandomProxySession: true,
	}

	conn, err := NewConn(cfg)
	if err != nil {
		t.Errorf("%s error: %s", "NewConn", err.Error())
	}
	defer conn.Global.DisconnectFromServer()

	doc, err := conn.Global.OpenDoc(ConnCtx, "071f466e-c592-4f42-bd1b-b773329bb0e0", "", "", "", false)
	if err != nil {
		t.Errorf("%s error: %s", "OpenDoc", err.Error())
	}

	obj, err := doc.GetObject(ConnCtx, "yDTf")
	if err != nil {
		t.Errorf("%s error: %s", "GetObject", err.Error())
	}

	layout, err := obj.GetLayout(ConnCtx)
	if err != nil {
		t.Errorf("%s error: %s", "GetLayout", err.Error())
	}
	if layout.HyperCube != nil {
		fmt.Printf("sz: cols: %d, rows: %d\n", layout.HyperCube.Size.Cx, layout.HyperCube.Size.Cy)
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
