package log

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	Logger       = logrus.New()
	EventLogger  = logrus.New()
	MailLogger   = logrus.New()
	HTTPLogger   = logrus.New()
	DBLogger     = logrus.New()
	ConfigLogger = logrus.New()
)

func InitLogger(logger *logrus.Logger, path string) error {
	writer, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("open log file error: ", err)
		return err
	}
	logger.Out = writer
	logger.Formatter = &logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
	}
	logger.Level = logrus.DebugLevel
	return nil
}

func InitJsonLogger(logger *logrus.Logger, path string) error {
	writer, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("open log file error: ", err)
		return err
	}
	logger.Out = writer
	logger.Formatter = &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	}
	logger.Level = logrus.InfoLevel
	return nil
}
