package logger

import (
	"github.com/sirupsen/logrus"
)

type MinimalLogger interface {
	Print(v ...interface{})
}

type decoratedLogger struct {
	logger *logrus.Logger
	level  logrus.Level
}

func (d *decoratedLogger) Print(v ...interface{}) {
	d.logger.Log(d.level, v...)
}

func DecorateAtLevel(l *logrus.Logger, level logrus.Level) MinimalLogger {
	return &decoratedLogger{
		logger: l,
		level:  level,
	}
}
