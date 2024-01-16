package engine

import (
	"fmt"
	"github.com/soderasen-au/go-common/crypto"
	"testing"
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
