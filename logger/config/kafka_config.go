package config

import (
	"time"

	"github.com/RoyceAzure/lab/rj_kafka/pkg/config"
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
	cfg.Timeout = time.Second
	cfg.BatchSize = 8000
	cfg.RequiredAcks = 1
	cfg.CommitInterval = time.Millisecond * 300
	return cfg
}
