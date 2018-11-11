package main

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type LoggerOptions struct {
	Application string
	//UseSentry   bool
	LogFile string
	//SentryDSN   string
}

func NewLogger(options LoggerOptions) *logrus.Entry {

	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{
		//ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		//DisableColors:   true,
	}
	// &logrus.JSONFormatter{}

	if options.LogFile != "" {
		log.Out = os.Stdout
		file, err := os.OpenFile(options.LogFile, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			log.Out = io.MultiWriter(file, os.Stdout)
		} else {
			log.Info("Failed to log to file, using default stderr")
		}
	}

	// if options.UseSentry {
	// 	hook, err := logrus_sentry.NewSentryHook(sentryDSN, []logrus.Level{
	// 		logrus.PanicLevel,
	// 		logrus.FatalLevel,
	// 		logrus.ErrorLevel,
	// 	})
	// 	if err == nil {
	// 		log.Hooks.Add(hook)
	// 	}
	// }

	logger := log.WithFields(logrus.Fields{"app": options.Application})
	return logger

}
