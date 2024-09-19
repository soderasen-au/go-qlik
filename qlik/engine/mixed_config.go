package engine

import (
	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/qlik/rac"
)

type MixedConfig struct {
	AppId         string      `json:"app_id" yaml:"app_id" bson:"app_id"`
	OnPrem        *Config     `json:"on_prem,omitempty" yaml:"on_prem,omitempty" bson:"on_prem,omitempty"`
	OnPremCluster *Cluster    `json:"on_prem_cluster,omitempty" yaml:"on_prem_cluster,omitempty" bson:"on_prem_cluster,omitempty"`
	QCS           *rac.Config `json:"qcs,omitempty" yaml:"qcs,omitempty" bson:"qcs,omitempty"`
}

func (mc MixedConfig) Connect() (*Conn, *util.Result) {
	if mc.OnPrem != nil {
		cfg := *mc.OnPrem
		cfg.AppID = mc.AppId
		conn, err := NewConn(cfg)
		if err != nil {
			return nil, util.Error("OnPrem.NewConn", err)
		}
		return conn, nil
	} else if mc.OnPremCluster != nil {
		return NewConnFromCluster(mc.OnPremCluster, mc.AppId)
	} else if mc.QCS != nil {
		rac, res := rac.New(*mc.QCS)
		if res != nil {
			return nil, res.With("rac.NewClient")
		}
		return NewConnFromRAC(rac, mc.AppId)
	}
	return nil, util.MsgError("Dispatch", "empty engine config")
}
