package loggerconsumer

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/RoyceAzure/rj/infra/mq"
	customerlogger "github.com/RoyceAzure/rj/logger/server/infrastructure/customer_logger"
	"github.com/RoyceAzure/rj/repo/file"
)

// 不使用DI
type FileLoggerConsumer struct {
	consumer   mq.IConsumer
	fileLogger *customerlogger.FileLogger
	closed     atomic.Bool // 添加狀態追踪
}

func NewFileLoggerConsumer(logFilePath string) (*FileLoggerConsumer, error) {
	textFileDao, err := file.NewTxtFileDAO(logFilePath)
	if err != nil {
		return nil, err
	}

	fileLogger := customerlogger.NewFileLogger(textFileDao)

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
