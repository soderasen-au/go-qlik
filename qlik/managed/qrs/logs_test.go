package qrs

import (
	"fmt"
	"testing"
)

func Test_ReadTraceLogFile(t *testing.T) {
	entries, res := ReadTraceLogFile("../../../test\\qrs\\Audit_Repository.txt")
	if res != nil {
		t.Errorf("ReadTraceLogFile res: %s", res.Error())
		return
	}
	fmt.Printf("ReadTraceLogFile: %d\n", len(entries))
}
