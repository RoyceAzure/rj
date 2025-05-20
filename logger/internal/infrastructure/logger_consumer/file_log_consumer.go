package logger_consumer

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/logger/internal/infrastructure/logger"
	"github.com/RoyceAzure/rj/repo/file"
)

// 不使用DI
type FileLoggerConsumer struct {
	consumer   mq.IConsumer
	fileLogger logger.ILogger
	closed     atomic.Bool // 添加狀態追踪
}

func NewFileLoggerConsumer(logFilePath string) (*FileLoggerConsumer, error) {
	textFileDao, err := file.NewTxtFileDAO(logFilePath)
	if err != nil {
		return nil, err
	}

	fileLogger := logger.NewFileLogger(textFileDao)

	consumer, err := mq.NewConsumerV2("file_consumer")
	if err != nil {
		fileLogger.Close()
		return nil, err
	}

	return &FileLoggerConsumer{
		consumer:   consumer,
		fileLogger: fileLogger,
	}, nil
}

func (fc *FileLoggerConsumer) handler(message []byte) error {
	if fc.closed.Load() {
		return fmt.Errorf("consumer is closed")
	}

	_, err := fc.fileLogger.Write(message)
	if err != nil {
		return fmt.Errorf("failed to write message to log file: %w", err)
	}
	return nil
}

func (fc *FileLoggerConsumer) Start(queueName string) error {
	if fc.closed.Load() {
		return fmt.Errorf("consumer is closed")
	}

	return fc.consumer.Consume(queueName, fc.handler)
}

func (fc *FileLoggerConsumer) Close() error {
	return errors.Join(
		fc.fileLogger.Close(),
		fc.consumer.Close(),
	)
}
