package main

import (
	"github.com/haimgel/mqtt-buttons/internal/config"
	"github.com/haimgel/mqtt-buttons/internal/mqtt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

func createLogger() *zap.SugaredLogger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	loggerCfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:      true,
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := loggerCfg.Build(zap.AddStacktrace(zap.ErrorLevel), zap.WithCaller(false))
	if err != nil {
		panic(err)
	}
	return logger.Sugar()
}

func main() {
	logger := createLogger()
	// noinspection GoUnhandledErrorResult
	defer logger.Sync()
	appConfig, err := config.Load()
	if err != nil {
		logger.Panic(err)
	}
	logger.Infow("Application started", "config_file", viper.ConfigFileUsed())
	client, err := mqtt.Connect(&appConfig.Mqtt, appConfig.Switches, logger)
	if err != nil {
		logger.Panic(err)
	}
	logger.Infow("Connected to MQTT", "client", client)
	err = client.Subscribe()
	if err != nil {
		logger.Panic(err)
	}
	time.Sleep(60 * time.Second)
}
