package qrs

import (
	"testing"
)

func TestClient_GetHubApps(t *testing.T) {
	qrs, logger, tearDown := setupTestSuite("../../../test/qrs/localhost.yaml", t)
	defer tearDown(t)
	t.Run("GetHubApps", func(t *testing.T) {
		apps, res := qrs.GetHubApps()
		if res != nil {
			t.Errorf("GetHubApps failed: %s", res.Error())
		}
		logger.Info().Msgf("got %d apps for %s\\%s", len(apps), qrs.Cfg.Auth.User.Directory, qrs.Cfg.Auth.User.Id)
		for ai, app := range apps {
			stream := ""
			if app.Stream != nil {
				stream = app.Stream.Name
			}
			logger.Trace().Msgf("app[%d]: %s (%s)", ai, app.Name, stream)
		}
	})
}
