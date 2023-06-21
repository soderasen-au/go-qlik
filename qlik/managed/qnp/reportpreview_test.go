package qnp

import (
	"fmt"
	"testing"

	"github.com/Click-CI/common/crypto"
	"github.com/Click-CI/common/util"
	"github.com/soderasen-au/go-qlik/qlik"
)

func TestClient_GetReportPreviewList(t *testing.T) {
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

	reportPreviews, res := client.GetReportPreviewList()
	if res != nil {
		t.Errorf("Get: %s", res.Error())
		return
	}

	for i, r := range reportPreviews {
		data, res := client.GetReportPreviewThumbnail(r.ID, "1", "70", "100")
		if res != nil {
			fmt.Printf("report[%d]: %s thumbnail failed: %s\n", i, r.ID, res.Error())
		} else {
			fmt.Printf("report[%d]:%s thumbnail len: %d\n", i, r.ID, len(data))
		}
	}

	fmt.Println(string(util.Jsonify(&reportPreviews)))
}
