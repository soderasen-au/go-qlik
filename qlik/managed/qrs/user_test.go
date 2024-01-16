package qrs

import (
	"github.com/soderasen-au/go-common/util"
	"testing"
)

func TestClient_GetUserId(t *testing.T) {
	client, logger, tearDown := setupTestSuite("../../../test/qrs/soderasen-au-qs.yaml", t)
	if t.Failed() {
		return
	}
	defer tearDown(t)

	user, res := client.GetUserByName("soderasen-au-qs\\sa")
	if res != nil {
		logger.Error().Msg(res.Error())
		t.Errorf("failed: %s", res.Error())
		return
	}

	logger.Info().Msgf("user: %s", util.JsonStr(user))
}
