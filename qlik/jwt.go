package qlik

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/soderasen-au/go-common/util"
)

type CloudJwtClaim struct {
	keyId         string `json:"-" yaml:"-" bson:"-"`
	SubType       string `json:"subType" yaml:"subType" bson:"subType"`
	Name          string `json:"name" yaml:"name" bson:"name"`
	Email         string `json:"email" yaml:"email" bson:"email"`
	EmailVerified string `json:"email_verified" yaml:"email_verified" bson:"email_verified"`
	jwt.RegisteredClaims
}

func NewCloudJwtClaim(keyid, issuer, sub, name, email string) *CloudJwtClaim {
	ret := CloudJwtClaim{
		keyId:         keyid,
		SubType:       "user",
		Name:          name,
		Email:         email,
		EmailVerified: "true",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   sub,
			Audience:  jwt.ClaimStrings{"qlik.api/login/jwt-session"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Minute)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
	}
	return &ret
}

func (p *CloudJwtClaim) GetJWT(privateKey *rsa.PrivateKey) (string, *util.Result) {
	jwt.MarshalSingleStringAsArray = false
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, *p)
	token.Header["kid"] = p.keyId

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", util.Error("SignedString", err)
	}

	return tokenString, nil
}

// json keys are for configuration
// actually payload is still using userDirectory + userId
type JwtClaim struct {
	Name          *string `json:"name,omitempty" yaml:"name,omitempty" bson:"name,omitempty"`
	UserID        *string `json:"userId,omitempty" yaml:"userId,omitempty" bson:"userId,omitempty"`
	UserDirectory *string `json:"userDirectory,omitempty" yaml:"userDirectory,omitempty" bson:"userDirectory,omitempty"`
	Email         *string `json:"email,omitempty" yaml:"email,omitempty" bson:"email,omitempty"`
	jwt.RegisteredClaims
}

func (t *JwtClaim) GetQlikClaims() *jwt.MapClaims {
	claim := jwt.MapClaims{}
	if t.Name != nil {
		claim["Name"] = *t.Name
	}
	if t.UserID != nil {
		claim["userId"] = *t.UserID
	}
	if t.UserDirectory != nil {
		claim["userDirectory"] = *t.UserDirectory
	}
	if t.Email != nil {
		claim["email"] = *t.Email
	}
	return &claim
}

func (t *JwtClaim) GetJWT(privateKey *rsa.PrivateKey) (string, *util.Result) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, t.GetQlikClaims())

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", util.Error("SignedString", err)
	}

	return tokenString, nil
}

func (t *JwtClaim) GetHeader(privateKey *rsa.PrivateKey) (map[string]string, *util.Result) {
	jwt, res := t.GetJWT(privateKey)
	if res != nil {
		return nil, res.With("GetJWT")
	}

	return map[string]string{"Authorization": fmt.Sprintf("Bearer %s", jwt)}, nil
}
