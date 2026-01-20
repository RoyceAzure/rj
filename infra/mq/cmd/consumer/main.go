package main

import (
	"log"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/infra/mq/client"
)

func main() {
	err := mq.SelectConnFactory.Init(mq.MQConnParams{
		MqHost:  "localhost",
		MqUser:  "royce",
		MqPas:   "password",
		MqPort:  "5672",
		MqVHost: "/",
	})

	if err != nil {
		log.Fatal(err)
	}

	consumer, err := client.NewConsumerV2("test")
	if err != nil {
		log.Fatal(err)
	}

	consumer.Consume("local_file_logs", "test", func(msg []byte) error {
		log.Println(string(msg))
		return nil
	})
	select {}
}
