package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/logger/client/consumer"
	"github.com/RoyceAzure/rj/logger/config"
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

	var logFactory consumer.ILoggerConsumerFactory

	logFactory, err = consumer.NewElasticFactory(&consumer.LoggerConsumerConfig{
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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigCh
	log.Printf("收到信號: %v, 開始優雅關閉", sig)

	// 關閉資源
	if err := loggerconsumer.Close(); err != nil {
		log.Printf("關閉消費者時出錯: %v", err)
	}
	log.Printf("closed completed")
}
