package client

import (
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/qlik/config"
	"github.com/soderasen-au/go-qlik/qlik/managed/qnp"
	"github.com/soderasen-au/go-qlik/qlik/managed/qps"
	"github.com/soderasen-au/go-qlik/qlik/managed/qrs"
	"github.com/soderasen-au/go-qlik/qlik/rac"
)

type Managed struct {
	Config config.Config

	//Engine engine.Conn
	NP  *qnp.Client
	QPS *qps.Client
	QRS *qrs.Client
	HUB *rac.RestApiClient
}

func NewManaged(config config.Config, l *zerolog.Logger) (c *Managed, res *util.Result) {
	c = &Managed{
		Config: config,
	}

	if config.NPrinting != nil {
		c.NP, res = qnp.NewClient(*config.NPrinting)
		if res != nil {
			return nil, res.With("qnp.NewClient")
		}
		c.NP.Logger = util.Ptr(l.With().Str("service", "NPrinting").Logger())
	}

	if s := config.Sense; s != nil {
		if s.QRS != nil {
			c.QRS, res = qrs.NewClient(*s.QRS)
			if res != nil {
				return nil, res.With("qrs.NewClient")
			}
			c.QRS.SetLogger(util.Ptr(l.With().Str("service", "repository").Logger()))
		}
		if s.QPS != nil {
			c.QPS, res = qps.NewClient(*s.QPS)
			if res != nil {
				return nil, res.With("qps.NewClient")
			}
			c.QPS.SetLogger(util.Ptr(l.With().Str("service", "proxy").Logger()))
		}
	}

	return c, nil
}
