package logger

import (
	log "github.com/sirupsen/logrus"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/pkg/errors"
	"path"
	"os"
	"time"
	"openrasp-cloud/tools"
	. "openrasp-cloud/config"
)

func init() {
	logDir, err := tools.GetCurrentPath()
	if err != nil {
		log.Panicf("Failed to log to file: v%", err)
	}
	err = os.MkdirAll(path.Join(logDir, "logs"), os.ModePerm)
	if err != nil {
		log.Panicf("config local file system logger error. %v", errors.WithStack(err))
	}
	baseLogPath := path.Join(logDir, "logs", Conf.LogFileName)
	writer, err := rotatelogs.New(
		baseLogPath+".%Y-%m-%d",
		rotatelogs.WithLinkName(baseLogPath),
		rotatelogs.WithMaxAge(time.Duration(Conf.LogMaxAge*24)*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(Conf.LogRotationTime)*time.Hour),
	)
	if err != nil {
		log.Panicf("config local file system logger error. %v", errors.WithStack(err))
	}

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(writer)
}
