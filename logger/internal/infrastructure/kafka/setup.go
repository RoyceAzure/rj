package kafka

import (
	"context"
	"time"

	"github.com/RoyceAzure/lab/rj_kafka/pkg/admin"
	"github.com/RoyceAzure/lab/rj_kafka/pkg/config"
	"github.com/segmentio/kafka-go"
)

func SetupKafkaTopic(cfg *config.Config, m map[string]interface{}) error {
	adminClient, err := admin.NewAdmin(cfg.Brokers)
	if err != nil {
		return err
	}
	defer adminClient.Close()

	err = adminClient.CreateTopic(context.Background(), admin.TopicConfig{
		Name:              cfg.Topic,
		Partitions:        cfg.Partition,
		ReplicationFactor: 3,
		Configs:           m,
	})
	if err != nil {
		return err
	}

	// 等待 topic 創建完成
	err = adminClient.WaitForTopics(context.Background(), []string{cfg.Topic}, 30*time.Second)
	if err != nil {
		return err
	}
	return nil
}

func NewKafkaWriter(cfg *config.Config) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     cfg.GetBalancer(),
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.CommitInterval + time.Millisecond*100,
		RequiredAcks: kafka.RequiredAcks(cfg.RequiredAcks),
		Compression:  kafka.Lz4,
	}
}

func NewKafkaReader(cfg *config.Config) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		Topic:       cfg.Topic,
		GroupID:     cfg.ConsumerGroup,
		StartOffset: kafka.LastOffset,
	})
}
