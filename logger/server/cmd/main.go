package main

import (
	"log"
	"path/filepath"
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

	fileLogConsumer, err := loggerconsumer.NewFileLoggerConsumer(filepath.Join(cf.LocalLogFileDir, "savefile.txt"))
	if err != nil {
		log.Fatal(err)
	}
	err = fileLogConsumer.Start("local_file_logs")
	if err != nil {
		log.Fatal(err)
	}

	var wait chan struct{}
	<-wait
}
