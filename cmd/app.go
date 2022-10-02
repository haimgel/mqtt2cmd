package cmd

import (
	"github.com/haimgel/mqtt2cmd/internal/config"
	"github.com/haimgel/mqtt2cmd/internal/logging"
	"github.com/haimgel/mqtt2cmd/internal/mqtt"
	"github.com/spf13/viper"
	"time"
)

func Execute(version string, exit func(int), args []string) {
	appConfig, err := config.Load(version, exit, args)
	if err != nil {
		panic(err)
	}

	logger := logging.CreateLogger(appConfig.LoggerConfig.Path)
	// noinspection GoUnhandledErrorResult
	defer logger.Sync()
	logger.Infow("Application started", "config_file", viper.ConfigFileUsed())

	logging.InstrumentMqtt(logger)
	client, err := mqtt.Init(appConfig.AppId, &appConfig.Mqtt, appConfig.Switches, logger)
	if err != nil {
		logger.Panic(err)
	}
	for {
		client.Refresh()
		time.Sleep(10 * time.Second)
	}
}
