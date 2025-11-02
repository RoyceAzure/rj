package kafka

import (
	"context"
	"time"

	"github.com/RoyceAzure/lab/rj_kafka/pkg/admin"
	"github.com/RoyceAzure/lab/rj_kafka/pkg/config"
	"github.com/segmentio/kafka-go"
)

// GetDefaultKafkaConfig 取得 Kafka Producer、Consumer 的預設配置。
// 同一份配置可以同時用於 創建kafka topic、Producer 和 Consumer。
//
// 職責：
//   - 提供 Kafka 生產者和消費者的通用配置參數
//   - 設定分區數量、超時時間、批次大小等參數
//   - 配置適合高吞吐量場景的確認機制
//
// 配置說明：
//   - Partition: 6 個分區，提高並行處理能力
//   - Timeout: 1 秒超時時間
//   - BatchSize: 10000 筆批次大小
//   - RequiredAcks: 0，不等待確認，提高寫入效能
//   - CommitInterval: 500 毫秒提交間隔
//
// 使用範例：
//
//	cfg := GetDefaultKafkaConfig()
//	// 可根據需求調整配置
//	cfg.Partition = 3
func GetDefaultKafkaConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.Partition = 6
	cfg.Timeout = time.Second
	cfg.BatchSize = 10000
	cfg.RequiredAcks = 0 //writer不需要等待確認
	cfg.CommitInterval = time.Millisecond * 500
	return cfg
}

func SetupKafkaTopic(cfg *config.Config) error {
	adminClient, err := admin.NewAdmin(cfg.Brokers)
	if err != nil {
		return err
	}
	defer adminClient.Close()

	err = adminClient.CreateTopic(context.Background(), admin.TopicConfig{
		Name:              cfg.Topic,
		Partitions:        cfg.Partition,
		ReplicationFactor: 3,
		Configs: map[string]interface{}{
			"cleanup.policy":      "delete",
			"retention.ms":        "60000", // 1分鐘後自動刪除
			"min.insync.replicas": "2",
			"delete.retention.ms": "1000", // 標記刪除後1秒鐘清理
		},
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
