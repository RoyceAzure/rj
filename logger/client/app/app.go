package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RoyceAzure/lab/rj_kafka/pkg/config"
	kafka_consumer "github.com/RoyceAzure/lab/rj_kafka/pkg/consumer"
	"github.com/RoyceAzure/rj/logger/client/consumer"
	"github.com/RoyceAzure/rj/logger/internal/infrastructure/kafka"
)

type KafkaLoggerApp struct {
	cfg         *config.Config
	consumers   []*kafka_consumer.Consumer
	topicConfig map[string]interface{}
}

// NewKafkaLoggerApp 創建 KafkaLoggerApp
//
// cfg: kafka 配置
// elHost: elasticsearch 主機
// elPas: elasticsearch 密碼
// m: topic 配置 例如：
//
//	"cleanup.policy":      "delete",
//	"retention.ms":        "60000", // 1分鐘後自動刪除
//	"min.insync.replicas": "1",
//	"delete.retention.ms": "1000", // 標記刪除後1秒鐘清理
func NewKafkaLoggerApp(cfg *config.Config, elHost, elPas string, m map[string]interface{}) (*KafkaLoggerApp, error) {
	kafkaConsumerFactory, err := consumer.NewKafkaConsumerFactory(&consumer.LoggerConsumerConfig{
		ElUrl:       elHost,
		ElPas:       elPas,
		KafkaConfig: cfg,
	})
	if err != nil {
		return nil, err
	}

	var consumers []*kafka_consumer.Consumer
	for i := 0; i < cfg.Partition; i++ {
		consumer, err := kafkaConsumerFactory.GetLoggerConsumer()
		if err != nil {
			return nil, err
		}
		consumers = append(consumers, consumer)
	}

	return &KafkaLoggerApp{
		cfg:         cfg,
		consumers:   consumers,
		topicConfig: m,
	}, nil
}

// 設置kafka topic 並啟動所有消費者
func (k *KafkaLoggerApp) Start() error {
	err := kafka.SetupKafkaTopic(k.cfg, k.topicConfig)

	if err != nil {
		return err
	}

	for i, c := range k.consumers {
		log.Println("啟動消費者: ", i)
		c.Start()
	}

	return nil
}

// 停止所有消費者
//
// will block until all consumers are closed or timeout
func (k *KafkaLoggerApp) Stop(ctx context.Context) error {
	for _, c := range k.consumers {
		go c.Close(time.Second * 15)
	}

	allClosed := make(chan struct{})
	go func() {
		defer close(allClosed)
		for _, c := range k.consumers {
			<-c.C()
		}
	}()

	select {
	case <-allClosed:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for consumers to close")
	}
}
