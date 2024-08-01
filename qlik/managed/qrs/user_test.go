package qrs

import (
	"testing"

	"github.com/soderasen-au/go-common/util"
)

func TestClient_GetUserId(t *testing.T) {
	client, logger, tearDown := setupTestSuite("../../../test/qrs/localhost.yaml", t)
	if t.Failed() {
		return
	}
	defer tearDown(t)

	user, res := client.GetUserByDomainName("jzs-thinkpad", "soder")
	if res != nil {
		logger.Error().Msg(res.Error())
		t.Errorf("failed: %s", res.Error())
		return
	}
	logger.Info().RawJSON("user", util.Jsonify(user)).Msg("end")
}
