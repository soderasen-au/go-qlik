package qlik

import (
	"fmt"
	"testing"

	"github.com/soderasen-au/go-common/util"
)

func TestParseUser(t *testing.T) {
	user, res := ParseUser("soderasen-au-qs\\sa")
	if res != nil {
		t.Error(res)
	}
	fmt.Printf("user: %s\n", util.JsonStr(user))
}
