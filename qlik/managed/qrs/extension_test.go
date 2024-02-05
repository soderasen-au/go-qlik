package qrs

import (
	"io"
	"os"
	"testing"
)

func TestClient_ImportExtension(t *testing.T) {
	client, logger, tearDown := setupTestSuite("../../../test/qrs/soderasen-au-qs.yaml", t)
	if t.Failed() {
		return
	}
	defer tearDown(t)

	extFile, err := os.Open("../../../test/qrs/qlik-alert-extension.zip")
	if err != nil {
		t.Errorf("open ext file failed: %s", err.Error())
		return
	}
	defer extFile.Close()

	fileContents, err := io.ReadAll(extFile)
	if err != nil {
		t.Errorf("io.ReadAll: %s", err.Error())
		return
	}

	ext, res := client.ImportExtension(fileContents)
	if res != nil {
		logger.Error().Msg(res.Error())
		t.Errorf("failed: %s", res.Error())
		return
	}

	logger.Info().Msgf("extension replaced: %v`", ext)
}
