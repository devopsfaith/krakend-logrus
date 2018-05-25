//Package logrus provides a logger implementation based on the github.com/sirupsen/logrus pkg
package logrus

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/devopsfaith/krakend/config"
	"github.com/sirupsen/logrus"
)

// Namespace is the key to look for extra configuration details
const Namespace = "github_com/devopsfaith/krakend-logrus"

// ErrWrongConfig is the error returned when there is no config under the namespace
var ErrWrongConfig = errors.New("getting the extra config for the krakend-logrus module")

// NewLogger returns a krakend logger wrapping a logrus logger
func NewLogger(cfg config.ExtraConfig, ws ...io.Writer) (*Logger, error) {
	logConfig, ok := ConfigGetter(cfg).(Config)
	if !ok {
		return nil, ErrWrongConfig
	}

	level, ok := logLevels[logConfig.Level]
	if !ok {
		return nil, fmt.Errorf("unknown log level: %s", logConfig.Level)
	}

	l := logrus.New()
	setFormatter(l, logConfig)
	setOutput(l, logConfig, ws...)
	l.Level = logrus.DebugLevel

	return &Logger{
		logger: l,
		level:  level,
		module: logConfig.Module,
	}, nil
}

func setFormatter(l *logrus.Logger, cfg Config) {
	switch {
	case cfg.JSONFormatter != nil:
		l.Formatter = cfg.JSONFormatter
	case cfg.TextFormatter != nil:
		l.Formatter = cfg.TextFormatter
	default:
		l.Formatter = &logrus.TextFormatter{}
	}
}

func setOutput(l *logrus.Logger, cfg Config, ws ...io.Writer) {
	if cfg.StdOut {
		ws = append(ws, os.Stdout)
	}
	if cfg.Syslog {
		// ws = append(ws, b)
	}

	if len(ws) == 1 {
		l.Out = ws[0]
		return
	}
	l.Out = io.MultiWriter(ws...)
}

// ConfigGetter implements the config.ConfigGetter interface
func ConfigGetter(e config.ExtraConfig) interface{} {
	v, ok := e[Namespace]
	if !ok {
		return nil
	}
	cfg := Config{}

	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	if json.Unmarshal(data, &cfg); err != nil {
		return nil
	}

	return cfg
}

// Config is the custom config struct containing the params for the logger
type Config struct {
	Level         string                `json:"level"`
	StdOut        bool                  `json:"stdout"`
	Syslog        bool                  `json:"syslog"`
	Module        string                `json:"module"`
	TextFormatter *logrus.TextFormatter `json:"text"`
	JSONFormatter *logrus.JSONFormatter `json:"json"`
}

// Logger is a wrapper over a github.com/sirupsen/logrus logger
type Logger struct {
	logger *logrus.Logger
	level  logrus.Level
	module string
}

// Debug implements the logger interface
func (l *Logger) Debug(v ...interface{}) {
	if l.level < logrus.DebugLevel {
		return
	}
	l.logger.WithField("module", l.module).Debug(v...)
}

// Info implements the logger interface
func (l *Logger) Info(v ...interface{}) {
	if l.level < logrus.InfoLevel {
		return
	}
	l.logger.WithField("module", l.module).Info(v...)
}

// Warning implements the logger interface
func (l *Logger) Warning(v ...interface{}) {
	if l.level < logrus.WarnLevel {
		return
	}
	l.logger.WithField("module", l.module).Warning(v...)
}

// Error implements the logger interface
func (l *Logger) Error(v ...interface{}) {
	if l.level < logrus.ErrorLevel {
		return
	}
	l.logger.WithField("module", l.module).Error(v...)
}

// Critical implements the logger interface but demotes to the error level
func (l *Logger) Critical(v ...interface{}) {
	l.logger.WithField("module", l.module).Error(v...)
}

// Fatal implements the logger interface
func (l *Logger) Fatal(v ...interface{}) {
	l.logger.WithField("module", l.module).Fatal(v...)
}

var logLevels = map[string]logrus.Level{
	"DEBUG":    logrus.DebugLevel,
	"INFO":     logrus.InfoLevel,
	"WARNING":  logrus.WarnLevel,
	"ERROR":    logrus.ErrorLevel,
	"CRITICAL": logrus.FatalLevel,
}
