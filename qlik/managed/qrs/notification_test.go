package qrs

import (
	"github.com/Click-CI/common/util"
	"testing"
)

func TestClient_Subscribe(t *testing.T) {
	client, logger, tearDown := setupTestSuite("../../../test/qrs/localhost.yaml", t)
	defer tearDown(t)

	subs := []Subscription{
		{TypeName: util.Ptr("app"), CallbackURL: "\"http://192.168.50.113:8080/hub/subscription/0\""},
		{TypeName: util.Ptr("user"), CallbackURL: "\"http://192.168.50.113:8080/hub/subscription/1\""},
		{TypeName: util.Ptr("appobject"), CallbackURL: "\"http://192.168.50.113:8080/hub/subscription/2\""},
		{TypeName: util.Ptr("stream"), CallbackURL: "\"http://192.168.50.113:8080/hub/subscription/3\""},
		{TypeName: util.Ptr("dataconnection"), CallbackURL: "\"http://192.168.50.113:8080/hub/subscription/4\""},
	}
	t.Run("Subscribe", func(t *testing.T) {
		for i, sub := range subs {
			subId, res := client.Subscribe(sub)
			if res != nil {
				t.Errorf("Subscribe failed: %s", res.Error())
			}
			logger.Trace().Msgf("sub[%d] Id: %s\n", i, subId)
		}
	})
}
