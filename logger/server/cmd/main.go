package main

import (
	"fmt"
	"log"
	"time"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/logger/server/config"
	loggerconsumer "github.com/RoyceAzure/rj/logger/server/service/logger_consumer"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339

	cf := config.GetConfig()

	err := mq.SelectConnFactory.Init(mq.MQConnParams{
		MqHost:  cf.MqHost,
		MqUser:  cf.MqUser,
		MqPas:   cf.MqPassword,
		MqPort:  cf.MqPort,
		MqVHost: cf.MqVHost,
	})
	if err != nil {
		log.Fatal(err)
	}

	var logFactory loggerconsumer.IFactory

	logFactory, err = loggerconsumer.NewElasticFactory(&loggerconsumer.LogConsumerConfig{
		ElUrl: fmt.Sprintf("%s:%s", cf.ElSearchHost, cf.ElSearchPort),
		ElPas: cf.ElSearchPassword,
	})
	if err != nil {
		log.Fatal(err)
	}

	loggerconsumer, err := logFactory.GetLoggerConsumer()
	if err != nil {
		log.Fatal(err)
	}

	err = loggerconsumer.Start("local_file_logs")
	if err != nil {
		log.Fatal(err)
	}

	var wait chan struct{}
	<-wait
}
