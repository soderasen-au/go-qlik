package qlik

import (
	"fmt"
	"strings"

	"github.com/soderasen-au/go-common/util"
)

type User struct {
	Id        string `json:"userId" yaml:"userId"`
	Directory string `json:"userDirectory" yaml:"userDirectory"`
}

func (u User) DomainId() string {
	return fmt.Sprintf("%s\\%s", u.Directory, u.Id)
}

func ParseUser(name string) (*User, *util.Result) {
	parts := strings.Split(strings.TrimSpace(name), "\\")
	if len(parts) != 2 {
		return nil, util.MsgError("split", fmt.Sprintf("get %d parts", len(parts)))
	}

	return &User{Directory: parts[0], Id: parts[1]}, nil
}
