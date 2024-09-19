package client

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"

	"github.com/soderasen-au/go-qlik/qlik/config"
	"github.com/soderasen-au/go-qlik/qlik/managed/qps"
)

func setupTestSuite(conf string, t *testing.T) (*Managed, *zerolog.Logger, func(t2 *testing.T)) {
	wd, _ := os.Getwd()
	buf, err := os.ReadFile(conf)
	if err != nil {
		t.Errorf("can't load config file: %s", err.Error())
		return nil, nil, nil
	}

	_logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}.Out).Level(zerolog.TraceLevel)

	var cfg config.Config
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		t.Errorf("can't parse config file: %s", err.Error())
		return nil, nil, nil
	}

	client, res := NewManaged(cfg, &_logger)
	if res != nil {
		t.Errorf("can't creaet managed client: %s", res.Error())
		return nil, nil, nil
	}

	_logger.Info().Msgf("test in %s using %s", wd, conf)

	return client, &_logger, func(t2 *testing.T) {}
}

func TestNewManaged(t *testing.T) {
	c, logger, _ := setupTestSuite("../../test/unittest_qlik-ci.yaml", t)
	logger.Info().Msgf("start")
	about, res := c.QRS.About()
	if res != nil {
		t.Errorf("can't get about: %s", res.Error())
		return
	}
	logger.Info().Msgf("qrs about: %v", about)

	qpsUser := &qps.User{
		UserId:  c.Config.Sense.QPS.Auth.User.Id,
		UserDir: c.Config.Sense.QPS.Auth.User.Directory,
	}

	res = c.QPS.GetWebTicket(qpsUser)
	if res != nil {
		t.Errorf("GetWebTicket: %s", res.Error())
		return
	}
	logger.Info().Msgf("ticket: %v", qpsUser)

	apps, res := c.QRS.GetAppHubList()
	if res != nil {
		t.Errorf("GetAppList: %s", res.Error())
		return
	}
	for _, app := range apps {
		if app.Thumbnail != "" {
			data, res := c.QRS.GetAppContent(app.Thumbnail)
			if res != nil {
				t.Errorf("GetAppContent: %s", res.Error())
				return
			}
			logger.Info().Msgf("qrs data: %v", string(data)[:100])
		}
	}

}
