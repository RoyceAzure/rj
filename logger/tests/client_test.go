package tests

import (
	"testing"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/logger/client"
	"github.com/stretchr/testify/require"
)

func TestXxx(t *testing.T) {
	err := mq.SelectConnFactory.Init(mq.MQConnParams{
		MqHost:  "localhost",
		MqUser:  "stock_ana_mq",
		MqPas:   "123456",
		MqPort:  "5672",
		MqVHost: "/",
	})
	require.NoError(t, err)

	var f client.IClientFactory
	f, err = client.NewElasticFactory(&client.BaseMQClientLoggerParams{
		Exchange:   "system_logs",
		RoutingKey: "log.file.el",
		Module:     "logger-client",
		Project:    "logger",
	})
	require.NoError(t, err)

	loggerProducer, err := f.GetLoggerProcuder()
	require.NoError(t, err)
	loggerProducer.Info().Str("action", "unit test 3").Msg("this is test")
}
