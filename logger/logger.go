package logger

import (
	"github.com/oogway93/taskmanager/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init(cfg *config.Config) {
	var err error

	if cfg.IsProduction() {
		Log, err = zap.NewProduction() // JSON Output
	} else {
		config := zap.NewDevelopmentConfig() // Pretty Output

		// Включаем цветной вывод для уровней
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		// Включаем цветной вывод для времени (опционально)
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
}

func Sync() {
	_ = Log.Sync()
}
