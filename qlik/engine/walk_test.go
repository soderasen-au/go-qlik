package engine

import (
	"testing"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/crypto"
	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-common/util"
)

func TestWalkApp_1(t *testing.T) {
	logger, err := loggers.GetLogger("../../test/log.json")
	if err != nil {
		t.Errorf("%s error: %s", "GetLogger", err.Error())
	}

	cfg := Config{
		EngineURI:     "wss://soderasen-au-qs:4747/app",
		AppID:         "7ec5decd-c6ca-432c-a1dc-9703ebd873a7",
		UserName:      "Administrator",
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

	doc, err := conn.Global.OpenDoc(ConnCtx, "7ec5decd-c6ca-432c-a1dc-9703ebd873a7", "", "", "", false)
	if err != nil {
		t.Errorf("%s error: %s", "OpenDoc", err.Error())
	}

	objWalker := func(doc *enigma.Doc, listName string, item NxContainerEntry, _logger *zerolog.Logger) (*ObjectSnapshot, *util.Result) {
		_logger.Info().Msgf(" - snapshot[%s] object[%s/%s]:", listName, item.Info.Type, item.Info.Id)
		shot := ObjectSnapshot{
			Info:        *item.Info,
			Parent:      nil,
			Title:       nil,
			Description: nil,
			Properties:  nil,
			Digest:      "",
		}
		return &shot, nil
	}

	WalkApp(doc, objWalker, logger)
}