package rac

import (
	"crypto/rsa"

	"github.com/soderasen-au/go-common/crypto"
	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/qlik"
)

type AuthMethod string

const (
	AuthMethodCert      AuthMethod = "cert"
	AuthMethodJWT       AuthMethod = "jwt"
	AuthMethodAPIKey    AuthMethod = "api_key"
	AuthMethodWebTicket AuthMethod = "webticket"
)

type CloudJwtConfig struct {
	KeyId            string `json:"key_id,omitempty" yaml:"key_id,omitempty" bson:"key_id,omitempty"`
	Issuer           string `json:"issuer,omitempty" yaml:"issuer,omitempty" bson:"issuer,omitempty"`
	WebIntegrationID string `json:"web_integration_id,omitempty" yaml:"web_integration_id,omitempty" bson:"web_integration_id,omitempty"`
	UserName         string `json:"user_id,omitempty" yaml:"user_id,omitempty" bson:"user_name,omitempty"`
	UserEmail        string `json:"user_email,omitempty" yaml:"user_email,omitempty" bson:"user_email,omitempty"`
	UserSub          string `json:"user_sub,omitempty" yaml:"user_sub,omitempty" bson:"user_sub,omitempty"`

	//assgined when Cloud session is created on-the-fly, otherwise it's empty
	CsrfToken string `json:"csrf_token,omitempty" yaml:"csrf_token,omitempty" bson:"csrf_token,omitempty"`
}

func (c CloudJwtConfig) IsValid() bool {
	return c.KeyId != "" && c.Issuer != "" && c.UserSub != "" && c.UserName != "" && c.UserEmail != "" && c.WebIntegrationID != ""
}

func (c CloudJwtConfig) NewPayload() *qlik.CloudJwtClaim {
	return qlik.NewCloudJwtClaim(c.KeyId, c.Issuer, c.UserSub, c.UserName, c.UserEmail)
}

func (c CloudJwtConfig) GenerateJWT(key *rsa.PrivateKey) (string, *util.Result) {
	payload := c.NewPayload()
	jwt, res := payload.GetJWT(key)
	if res != nil {
		return "", res.With("GenerateManagedJWT")
	}
	return jwt, nil
}

type AuthConfig struct {
	Method   AuthMethod           `json:"method" yaml:"method" bson:"method"`
	Xrf      bool                 `json:"xrf" yaml:"xrf" bson:"xrf"`
	User     *qlik.User           `json:"user,omitempty" yaml:"user,omitempty" bson:"user,omitempty"`
	Certs    *crypto.Certificates `json:"certs,omitempty" yaml:"certs" bson:"certs,omitempty"`
	Token    *string              `json:"token,omitempty" yaml:"token,omitempty" bson:"token,omitempty"`
	CloudJwt *CloudJwtConfig      `json:"cloud_jwt,omitempty" yaml:"cloud_jwt,omitempty" bson:"cloud_jwt,omitempty"`
}
