package qps

import (
	"github.com/soderasen-au/go-common/util"
	"testing"
)

func TestClient_GetWebTicket(t *testing.T) {
	client, _logger, tearDown := setupTestSuite("../../../test/qps/localhost.yaml", true, t)
	defer tearDown(t)
	usrTicket := &User{
		UserId:  "soder",
		UserDir: ".",
	}
	res := client.GetWebTicket(usrTicket)
	if res != nil {
		t.Errorf("GetWebTicket: %s", res.Error())
		return
	}
	_logger.Info().Msgf("Ticket: %v", util.JsonStr(usrTicket))
}
