package loggerconsumer

import "github.com/RoyceAzure/rj/infra/mq"

type ILoggerConsumer interface {
	mq.IConsumer
}
