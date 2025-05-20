package logger_producer

import (
	"fmt"
	"os"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/rs/zerolog"
)

type IClientFactory interface {
	GetLoggerProcuder() (*zerolog.Logger, error)
}

var (
	mqFileLogger *zerolog.Logger
)

type FileLoggerFactory struct {
	config *BaseMQClientLoggerParams
}

func NewFileLoggerFactory(config *BaseMQClientLoggerParams) (*FileLoggerFactory, error) {
	return &FileLoggerFactory{
		config: config,
	}, nil
}

func (e *FileLoggerFactory) GetLoggerProcuder() (*zerolog.Logger, error) {
	if mqElLogger != nil {
		return mqElLogger, nil
	}

	logger, err := e.createFileLogger()
	if err != nil {
		return nil, err
	}

	mqFileLogger = logger
	return mqElLogger, nil
}

func (e *FileLoggerFactory) createFileLogger() (*zerolog.Logger, error) {
	producer, err := mq.NewThreadSafeProducer()
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	logger, err := NewClientFileLogger(producer, e.config.Exchange, e.config.RoutingKey, e.config.LogFileSavePath)
	if err != nil {
		producer.Close() // 清理資源
		return nil, fmt.Errorf("failed to create client file logger: %w", err)
	}
	producer.Start()

	zeroLogger := setUpClientLToZeroL(logger, e.config)
	return &zeroLogger, nil
}

var (
	mqElLogger *zerolog.Logger
)

// Elasticsearch 工廠
type ElasticFactory struct {
	config *BaseMQClientLoggerParams
}

func NewElasticFactory(config *BaseMQClientLoggerParams) (*ElasticFactory, error) {
	return &ElasticFactory{
		config: config,
	}, nil
}

func (e *ElasticFactory) GetLoggerProcuder() (*zerolog.Logger, error) {
	if mqElLogger != nil {
		return mqElLogger, nil
	}

	logger, err := e.createLoggerProcuder()
	if err != nil {
		return nil, err
	}

	mqElLogger = logger
	return mqElLogger, nil
}

func (e *ElasticFactory) createLoggerProcuder() (*zerolog.Logger, error) {
	producer, err := mq.NewThreadSafeProducer()
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	el_logger, err := NewClientELLogger(producer, e.config.Exchange, e.config.RoutingKey)
	if err != nil {
		producer.Close() // 清理資源
		return nil, fmt.Errorf("failed to create client el logger: %w", err)
	}
	producer.Start()

	zeroLogger := setUpClientLToZeroL(el_logger, e.config)
	return &zeroLogger, nil
}

// ILoggerProcuder 轉換成zero logger
func setUpClientLToZeroL(logger ILoggerProcuder, config *BaseMQClientLoggerParams) zerolog.Logger {
	multiLogger := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stdout},
		logger,
	)

	l := zerolog.New(multiLogger).With()

	if config.Project != "" {
		l = l.Str("project", config.Project)
	}

	if config.Module != "" {
		l = l.Str("module", config.Module)
	}

	return l.
		Timestamp().
		Logger()
}
