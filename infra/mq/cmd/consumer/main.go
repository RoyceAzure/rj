package main

import (
	"fmt"
	"log"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/infra/mq/internal/client"
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

	consumer.Consume("local_file_logs", func(msg []byte) error {
		fmt.Println(string(msg))
		return nil
	})
	select {}
}
