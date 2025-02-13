package loggerconsumer

import (
	"io"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func SetUpMutiLogger(logger ...io.Writer) error {
	multiLogger := zerolog.MultiLevelWriter(logger...)
	Logger = zerolog.New(multiLogger).With().Timestamp().Logger()
	return nil
}

type ILoggerConsumer interface {
	Start(queueName string) error
	Close() error
}
