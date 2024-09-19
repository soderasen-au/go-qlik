package engine

import (
	"testing"
)

func TestWalkApp_1(t *testing.T) {
	//logger, err := loggers.GetLogger("../../test/log.json")
	//if err != nil {
	//    t.Errorf("%s error: %s", "GetLogger", err.Error())
	//}
	//
	//cfg := Config{
	//    EngineURI:     "wss://soderasen-au-qs:4747/app",
	//    AppID:         "7ec5decd-c6ca-432c-a1dc-9703ebd873a7",
	//    UserName:      "Administrator",
	//    UserDirectory: "soderasen-au-qs",
	//    AuthMode:      "cert",
	//    ServerType:    "on_prem",
	//    Certs: crypto.Certificates{
	//        ClientFile:    "\\\\SODERASEN-AU-PC\\certs\\soderasen-au.com\\qlik\\client.pem",
	//        ClientkeyFile: "\\\\SODERASEN-AU-PC\\certs\\soderasen-au.com\\qlik\\client_key.pem",
	//        CAFile:        "\\\\SODERASEN-AU-PC\\certs\\soderasen-au.com\\qlik\\root.pem",
	//    },
	//    RandomProxySession: true,
	//}
	//
	//conn, err := NewConn(cfg)
	//if err != nil {
	//    t.Errorf("%s error: %s", "NewConn", err.Error())
	//}
	//defer conn.Global.DisconnectFromServer()
	//
	//doc, err := conn.Global.OpenDoc(ConnCtx, "7ec5decd-c6ca-432c-a1dc-9703ebd873a7", "", "", "", false)
	//if err != nil {
	//    t.Errorf("%s error: %s", "OpenDoc", err.Error())
	//}
	//
	//walkers := make(ListWalkFuncMap[ObjectSnapshot])
	//walkers[SHEET_LIST] = NewRecurObjWalkFunc(func(e ObjWalkEntry) (*ObjWalkResult[ObjectSnapshot], *util.Result) {
	//    e.Logger.Info().Msgf(" - walk object[%s/%s]:", e.Info.Type, e.Info.Id)
	//    shot := ObjWalkResult[ObjectSnapshot]{
	//        Info:   e.Info,
	//        Parent: e.Parent,
	//    }
	//    return &shot, nil
	//})
	//walkers[ANY_LIST] = func(e ObjWalkEntry) (*ObjWalkResult[ObjectSnapshot], *util.Result) {
	//    e.Logger.Info().Msgf(" - walk any object[%s/%s]:", e.Item.Info.Type, e.Item.Info.Id)
	//    shot := ObjWalkResult[ObjectSnapshot]{
	//        Info:   e.Info,
	//        Parent: e.Parent,
	//    }
	//    return &shot, nil
	//}
	//
	//mc := MixedConfig{
	//    AppId:         "",
	//    OnPrem:        &cfg,
	//    OnPremCluster: nil,
	//    QCS:           nil,
	//}
	//WalkApp(doc, mc, nil, walkers, logger)
}
