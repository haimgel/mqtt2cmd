package logging

import (
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

func CreateLogger(logFile string) *zap.SugaredLogger {
	var syncs []zapcore.WriteSyncer
	if logFile != "" {
		syncs = []zapcore.WriteSyncer{zapcore.AddSync(&lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     0,
		}), zapcore.AddSync(os.Stderr)}
	} else {
		syncs = []zapcore.WriteSyncer{zapcore.AddSync(os.Stderr)}
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zap.CombineWriteSyncers(syncs...),
		zap.DebugLevel,
	)
	logger := zap.New(core).WithOptions(
		zap.AddStacktrace(zap.ErrorLevel),
		zap.WithCaller(false),
	)
	return logger.Sugar().With(zap.Int("pid", os.Getpid()))
}

type MqttLogger struct {
	level zapcore.Level
	inner *zap.SugaredLogger
}

func (logger MqttLogger) Println(v ...interface{}) {
	switch logger.level {
	case zapcore.DebugLevel:
		logger.inner.Debugln(v)
	case zapcore.WarnLevel:
		logger.inner.Warnln(v)
	default:
		logger.inner.Errorln(v)
	}
}

func (logger MqttLogger) Printf(format string, v ...interface{}) {
	switch logger.level {
	case zapcore.DebugLevel:
		logger.inner.Debugf(format, v)
	case zapcore.WarnLevel:
		logger.inner.Warnf(format, v)
	default:
		logger.inner.Errorf(format, v)
	}
}

func InstrumentMqtt(logger *zap.SugaredLogger) {
	MQTT.CRITICAL = MqttLogger{inner: logger.With(zap.String("component", "mqtt")), level: zapcore.ErrorLevel}
	MQTT.ERROR = MqttLogger{inner: logger.With(zap.String("component", "mqtt")), level: zapcore.ErrorLevel}
	MQTT.WARN = MqttLogger{inner: logger.With(zap.String("component", "mqtt")), level: zapcore.WarnLevel}
	// Debug logging is disabled, it's too chatty!
	// MQTT.DEBUG = MqttLogger{inner: logger.With(zap.String("component", "mqtt")), level: zapcore.DebugLevel}
}
