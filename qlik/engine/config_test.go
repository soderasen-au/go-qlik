package engine

import (
	"testing"
)

func TestConfig_GetAppUrl(t *testing.T) {
	cfg := Config{EngineURI: "wss://soderasen-au.com/jwt/app/7ec5decd-c6ca-432c-a1dc-9703ebd873a7"}
	appUrl, res := cfg.GetAppUrl()
	if res != nil {
		t.Errorf("failed: %s", res.Error())
	} else if appUrl != "https://soderasen-au.com/jwt/app/7ec5decd-c6ca-432c-a1dc-9703ebd873a7" {
		t.Errorf("got: %s", appUrl)
	}
}
