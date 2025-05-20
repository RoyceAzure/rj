package tests

import (
	"testing"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/logger/internal/infrastructure/logger_producer"
	"github.com/stretchr/testify/require"
)

func TestXxx(t *testing.T) {
	err := mq.SelectConnFactory.Init(mq.MQConnParams{
		MqHost:  "localhost",
		MqUser:  "royce",
		MqPas:   "password",
		MqPort:  "5672",
		MqVHost: "/",
	})
	require.NoError(t, err)

	var f logger_producer.IClientFactory
	f, err = logger_producer.NewElasticFactory(&logger_producer.BaseMQClientLoggerParams{
		Exchange:   "system_logs",
		RoutingKey: "log.file.el",
		Module:     "logger-client",
		Project:    "test",
	})
	require.NoError(t, err)

	loggerProducer, err := f.GetLoggerProcuder()
	require.NoError(t, err)
	loggerProducer.Info().
		Str("action", "unit test 22").
		Str("url", "test url").
		Str("upn", "test upn").
		Str("random", "test random").
		Msg("this is test 38")
}
