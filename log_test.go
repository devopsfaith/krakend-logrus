package logrus

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	debugMsg    = "Debug msg"
	infoMsg     = "Info msg"
	warningMsg  = "Warning msg"
	errorMsg    = "Error msg"
	criticalMsg = "Critical msg"
)

func TestNewLogger(t *testing.T) {
	levels := []string{"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"}
	regexps := []*regexp.Regexp{
		regexp.MustCompile(debugMsg),
		regexp.MustCompile(infoMsg),
		regexp.MustCompile(warningMsg),
		regexp.MustCompile(errorMsg),
		regexp.MustCompile(criticalMsg),
	}

	for i, level := range levels {
		output, err := logSomeStuff(level)
		if err != nil {
			t.Error(err)
			return
		}
		time.Sleep(100 * time.Millisecond)
		for j := i; j < len(regexps); j++ {
			if !regexps[j].MatchString(output) {
				t.Errorf("test #%d: The output doesn't contain the expected msg for the level: %s. [%s]", i, level, output)
			}
		}
	}
}

func BenchmarkNewLogger(b *testing.B) {
	buff := bytes.NewBuffer(make([]byte, 10*1024*1024))
	logger, err := NewLogger(newExtraConfig("DEBUG"), buff)
	if err != nil {
		b.Error(err.Error())
		return
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug(debugMsg)
	}
}

func BenchmarkNewLogger_parallel(b *testing.B) {
	buff := bytes.NewBuffer(make([]byte, 10*1024*1024))
	logger, err := NewLogger(newExtraConfig("DEBUG"), buff)
	if err != nil {
		b.Error(err.Error())
		return
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Debug(debugMsg)
		}
	})
}

func ExampleNewLogger() {
	cfg := map[string]interface{}{
		Namespace: Config{
			Level:  "DEBUG",
			Module: "pref",
			Syslog: false,
			StdOut: false,
			TextFormatter: &logrus.TextFormatter{
				DisableTimestamp: true,
			},
		},
	}
	buff := new(bytes.Buffer)
	logger, err := NewLogger(cfg, buff)
	if err != nil {
		fmt.Println(err.Error())
	}

	logger.Debug(debugMsg)
	logger.Info(infoMsg)
	logger.Warning(warningMsg)
	logger.Error(errorMsg)
	logger.Critical(criticalMsg)

	fmt.Println(buff.String())
	// output:
	// level=debug msg="Debug msg" module=pref
	// level=info msg="Info msg" module=pref
	// level=warning msg="Warning msg" module=pref
	// level=error msg="Error msg" module=pref
	// level=error msg="Critical msg" module=pref
}

func ExampleConfigGetter() {
	cfg := ConfigGetter(newExtraConfig("DEBUG")).(Config)
	fmt.Println(cfg.JSONFormatter)
	fmt.Printf("%+v\n", cfg.TextFormatter)
	// output:
	// <nil>
	// &{ForceColors:false DisableColors:false DisableTimestamp:false FullTimestamp:true TimestampFormat: DisableSorting:false QuoteEmptyFields:false isTerminal:false Once:{m:{state:0 sema:0} done:0}}
}

func TestNewLogger_unknownLevel(t *testing.T) {
	_, err := NewLogger(newExtraConfig("UNKNOWN"), bytes.NewBuffer(make([]byte, 1024)))
	if err == nil {
		t.Error("The factory didn't return the expected error")
		return
	}
	if err.Error() != "unknown log level: UNKNOWN" {
		t.Errorf("The factory didn't return the expected error. Got: %s", err.Error())
	}
}

func newExtraConfig(level string) map[string]interface{} {
	return map[string]interface{}{
		Namespace: Config{
			Level:  level,
			Module: "pref",
			Syslog: false,
			StdOut: false,
			TextFormatter: &logrus.TextFormatter{
				FullTimestamp: true,
			},
		},
	}
}

func logSomeStuff(level string) (string, error) {
	buff := new(bytes.Buffer)
	logger, err := NewLogger(newExtraConfig(level), buff)
	if err != nil {
		return "", err
	}

	logger.Debug(debugMsg)
	logger.Info(infoMsg)
	logger.Warning(warningMsg)
	logger.Error(errorMsg)
	logger.Critical(criticalMsg)

	return buff.String(), nil
}
