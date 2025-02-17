package tests

import (
	"testing"

	"github.com/RoyceAzure/rj/infra/mq"
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

	num := 10
	prodecers := []mq.IProducer{}
	msg := []byte{1, 2, 3, 4}
	for i := 0; i < num; i++ {
		temp, err := mq.NewProducer()
		require.NoError(t, err)
		prodecers = append(prodecers, temp)
	}

	for i := 0; i < 4; i++ {
		for _, prodcuer := range prodecers {
			prodcuer.Publish("data_processer", "backtesting", msg)
		}
	}

	for _, prodcuer := range prodecers {
		prodcuer.Close()
	}

}
