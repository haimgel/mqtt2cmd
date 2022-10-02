package main

import (
	"github.com/haimgel/mqtt2cmd/internal/config"
	"github.com/haimgel/mqtt2cmd/internal/logging"
	"github.com/haimgel/mqtt2cmd/internal/mqtt"
	"github.com/spf13/viper"
	"time"
)

func main() {
	appConfig, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := logging.CreateLogger(appConfig.LoggerConfig.Path)
	// noinspection GoUnhandledErrorResult
	defer logger.Sync()
	logger.Infow("Application started", "config_file", viper.ConfigFileUsed())

	client, err := mqtt.Init(&appConfig.Mqtt, appConfig.Switches, logger)
	if err != nil {
		logger.Panic(err)
	}
	for {
		time.Sleep(10 * time.Second)
		client.Refresh(false)
	}
}
