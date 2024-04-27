package main

import (
	"github.com/markgregr/bestHack_support_REST_server/internal/app"
	"github.com/markgregr/bestHack_support_REST_server/internal/config"
	"github.com/markgregr/bestHack_support_REST_server/internal/lib/logger/handlers/logruspretty"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env, cfg.LogsPath)

	log.WithField("config", cfg).Info("Application start!")

	application, err := app.New(cfg, log)
	if err != nil {
		panic(err)
	}

	application.Run()

	<-application.Done
	log.Info("Application stopped")
}

func setupLogger(env string, logFilePath string) *logrus.Entry {
	var log = logrus.New()

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	switch env {
	case envLocal:
		log.SetLevel(logrus.DebugLevel)
		return setupPrettySlog(log)
	case envDev:
		log.SetOutput(logFile)
		log.SetFormatter(&logrus.TextFormatter{
			DisableColors: true,
			FullTimestamp: true, // Добавляем временные метки к сообщениям
		})
		log.SetLevel(logrus.InfoLevel)
	case envProd:
		log.SetOutput(logFile)
		log.SetFormatter(&logrus.TextFormatter{
			DisableColors: true,
			FullTimestamp: true, // Добавляем временные метки к сообщениям
		})
		log.SetLevel(logrus.WarnLevel)
	default:
		log.SetOutput(logFile)
		log.SetFormatter(&logrus.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
		log.SetLevel(logrus.WarnLevel)
	}

	return logrus.NewEntry(log)
}

func setupPrettySlog(log *logrus.Logger) *logrus.Entry {
	prettyHandler := logruspretty.NewPrettyHandler(os.Stdout)
	log.SetFormatter(prettyHandler)
	return logrus.NewEntry(log)
}
