package logger_consumer

import (
	"io"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func SetUpMutiLogger(logger ...io.Writer) error {
	multiLogger := zerolog.MultiLevelWriter(logger...)
	Logger = zerolog.New(multiLogger).With().Timestamp().Logger()
	return nil
}

type ILoggerConsumer interface {
	mq.IConsumer
	Start(queueName string) error
	Close() error
}
