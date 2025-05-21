package test

import (
	"fmt"
	"testing"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/infra/mq/client"
	"github.com/stretchr/testify/require"
)

func TestMQ(t *testing.T) {
	err := mq.SelectConnFactory.Init(mq.MQConnParams{
		MqHost:  "localhost",
		MqUser:  "royce",
		MqPas:   "password",
		MqPort:  "5672",
		MqVHost: "/",
	})
	require.NoError(t, err)

	consumer, err := client.NewConsumerV2("test")
	require.NoError(t, err)

	consumer.Consume("local_file_logs", func(msg []byte) error {
		fmt.Println(string(msg))
		return nil
	})
	select {}
}
