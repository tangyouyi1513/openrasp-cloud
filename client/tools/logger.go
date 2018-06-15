package tools

import (
	log "github.com/sirupsen/logrus"
	"openrasp-cloud/client/config"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/pkg/errors"
	"path"
)

func InitLogger() {
	logDir, err := GetCurrentPath()
	if err != nil {
		log.Fatalf("Failed to log to file: v%", err)
	}
	baseLogPath := path.Join(logDir, config.LogFileName)
	writer, err := rotatelogs.New(
		baseLogPath+".%Y-%m-%d",
		rotatelogs.WithLinkName(baseLogPath),
		rotatelogs.WithMaxAge(config.LogMaxAge),
		rotatelogs.WithRotationTime(config.LogRotationTime),
	)
	if err != nil {
		log.Fatalf("config local file system logger error. %v", errors.WithStack(err))
	}

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(writer)
}
