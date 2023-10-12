package ss

import (
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"
)

type Request struct {
	Script *Script
	Tasks  []TaskRunner
}

func (r *Request) ID() string {
	if r.Script.ID == "" {
		r.Script.ID = uuid.NewString()
	}
	return r.Script.ID
}

func (r *Request) Name() string {
	return r.Script.Name
}

func (r *Request) Run() (bool, []*util.Result) {
	defer r.Script.Env.CleanUp()

	logger := r.Script.Env.Logger().With().Str("script", r.Script.ID).Logger()
	logger.Info().Msg("run")
	results := make([]*util.Result, 0)
	for i, task := range r.Tasks {
		logger.Info().Msgf("running script task[%d]", i)
		res := task.Run()
		if res.Code != 0 || strings.HasSuffix(res.Ctx, CMD_NAME_REPORT) {
			results = append(results, res)
		}
		if res.Code != 0 {
			logger.Err(res).Msgf("script task[%d] failed", i)
			return false, results
		}
		logger.Info().Msgf("script task[%d] succeeded", i)
	}
	return true, results
}

func (r *Request) Logger() *zerolog.Logger {
	return r.Script.Env.Logger()
}
