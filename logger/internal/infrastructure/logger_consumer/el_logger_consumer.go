package logger_consumer

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/RoyceAzure/rj/infra/elsearch"
	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/logger/internal/infrastructure/logger"
)

type ElLoggerConsumer struct {
	mq.IConsumer
	elLogger logger.ILogger
	closed   atomic.Bool // 添加狀態追踪
}

func NewElLoggerConsumer(elDao elsearch.IElSearchDao) (*ElLoggerConsumer, error) {
	elLogger := logger.NewElLogger(elDao)

	consumer, err := mq.NewConsumerV2("el_logger_consumer")
	if err != nil {
		elLogger.Close()
		return nil, err
	}

	return &ElLoggerConsumer{
		IConsumer: consumer,
		elLogger:  elLogger,
	}, nil
}

func (fc *ElLoggerConsumer) handler(message []byte) error {
	if fc.closed.Load() {
		return fmt.Errorf("consumer is closed")
	}

	_, err := fc.elLogger.Write(message)
	if err != nil {
		return fmt.Errorf("failed to write message to log file: %w", err)
	}
	return nil
}

func (fc *ElLoggerConsumer) Start(queueName string) error {
	if fc.closed.Load() {
		return fmt.Errorf("consumer is closed")
	}

	return fc.Consume(queueName, fc.handler)
}

func (fc *ElLoggerConsumer) Close() error {
	return errors.Join(
		fc.elLogger.Close(),
		fc.IConsumer.Close(),
	)
}
