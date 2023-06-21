package engine

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/qlik-oss/enigma-go/v3"
	"github.com/soderasen-au/go-qlik/qlik"
)

var (
	// ConnCtx is used for Engine requests per session/connection;
	ConnCtx context.Context
)

// Conn creates wss connection to Engine
type Conn struct {
	Global *enigma.Global `json:"-"`
	Cfg    Config         `json:"config"`
}

func newCertConn(cfg Config) (con *Conn, err error) {
	headers := make(http.Header, 1)
	headers.Set("X-Qlik-User", fmt.Sprintf("UserDirectory=%s; UserId=%s", cfg.UserDirectory, cfg.UserName))
	tlsConfig, err := cfg.Certs.NewTlsConfig()
	if err != nil {
		return nil, err
	}
	global, err := enigma.Dialer{TLSClientConfig: tlsConfig}.Dial(ConnCtx, cfg.EngineURI, headers)
	if err != nil {
		return nil, errors.New(fmt.Sprintln("Could not connect", err))
	}
	conn := Conn{
		Global: global,
		Cfg:    cfg,
	}
	return &conn, nil
}

func newJwtConn(cfg Config) (con *Conn, err error) {
	if cfg.JWT == "" {
		jwtPayload := qlik.JwtClaim{
			UserID:        &cfg.UserName,
			UserDirectory: &cfg.UserDirectory,
		}
		keypair, res := cfg.Certs.NewRsaKeyPair()
		if res != nil {
			return nil, res.With("NewRsaKeyPair")
		}
		cfg.JWT, res = jwtPayload.GetJWT(keypair.Key)
		if res != nil {
			return nil, res.With("GetJWT")
		}
	}
	headers := make(http.Header, 1)
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.JWT))

	global, err := enigma.Dialer{}.Dial(ConnCtx, cfg.EngineURI, headers)
	if err != nil {
		return nil, errors.New(fmt.Sprintln("Could not connect", err))
	}
	conn := Conn{
		Global: global,
		Cfg:    cfg,
	}
	return &conn, nil
}

func newDesktopConn(cfg Config) (con *Conn, err error) {
	global, err := enigma.Dialer{}.Dial(ConnCtx, cfg.EngineURI, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintln("Could not connect", err))
	}
	conn := Conn{
		Global: global,
		Cfg:    cfg,
	}
	return &conn, nil
}

// NewConn creates a Engine wss according to cfg.
func NewConn(cfg Config) (con *Conn, err error) {
	res := cfg.QCSEngineURIAppendAppID(cfg.AppID)
	if res != nil {
		return nil, res
	}

	switch authMode := strings.ToLower(string(cfg.AuthMode)); authMode {
	case "cert":
		return newCertConn(cfg)
	case "jwt":
		return newJwtConn(cfg)
	case "desktop":
		return newDesktopConn(cfg)
	}

	return nil, errors.New("invalid auth mode in config")
}

func init() {
	ConnCtx = context.Background()
}
