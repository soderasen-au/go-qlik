package qnp

import (
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/crypto"
	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-qlik/qlik"
)

func TestClient_Get(t *testing.T) {
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
	}
	cl := zerolog.New(loggers.DefaultConsoleWriter)
	client.Logger = &cl
	resp, res := client.Get("/reports", nil)
	if res != nil {
		t.Errorf("Get: %s", res.Error())
	}
	fmt.Println(string(resp))
}
