package test_helpers

import (
	"github.com/go-kit/kit/log"
	kitLogrus "github.com/go-kit/kit/log/logrus"
	"github.com/sirupsen/logrus"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmlogrus"
	"os"
	"testing"
)

var logrusLogger = &logrus.Logger{
	Out:       os.Stdout,
	Hooks:     make(logrus.LevelHooks),
	Level:     logrus.DebugLevel,
	Formatter: &logrus.JSONFormatter{},
}

func LogrusLogger(t *testing.T) log.Logger {
	t.Helper()

	logrusLogger.Out = log.NewSyncWriter(os.Stdout)
	logger := kitLogrus.NewLogrusLogger(logrusLogger)

	apm.DefaultTracer.SetLogger(logrusLogger)
	logrusLogger.AddHook(&apmlogrus.Hook{})

	return log.NewSyncLogger(logger)
}
