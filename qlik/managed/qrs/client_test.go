package qrs

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"testing"
)

func TestClient_GetOpenAPISpec(t *testing.T) {
	client, _, tearDown := setupTestSuite("../../../test/qrs/localhost.yaml", t)
	defer tearDown(t)
	t.Run("GetHubApps", func(t *testing.T) {
		openapi, res := client.Get("/about/openapi/main", nil)
		if res != nil {
			t.Errorf("GetHubApps failed: %s", res.Error())
		}
		var out bytes.Buffer
		_ = json.Indent(&out, openapi, "", "  ")
		_ = os.WriteFile("../../../test/qrs/openapi.json", out.Bytes(), fs.ModePerm)
	})
}
