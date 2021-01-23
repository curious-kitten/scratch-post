package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/curious-kitten/scratch-post/internal/info"
)

// Logger interface exposes the necessary logger functions
type Logger interface {
	Info(args ...interface{})
	Infof(msg string, args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Debug(args ...interface{})
	Debugf(msg string, args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Error(args ...interface{})
	Errorf(msg string, args ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatal(args ...interface{})
}

// New returns a new Logger instance with the desired fields
func New(app info.App, instance info.Instance, debug bool) (Logger, func() error, error) {
	initFields := map[string]interface{}{
		"app":      app.Name,
		"version":  app.Version,
		"instance": instance.Name,
	}
	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}
	zLogger, err := zap.Config{
		Encoding:          "json",
		Level:             zap.NewAtomicLevelAt(level),
		DisableCaller:     false,
		DisableStacktrace: false,
		OutputPaths:       []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			MessageKey: "msg",
		},
		InitialFields: initFields,
	}.Build()
	if err != nil {
		return nil, nil, err
	}
	log := zLogger.Sugar()
	return log, log.Sync, nil
}

// WithFields adds fields to the provided logger instance
func WithFields(log Logger, fields map[string]interface{}) Logger {
	args := make([]interface{}, len(fields)*2)
	i := 0
	for k, v := range fields {
		args[i] = k
		i++
		args[i] = v
		i++
	}
	return log.(*zap.SugaredLogger).With(args...)
}
