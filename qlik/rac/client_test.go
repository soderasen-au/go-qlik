package rac

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestRAC_Cloud_ApiKey(t *testing.T) {
	buf, err := ioutil.ReadFile("../../test/qcs/apikey.json")
	if err != nil {
		t.Error(err)
		return
	}
	var cfg Config
	err = json.Unmarshal(buf, &cfg)
	if err != nil {
		t.Error(err)
		return
	}

	client, res := New(cfg)
	if res != nil {
		t.Error(res)
		return
	}

	req, res := client.NewRequest(http.MethodGet, "/users", nil)
	if res != nil {
		t.Error(res)
		return
	}
	_, buf, res = client.DoRequest(req)
	if res != nil {
		t.Error(res)
		return
	}
}
