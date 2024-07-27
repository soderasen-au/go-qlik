package engine

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

func setupTestSuiteHttpClient(conf string, t *testing.T) (*HttpClient, *zerolog.Logger, func(t2 *testing.T)) {
	wd, _ := os.Getwd()
	buf, err := os.ReadFile(conf)
	if err != nil {
		t.Errorf("can't load config file: %s", err.Error())
		return nil, nil, nil
	}

	var cfg Config
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		t.Errorf("can't parse config file: %s", err.Error())
		return nil, nil, nil
	}

	client, res := NewHttpClient(cfg)
	if res != nil {
		t.Errorf("can't parse config file: %s", res.Error())
		return nil, nil, nil
	}
	_logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}.Out).Level(zerolog.DebugLevel)
	client.Logger = &_logger
	_logger.Info().Msgf("test in %s using %s", wd, conf)

	return client, &_logger, func(t2 *testing.T) {}
}

func setupTestSuiteConn(conf string, t *testing.T) (*Conn, *zerolog.Logger, func(t2 *testing.T)) {
	wd, _ := os.Getwd()
	buf, err := os.ReadFile(conf)
	if err != nil {
		t.Errorf("can't load config file: %s", err.Error())
		return nil, nil, nil
	}

	var cfg Config
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		t.Errorf("can't parse config file: %s", err.Error())
		return nil, nil, nil
	}

	conn, err := NewConn(cfg)
	if err != nil {
		t.Errorf("%s error: %s", "NewConn", err.Error())
	}
	_logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}.Out).Level(zerolog.DebugLevel)
	_logger.Info().Msgf("test in %s using %s", wd, conf)

	return conn, &_logger, func(t2 *testing.T) {}
}
