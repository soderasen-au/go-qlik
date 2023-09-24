package ss

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/qlik/engine"
	"github.com/soderasen-au/go-qlik/qlik/managed/qrs"
)

type ExecEnv struct {
	EngineConn *engine.Conn
	Doc        *enigma.Doc
	qrsClient  *qrs.Client
	AppID      string
	Log        *zerolog.Logger `json:"-"`
	bmMap      map[string]string
	csOrder    map[string]int
	dims       map[string]*engine.SessionDimensionLayout
	measures   map[string]*engine.SessionMeasureLayout
}

type ExecEnvOption func(env *ExecEnv) *ExecEnv

func WithQrsClient(c *qrs.Client) ExecEnvOption {
	return func(env *ExecEnv) *ExecEnv {
		env.qrsClient = c
		return env
	}
}

// NewExecEnv Note: please call CleanUp() afterwards to close engine connection properly.
// Script.Run() calls CleanUp() automatically
func NewExecEnv(cfg *engine.Config, appid string, logger *zerolog.Logger, opts ...ExecEnvOption) (*ExecEnv, *util.Result) {
	res := cfg.QCSEngineURIAppendAppID(appid)
	if res != nil {
		return nil, res.With("AppendAppID")
	}

	env := new(ExecEnv)

	conn, err := engine.NewConn(*cfg)
	if err != nil {
		return nil, util.Error("NewConn", err)
	}
	env.EngineConn = conn

	env.AppID = appid

	if logger == nil {
		_ = env.CreateLogger()
	} else {
		env.Log = logger
	}

	env.Doc, err = env.EngineConn.Global.OpenDoc(engine.ConnCtx, env.AppID, "", "", "", false)
	if err != nil {
		return nil, util.Error("OpenDoc", err)
	}

	res = env.GetBookmarkMap()
	if res != nil {
		return nil, res.With("GetBookmarkMap")
	}

	env.csOrder = make(map[string]int)

	if opts != nil {
		for _, opt := range opts {
			env = opt(env)
		}
	}

	return env, nil
}

func (env *ExecEnv) CreateLogger() error {
	logFolder := "logs"
	_ = util.MaybeCreate(logFolder)
	logFile := filepath.Join(logFolder, fmt.Sprintf("SmallScript-%s.log", time.Now().Format("2006-01-02-15_04_05")))
	logger, err := loggers.GetLogger(logFile)
	if err != nil {
		return fmt.Errorf("can't get logger: %s", err.Error())
	}
	env.Log = logger

	return nil
}

func (env *ExecEnv) Logger() *zerolog.Logger {
	if env.Log == nil {
		_ = env.CreateLogger()
	}

	return env.Log
}

func (env *ExecEnv) LogErr(ctx string, err error) error {
	res := util.Error(ctx, err)
	return env.LogErrorResult(res)
}

func (env *ExecEnv) LogErrMsg(ctx, msg string) error {
	res := util.MsgError(ctx, msg)
	return env.LogErrorResult(res)
}

func (env *ExecEnv) LogErrorResult(res *util.Result) error {
	if res != nil {
		env.Logger().Error().Msg(res.Error())
		return res
	}
	return nil
}

func (env *ExecEnv) CleanUp() {
	if env.EngineConn != nil && env.EngineConn.Global != nil {
		env.Logger().Info().Msg("disconnect from qlik engine")
		env.EngineConn.Global.DisconnectFromServer()
	}
}

func (env *ExecEnv) GetBookmarkMap() *util.Result {
	env.bmMap = make(map[string]string)
	sessionBMs, res := engine.GetSessionBookmarks(env.Doc)
	if res != nil {
		return res.With("GetSessionBookmarks")
	}

	for _, bm := range sessionBMs {
		if bm.Meta.Title == nil {
			env.Log.Warn().Msgf("BM %s doesn't have title", bm.Info.Id)
			continue
		}
		env.Log.Debug().Msgf("Init BM: '%s' => %s", *bm.Meta.Title, bm.Info.Id)
		env.bmMap[*bm.Meta.Title] = bm.Info.Id
	}

	return nil
}

func (env *ExecEnv) HasBookmark(title string) bool {
	if _, ok := env.bmMap[title]; ok {
		return true
	}
	return false
}

func (env *ExecEnv) SyncBookmark(title string) (bool, *util.Result) {
	for i := 0; i < 10; i++ {
		if env.HasBookmark(title) {
			return true, nil
		}

		time.Sleep(3 * time.Second)
		res := env.GetBookmarkMap()
		if res != nil {
			return false, res.With("GetBookmarkMap")
		}
	}
	return false, nil
}

func (env *ExecEnv) GetMasterItemsMap() *util.Result {
	env.dims = make(map[string]*engine.SessionDimensionLayout)
	list, res := engine.GetDimensionList(env.Doc)
	if res != nil {
		return res.With("GetDimensionList")
	}
	for _, m := range list {
		env.Log.Debug().Msgf("Init dim: '%s' => %s", m.Info.Id, *m.Meta.Title)
		env.dims[m.Info.Id] = m
	}

	env.measures = make(map[string]*engine.SessionMeasureLayout)
	mlist, res := engine.GetMeasureList(env.Doc)
	if res != nil {
		return res.With("GetDimensionList")
	}
	for _, m := range mlist {
		env.Log.Debug().Msgf("Init measure: '%s' => %s", m.Info.Id, *m.Meta.Title)
		env.measures[m.Info.Id] = m
	}
	return nil
}

func (env *ExecEnv) GetMeasureByName(name string) (*engine.SessionMeasureLayout, bool) {
	l, ok := env.measures[name]
	return l, ok
}

func (env *ExecEnv) GetDimensionByName(name string) (*engine.SessionDimensionLayout, bool) {
	l, ok := env.dims[name]
	return l, ok
}
