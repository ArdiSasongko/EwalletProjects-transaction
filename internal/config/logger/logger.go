package logger

import "github.com/sirupsen/logrus"

func NewLogger() *logrus.Logger {
	log := logrus.New()

	log.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	})

	return log
}
