package client

import (
	"fmt"
	"os"
	"sync"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/rs/zerolog"
)

var (
	mqFileLogger     *zerolog.Logger
	mqFileLoggerOnce sync.Once
)

func GetClientFileLogger() (*zerolog.Logger, error) {
	if mqFileLogger == nil {
		return &zerolog.Logger{}, fmt.Errorf("logger is not init, please call InitFileLogger first")
	}
	return mqFileLogger, nil
}

func InitFileLogger(params BaseMQClientLoggerParams, logFilePath string) error {
	var initErr error
	mqFileLoggerOnce.Do(func() {
		var logger zerolog.Logger
		logger, initErr = createFileLogger(params, logFilePath)
		if initErr == nil {
			mqFileLogger = &logger
		}
	})

	if initErr != nil {
		return fmt.Errorf("failed to initialize logger: %w", initErr)
	}
	return nil
}

func createFileLogger(params BaseMQClientLoggerParams, logFilePath string) (zerolog.Logger, error) {
	producer, err := mq.NewThreadSafeProducer()
	if err != nil {
		return zerolog.Logger{}, fmt.Errorf("failed to create producer: %w", err)
	}

	logger, err := NewClientFileLogger(producer, params.Exchange, params.RoutingKey, logFilePath)
	if err != nil {
		producer.Close() // 清理資源
		return zerolog.Logger{}, fmt.Errorf("failed to create client file logger: %w", err)
	}
	producer.Start()
	multiLogger := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stdout},
		logger,
	)

	return zerolog.New(multiLogger).
		With().
		Str("module_name", params.Module).
		Timestamp().
		Logger(), nil
}
