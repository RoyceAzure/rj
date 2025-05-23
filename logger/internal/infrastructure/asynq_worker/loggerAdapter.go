package asynq_worker

import (
	"fmt"

	"github.com/RoyceAzure/rj/logger/internal/infrastructure/logger_consumer"
	"github.com/rs/zerolog"
)

/*
zerolog adapter
*/
type loggerAdapter struct {
}

func NewLoggerAdapter() *loggerAdapter {
	return &loggerAdapter{}
}

/*
use to call zerolog.log.Withlevel
*/
func Print(level zerolog.Level, args ...interface{}) {
	logger_consumer.Logger.WithLevel(level).Msg(fmt.Sprint(args...))
}

func (logger loggerAdapter) Debug(args ...interface{}) {
	logger_consumer.Logger.WithLevel(zerolog.DebugLevel).Msg(fmt.Sprint(args...))
}

// Info logs a message at Info level.
func (logger loggerAdapter) Info(args ...interface{}) {
	logger_consumer.Logger.WithLevel(zerolog.InfoLevel).Msg(fmt.Sprint(args...))
}

// Warn logs a message at Warning level.
func (logger loggerAdapter) Warn(args ...interface{}) {
	logger_consumer.Logger.WithLevel(zerolog.WarnLevel).Msg(fmt.Sprint(args...))
}

// Error logs a message at Error level.
func (logger loggerAdapter) Error(args ...interface{}) {
	logger_consumer.Logger.WithLevel(zerolog.ErrorLevel).Msg(fmt.Sprint(args...))
}

// Fatal logs a message at Fatal level
// and process will exit with status set to 1.
func (logger loggerAdapter) Fatal(args ...interface{}) {
	logger_consumer.Logger.WithLevel(zerolog.FatalLevel).Msg(fmt.Sprint(args...))
}
