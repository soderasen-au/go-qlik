package qnp

import (
	"github.com/Click-CI/common/crypto"
	"github.com/Click-CI/common/util"
	"github.com/soderasen-au/go-qlik/qlik"
)

type Config struct {
	BaseURI      string               `json:"base_uri,omitempty" yaml:"base_uri,omitempty" bson:"base_uri,omitempty"`
	NewsStandURI string               `json:"news_stand_uri,omitempty" yaml:"news_stand_uri,omitempty" bson:"news_stand_uri,omitempty"`
	User         *qlik.JwtClaim       `json:"user,omitempty" yaml:"user,omitempty" bson:"user,omitempty"`
	KeyPair      *crypto.KeyPairFiles `json:"key_pair,omitempty" yaml:"key_pair,omitempty" bson:"key_pair,omitempty"`
	RsaKeyPair   *crypto.RsaKeyPair
}

func (c *Config) ParseKeyPair() *util.Result {
	if c.KeyPair == nil {
		return util.MsgError("GetKeyPairFiles", "No Key pair files")
	}
	kp, res := crypto.NewRsaKeyPair(*c.KeyPair)
	if res != nil {
		return res.With("NewRsaKeyPair")
	}
	c.RsaKeyPair = kp
	return nil
}
