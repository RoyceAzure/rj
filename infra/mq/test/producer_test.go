package test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/infra/mq/client"
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

func printIntergers(producer *client.ThreadSafeProducer, intergers []int) {
	for _, i := range intergers {
		msg := fmt.Sprintf("this is test %d", i)
		producer.Publish("system_logs", "log.file.el", []byte(msg))
		time.Sleep(1 * time.Second)
	}
}
func TestThreadSafeProducer_1(t *testing.T) {
	err := mq.SelectConnFactory.Init(mq.MQConnParams{
		MqHost:  "localhost",
		MqUser:  "royce",
		MqPas:   "password",
		MqPort:  "5672",
		MqVHost: "/",
	})
	require.NoError(t, err)

	numProducer := 3
	errChan := make(chan error, 100)
	require.NoError(t, err)
	intergers := [][]int{
		{1, 3, 5, 7, 9, 11},
		{2, 4, 6, 8, 10, 12},
		{13, 14, 15, 16, 17, 18},
	}

	wg := sync.WaitGroup{}
	for i := 0; i < numProducer; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			producerNmae := fmt.Sprintf("test_producer_%d", i)
			producer, err := client.NewThreadSafeProducer(producerNmae)
			if err != nil {
				return
			}
			producer.Start()
			printIntergers(producer, intergers[i])
			producer.Close()
		}(i)

	}

	wg.Wait()
	require.Empty(t, errChan)
}

func TestThreadSafeProducer(t *testing.T) {
	err := mq.SelectConnFactory.Init(mq.MQConnParams{
		MqHost:  "localhost",
		MqUser:  "royce",
		MqPas:   "password",
		MqPort:  "5672",
		MqVHost: "/",
	})
	require.NoError(t, err)

	numProducer := 3
	numMsg := 100
	errChan := make(chan error, 100)
	producers := make([]*client.ThreadSafeProducer, numProducer)
	require.NoError(t, err)
	for i := 0; i < numProducer; i++ {
		producerNmae := fmt.Sprintf("test_producer_%d_", i)
		producer, err := client.NewThreadSafeProducer(producerNmae)
		if err != nil {
			return
		}
		producers[i] = producer
		producer.Start()
	}

	for i := 0; i < numMsg; i++ {
		msg := fmt.Sprintf("this is test %d", i)
		index := i % numProducer
		producers[index].Publish("system_logs", "log.file.el", []byte(msg))
		time.Sleep(1 * time.Second)
	}

	for _, producer := range producers {
		producer.Close()
	}
	time.Sleep(10 * time.Second)

	require.Empty(t, errChan)
}

func TestThreadSafeProducerClose(t *testing.T) {
	err := mq.SelectConnFactory.Init(mq.MQConnParams{
		MqHost:  "localhost",
		MqUser:  "royce",
		MqPas:   "password",
		MqPort:  "5672",
		MqVHost: "/",
	})
	require.NoError(t, err)

	numProducer := 3
	numMsg := 100
	errChan := make(chan error, 100)
	producers := make([]*client.ThreadSafeProducer, numProducer)
	require.NoError(t, err)
	for i := 0; i < numProducer; i++ {
		producerNmae := fmt.Sprintf("test_producer_%d", i)
		producer, err := client.NewThreadSafeProducer(producerNmae)
		if err != nil {
			return
		}
		producers[i] = producer
		producer.Start()
	}

	go func() {
		for i := 0; i < numMsg; i++ {
			msg := fmt.Sprintf("this is test %d", i)
			producers[i%numProducer].Publish("system_logs", "log.file.el", []byte(msg))
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	time.Sleep(5 * time.Second)
	for _, producer := range producers {
		producer.Close()
	}

	time.Sleep(10 * time.Second)
	require.Empty(t, errChan)
}

func TestThreadSafeProducerReStart(t *testing.T) {
	err := mq.SelectConnFactory.Init(mq.MQConnParams{
		MqHost:  "localhost",
		MqUser:  "royce",
		MqPas:   "password",
		MqPort:  "5672",
		MqVHost: "/",
	})
	require.NoError(t, err)

	numProducer := 3
	numMsg := 100
	errChan := make(chan error, 100)
	producers := make([]*client.ThreadSafeProducer, numProducer)
	require.NoError(t, err)
	for i := 0; i < numProducer; i++ {
		producerNmae := fmt.Sprintf("test_producer_%d", i)
		producer, err := client.NewThreadSafeProducer(producerNmae)
		if err != nil {
			return
		}
		producers[i] = producer
		producer.Start()
	}

	go func() {
		for i := 0; i < numMsg; i++ {
			msg := fmt.Sprintf("this is test %d", i)
			producers[i%numProducer].Publish("system_logs", "log.file.el", []byte(msg))
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	time.Sleep(5 * time.Second)
	for _, producer := range producers {
		producer.Close()
	}

	time.Sleep(10 * time.Second)
	for _, producer := range producers {
		producer.ReStart()
	}
	time.Sleep(10 * time.Second)
	require.Empty(t, errChan)
}
