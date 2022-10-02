package logging

import (
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
