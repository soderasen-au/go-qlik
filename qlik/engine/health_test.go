package engine

import (
	"fmt"
	"testing"

	"github.com/soderasen-au/go-common/util"
)

func TestHttpClient_GetHealthInfo(t *testing.T) {
	client, logger, tearDown := setupTestSuiteHttpClient("../../test/engine/soderasen-au-qs.yaml", t)
	if t.Failed() {
		return
	}
	defer tearDown(t)

	health, res := client.GetHealthInfo()
	if res != nil {
		logger.Error().Msgf("GetHealthInfo: %s", res.Error())
	}

	fmt.Println(util.JsonStr(health))
}
