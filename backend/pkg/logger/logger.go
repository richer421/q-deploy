package logger

import (
	"os"

	"github.com/richer421/q-deploy/conf"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var sugar *zap.SugaredLogger

func Init(cfg conf.LogConfig) {
	// 1. Parse level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// 2. Build encoder
	var encoder zapcore.Encoder
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	switch cfg.Format {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	default:
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	// 3. Build write syncers — stdout is always on
	cores := []zapcore.Core{
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level),
	}

	// 4. Optionally add file syncer with lumberjack rotation
	if cfg.File.Enabled {
		fileEncoder := zapcore.NewJSONEncoder(encoderCfg) // file always JSON
		fileSyncer := zapcore.AddSync(&lumberjack.Logger{
			Filename: cfg.File.Path,
			MaxSize:  cfg.File.MaxSize,
			MaxAge:   cfg.File.MaxAge,
			Compress: cfg.File.Compress,
		})
		cores = append(cores, zapcore.NewCore(fileEncoder, fileSyncer, level))
	}

	// 5. Build logger
	core := zapcore.NewTee(cores...)
	l := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = l.Sugar()
}

func Sync() {
	if sugar != nil {
		_ = sugar.Sync()
	}
}

func Debugf(template string, args ...interface{}) { sugar.Debugf(template, args...) }
func Infof(template string, args ...interface{})  { sugar.Infof(template, args...) }
func Warnf(template string, args ...interface{})  { sugar.Warnf(template, args...) }
func Errorf(template string, args ...interface{}) { sugar.Errorf(template, args...) }
func Fatalf(template string, args ...interface{}) { sugar.Fatalf(template, args...) }
