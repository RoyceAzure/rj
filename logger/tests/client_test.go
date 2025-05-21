package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/logger/client/producer"
	"github.com/stretchr/testify/require"
)

func TestLoggerProducer(t *testing.T) {
	err := mq.SelectConnFactory.Init(mq.MQConnParams{
		MqHost:  "localhost",
		MqUser:  "royce",
		MqPas:   "password",
		MqPort:  "5672",
		MqVHost: "/",
	})
	require.NoError(t, err)

	var f producer.ILoggerProducerFactory
	f, err = producer.NewElasticFactory(&producer.LoggerProducerConfig{
		Exchange:   "system_logs",
		RoutingKey: "log.file.el",
		Module:     "logger-client",
		Project:    "test",
	})
	require.NoError(t, err)

	loggerProducer, err := f.GetLoggerProcuder()
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		loggerProducer.Info().
			Str("action", fmt.Sprintf("unit test %d", i)).
			Str("url", fmt.Sprintf("test url %d", i)).
			Str("upn", fmt.Sprintf("test upn %d", i)).
			Str("random", fmt.Sprintf("test random %d", i)).
			Msg(fmt.Sprintf("this is test %d", i))
		time.Sleep(1 * time.Second)
	}
}
