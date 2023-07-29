package qps

import (
	"encoding/json"

	"github.com/soderasen-au/go-common/util"
)

type User struct {
	UserId  string  `json:"UserId" yaml:"UserId"`
	UserDir string  `json:"UserDirectory" yaml:"UserDirectory"`
	Ticket  *string `json:"Ticket" yaml:"Ticket"`
}

func (qpsClient *Client) GetWebTicket(qpsUser *User) *util.Result {
	resp, res := qpsClient.Post("/ticket", nil, qpsUser)
	if res != nil {
		errRes := res.With("AddTicket")
		qpsClient.Logger().Err(errRes).Msg("AddTicket")
		return errRes
	}

	err := json.Unmarshal(resp, qpsUser)
	if err != nil {
		errRes := util.Error("parse ticket response", err)
		qpsClient.Logger().Err(errRes).Msg("Unmarshal")
		return errRes
	}

	if qpsUser.Ticket == nil {
		errRes := util.MsgError("parse ticket response", "empty ticket")
		qpsClient.Logger().Err(errRes).Msg("Unmarshal")
		return errRes
	}
	return nil
}
