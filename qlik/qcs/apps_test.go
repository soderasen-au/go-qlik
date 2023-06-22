package qcs

import (
	"testing"
)

func TestClient_Export(t *testing.T) {
	client, logger, tearDown := setupTestSuite("../../test/qcs/psdemo.yaml", t)
	defer tearDown(t)

	appId := "3eb6de59-d7d4-4c92-9908-db9009790ba3"
	_, res := client.Export(appId, "../../test/qcs/", false)
	if res != nil {
		logger.Error().Msgf("Export: %s", res.Error())
		t.Errorf("failed: %s", res.Error())
		return
	}
}
