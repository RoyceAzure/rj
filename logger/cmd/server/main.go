package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RoyceAzure/rj/logger/internal/app"
	"github.com/RoyceAzure/rj/logger/internal/infrastructure/kafka"
)

func main() {
	cfg := kafka.GetDefaultKafkaConfig()
	app, err := app.NewKafkaLoggerApp(cfg, "http://localhost:9200", "password")
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

	// 監聽退出訊號
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待退出訊號
	<-sigChan
	log.Println("Received shutdown signal")

	// 優雅關閉
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Stop(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	log.Printf("closed completed")
}
