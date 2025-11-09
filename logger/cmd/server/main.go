package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/RoyceAzure/rj/logger/client/app"
	"github.com/RoyceAzure/rj/logger/config"
)

func main() {
	envConfig := config.GetConfig()
	cfg := config.GetDefaultKafkaConfig()
	cfg.Brokers = strings.Split(envConfig.KafkaBrokers, ",")
	cfg.Topic = envConfig.KafkaTopic
	cfg.ConsumerGroup = envConfig.KafkaConsumerGroup
	cfg.Partition = envConfig.KafkaPartition
	cfg.BatchSize = envConfig.KafkaBatchSize

	app, err := app.NewKafkaLoggerApp(cfg,
		fmt.Sprintf("%s:%s",
			envConfig.ElSearchHost,
			envConfig.ElSearchPort),
		envConfig.ElSearchPassword,
		map[string]interface{}{
			"cleanup.policy":      "delete",
			"retention.ms":        "60000", // 1分鐘後自動刪除
			"min.insync.replicas": "1",
			"delete.retention.ms": "1000", // 標記刪除後1秒鐘清理
		})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("app starting")
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
	log.Println("app started successfully")
	// 監聽退出訊號
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待退出訊號
	<-sigChan
	log.Println("app received shutdown signal")

	// 優雅關閉
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Stop(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	log.Printf("app closed completed")
}
