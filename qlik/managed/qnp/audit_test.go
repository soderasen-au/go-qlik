package qnp

import (
	"fmt"
	"testing"

	"github.com/soderasen-au/go-common/crypto"

	"github.com/soderasen-au/go-qlik/qlik"
)

func TestClient_GetAuditLogs(t *testing.T) {
	email := "bl@s-cubed.dk"
	config := Config{
		BaseURI:      "https://nprintsand.s-cubed.local:4993",
		NewsStandURI: "https://nprintsand.s-cubed.local:4994",
		User: &qlik.JwtClaim{
			Email: &email,
		},
		KeyPair: &crypto.KeyPairFiles{
			Key:  "../test/qnp/key.pem",
			Cert: "../test/qnp/cert.pem",
		},
	}
	client, res := NewClient(config)
	if res != nil {
		t.Errorf("NewClient: %s", res.Error())
		return
	}

	//cl := zerolog.New(loggers.DefaultConsoleWriter)
	//client.Logger = &cl

	auditLogs, res := client.GetAuditLogRecords()
	if res != nil {
		t.Errorf("GetAuditLogs: %s", res.Error())
		return
	}
	fmt.Println("get ", len(auditLogs))
}
