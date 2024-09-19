package qrs

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/soderasen-au/go-common/util"
)

func TestClient_SelectApp(t *testing.T) {
	client, logger, tearDown := setupTestSuite("../../../test/qrs/localhost.yaml", t)
	defer tearDown(t)

	sel, res := client.SelectApp(happinessAppID)
	if res != nil {
		t.Errorf("failed: %s", res.Error())
	}

	sync, res := client.GetAppSynthetic(*sel.Id)
	if res != nil {
		t.Errorf("failed to GetAppSynthetic : %s", res.Error())
	}

	cpValue := CustomPropertySyntheticValue{
		Removed: []string{"value1"},
		Added:   []string{},
	}
	cpValueMessage, err := json.Marshal(&cpValue)
	if err != nil {
		t.Errorf("can't marshal cpvalue: %s", err.Error())
	}

	sz := len(sync.Properties)
	sync.Properties[sz-2].Value = cpValueMessage
	*sync.Properties[sz-2].ValueIsModified = true
	now := time.Now()
	sync.LatestModifiedDate = &now

	res = client.UpdateAppSynthetic(*sel.Id, *sync)
	if res != nil {
		t.Errorf("failed to UpdateAppSynthetic : %s", res.Error())
	}

	defer func() {
		res = client.DeleteSelection(*sel.Id)
		if res != nil {
			logger.Error().Msg(res.Error())
		}
	}()

	logger.Info().Msg(string(util.Jsonify(&sync)))
}
