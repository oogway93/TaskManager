package logger

import (
	"github.com/oogway93/taskmanager/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Init(cfg *config.Config) (*zap.Logger) {
	var err error

	var Log *zap.Logger
	if cfg.IsProduction() {
		Log, err = zap.NewProduction() // JSON Output
	} else {
		config := zap.NewDevelopmentConfig() // Pretty Output
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		var err error
		Log, err = config.Build()
		if err != nil {
			panic(err)
		}
	}

	if err != nil {
		panic(err)
	}
	return Log
}

func Sync(Log *zap.Logger) {
	_ = Log.Sync()
}
