package test

import (
	"sync"
	"testing"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/infra/mq/internal/client"
	"github.com/stretchr/testify/require"
)

func TestProducer(t *testing.T) {
	err := mq.SelectConnFactory.Init(mq.MQConnParams{
		MqHost:  "localhost",
		MqUser:  "stock_ana_mq",
		MqPas:   "123456",
		MqPort:  "5672",
		MqVHost: "/",
	})
	require.NoError(t, err)

	num := 200
	prodecers := []client.IProducer{}
	msg := []byte{1, 2, 3, 4}
	for i := 0; i < num; i++ {
		temp, err := client.NewProducer()
		require.NoError(t, err)
		prodecers = append(prodecers, temp)
	}

	for i := 0; i < 20; i++ {
		for _, prodcuer := range prodecers {
			err := prodcuer.Publish("data_processer", "backtesting", msg)
			require.NoError(t, err)
		}
	}

	for _, prodcuer := range prodecers {
		prodcuer.Close()
	}

}

func TestThreadSafeProducer(t *testing.T) {
	err := mq.SelectConnFactory.Init(mq.MQConnParams{
		MqHost:  "localhost",
		MqUser:  "stock_ana_mq",
		MqPas:   "123456",
		MqPort:  "5672",
		MqVHost: "/",
	})
	require.NoError(t, err)

	num := 200
	msg := []byte{1, 2, 3, 4}
	var wg sync.WaitGroup
	errChan := make(chan error, 100)
	producer, err := client.NewThreadSafeProducer()
	producer.Start()
	require.NoError(t, err)
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 30; i++ {
				err := producer.Publish("data_processer", "backtesting", msg)
				if err != nil {
					errChan <- err
				}
			}
		}()

		// err := producer.Publish("data_processer", "backtesting", msg)
		// require.NoError(t, err)
	}
	wg.Wait()
	// producer.Close()
	require.Empty(t, errChan)
	ch := make(chan struct{})
	<-ch
}
