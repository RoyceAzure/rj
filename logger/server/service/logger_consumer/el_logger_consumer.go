package loggerconsumer

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/RoyceAzure/rj/infra/elsearch"
	"github.com/RoyceAzure/rj/infra/mq"
	customerlogger "github.com/RoyceAzure/rj/logger/server/infrastructure/customer_logger"
)

type ElLoggerConsumer struct {
	consumer mq.IConsumer
	elLogger customerlogger.ILogger
	closed   atomic.Bool // 添加狀態追踪
}

func NewElLoggerConsumer(elDao elsearch.IElSearchDao) (*ElLoggerConsumer, error) {
	elLogger := customerlogger.NewElLogger(elDao)

	consumer, err := mq.NewConsumerV2("el_logger_consumer")
	if err != nil {
		elLogger.Close()
		return nil, err
	}

	return &ElLoggerConsumer{
		consumer: consumer,
		elLogger: elLogger,
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

	return fc.consumer.Consume(queueName, fc.handler)
}

func (fc *ElLoggerConsumer) Close() error {
	return errors.Join(
		fc.elLogger.Close(),
		fc.consumer.Close(),
	)
}
