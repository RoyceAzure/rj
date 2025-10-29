// Package logconsumer 提供日誌消費者的工廠模式實現。
// 支援多種日誌後端（Elasticsearch、File 等），透過工廠模式統一建立介面。
package logger_consumer

import (
	"fmt"

	"github.com/RoyceAzure/rj/infra/elsearch"
	"github.com/olivere/elastic/v7"
)

// IFactory 定義日誌消費者工廠介面。
// 使用工廠模式的原因：
//   - 封裝複雜的初始化邏輯（如 Elasticsearch 連線設定）
//   - 支援多種日誌後端的抽換（Elasticsearch、File、Kafka 等）
//   - 統一的建立介面，便於依賴注入和測試
type IFactory interface {
	// GetLoggerConsumer 建立並返回日誌消費者實例。
	// 返回錯誤表示初始化失敗（如連線失敗、設定錯誤等）。
	GetLoggerConsumer() (ILoggerConsumer, error)
}

// LogConsumerConfig 日誌消費者的配置參數。
// 包含所有可能後端的配置，實際使用哪些欄位取決於選擇的工廠實現。
type LogConsumerConfig struct {
	// Elasticsearch 相關配置
	ElUrl     string                     // Elasticsearch 服務位址，格式：http://host:port
	ElPas     string                     // Elasticsearch 認證密碼
	ElOptions []elastic.ClientOptionFunc // Elasticsearch 客戶端額外選項（可選）

	// File 相關配置
	LogFilePath string // 日誌檔案路徑（用於 FileFactory）
}

// ElasticFactory 實現基於 Elasticsearch 的日誌消費者工廠。
//
// 職責：
//   - 驗證 Elasticsearch 連線配置
//   - 初始化 Elasticsearch 客戶端單例
//   - 建立 ElLoggerConsumer 實例
//
// 使用範例：
//
//	config := &LogConsumerConfig{
//	    ElUrl: "http://localhost:9200",
//	    ElPas: "your-password",
//	}
//	factory, err := NewElasticFactory(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	consumer, err := factory.GetLoggerConsumer()
type ElasticFactory struct {
	config *LogConsumerConfig
}

// NewElasticFactory 建立 Elasticsearch 工廠實例。
//
// 參數：
//
//	config - 必須包含 ElUrl 和 ElPas，ElOptions 為可選
//
// 返回錯誤情況：
//   - ElUrl 或 ElPas 為空字串
//
// 注意：此方法僅驗證配置，不會實際連線 Elasticsearch。
// 連線會在呼叫 GetLoggerConsumer() 時才建立。
func NewElasticFactory(config *LogConsumerConfig) (*ElasticFactory, error) {
	if config.ElUrl == "" || config.ElPas == "" {
		return nil, fmt.Errorf("elasticsearch factory requires ElUrl and ElPas")
	}

	return &ElasticFactory{
		config: config,
	}, nil
}

// GetLoggerConsumer 實現 IFactory 介面，建立 Elasticsearch 日誌消費者。
//
// 執行流程：
//  1. 初始化 Elasticsearch 單例連線
//  2. 獲取 Elasticsearch DAO 實例
//  3. 建立並返回 ElLoggerConsumer
//
// 返回錯誤情況：
//   - Elasticsearch 連線失敗
//   - 認證失敗
//   - 網路不可達
//
// 注意：此方法是執行緒安全的，但首次呼叫可能較慢（需建立連線）。
func (e *ElasticFactory) GetLoggerConsumer() (ILoggerConsumer, error) {
	// 初始化 Elasticsearch 全域單例
	err := elsearch.InitELSearch(e.config.ElUrl, e.config.ElPas, e.config.ElOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize elasticsearch: %w", err)
	}

	// 獲取 DAO 實例
	elDao, err := elsearch.GetInstance()
	if err != nil {
		return nil, fmt.Errorf("failed to get elasticsearch instance: %w", err)
	}

	// 建立並返回消費者
	return NewElLoggerConsumer(elDao)
}
