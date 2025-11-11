package producer

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/RoyceAzure/lab/rj_kafka/pkg/config"
	"github.com/RoyceAzure/rj/infra/mq/client"
	"github.com/RoyceAzure/rj/logger/internal/infrastructure/mq/logger_producer"
	"github.com/RoyceAzure/rj/logger/internal/infrastructure/zero_logger_adapter"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

var globalLogger *zerolog.Logger

func init() {
	l := zerolog.New(os.Stdout).With().Timestamp().Logger()
	globalLogger = &l
}

type LoggerProducerConfig struct {
	Exchange        string         `validate:"required" json:"exchange"`    // required 表示必須有值
	RoutingKey      string         `validate:"required" json:"routing_key"` // required 表示必須有值
	Module          string         `json:"module"`                          // 非必填
	Project         string         `json:"project"`                         // 非必填
	LogFileSavePath string         //for file logger
	KafkaConfig     *config.Config //for kafka logger
	IsOutPutStd     bool           // for stdout
}

func (p *LoggerProducerConfig) Validate() error {
	validate := validator.New()

	if err := validate.Struct(p); err != nil {
		// 處理驗證錯誤
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			for _, e := range validationErrors {
				switch e.Field() {
				case "Exchange":
					return fmt.Errorf("exchange is required")
				case "RoutingKey":
					return fmt.Errorf("routing key is required")
				}
			}
		}
		return err
	}
	return nil
}

type ILoggerProducerFactory interface {
	GetLoggerProcuder() (*zerolog.Logger, error)
}

var (
	mqFileLogger *zerolog.Logger
)

type FileLoggerFactory struct {
	config *LoggerProducerConfig
}

func NewFileLoggerFactory(config *LoggerProducerConfig) (*FileLoggerFactory, error) {
	return &FileLoggerFactory{
		config: config,
	}, nil
}

func (e *FileLoggerFactory) GetLoggerProcuder() (*zerolog.Logger, error) {
	if mqElLogger != nil {
		return mqElLogger, nil
	}

	logger, err := e.createFileLogger()
	if err != nil {
		return nil, err
	}

	mqFileLogger = logger
	return mqElLogger, nil
}

func (e *FileLoggerFactory) createFileLogger() (*zerolog.Logger, error) {
	producer, err := client.NewThreadSafeProducer("file_logger_producer")
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	logger, err := logger_producer.NewClientFileLogger(producer, e.config.Exchange, e.config.RoutingKey, e.config.LogFileSavePath)
	if err != nil {
		producer.Close() // 清理資源
		return nil, fmt.Errorf("failed to create client file logger: %w", err)
	}
	producer.Start()

	zeroLogger := setUpClientLToZeroL(logger, e.config)
	return &zeroLogger, nil
}

var (
	mqElLogger *zerolog.Logger
)

// Elasticsearch 工廠
type ElasticFactory struct {
	config *LoggerProducerConfig
}

func NewElasticFactory(config *LoggerProducerConfig) (*ElasticFactory, error) {
	return &ElasticFactory{
		config: config,
	}, nil
}

func (e *ElasticFactory) GetLoggerProcuder() (*zerolog.Logger, error) {
	if mqElLogger != nil {
		return mqElLogger, nil
	}

	logger, err := e.createLoggerProcuder()
	if err != nil {
		return nil, err
	}

	mqElLogger = logger
	return mqElLogger, nil
}

func (e *ElasticFactory) createLoggerProcuder() (*zerolog.Logger, error) {
	producer, err := client.NewThreadSafeProducer("el_logger_producer")
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	el_logger, err := logger_producer.NewClientELLogger(producer, e.config.Exchange, e.config.RoutingKey)
	if err != nil {
		producer.Close() // 清理資源
		return nil, fmt.Errorf("failed to create client el logger: %w", err)
	}
	producer.Start()

	zeroLogger := setUpClientLToZeroL(el_logger, e.config)
	return &zeroLogger, nil
}

var (
	kafkaLogger *zerolog.Logger
	kmu         sync.RWMutex
)

// Kafka 工廠
type KafkaLoggerFactory struct {
	config *LoggerProducerConfig
}

func NewKafkaLoggerFactory(config *LoggerProducerConfig) (*KafkaLoggerFactory, error) {
	return &KafkaLoggerFactory{
		config: config,
	}, nil
}

func (k *KafkaLoggerFactory) GetLoggerProcuder() (*zerolog.Logger, error) {
	kmu.RLock()
	if kafkaLogger != nil {
		return kafkaLogger, nil
	}
	kmu.RUnlock()

	kmu.Lock()
	if kafkaLoggerAdapter != nil {
		kmu.Unlock()
		return kafkaLogger, nil
	}
	logger, err := k.createLoggerProcuder()
	if err != nil {
		return nil, err
	}

	kafkaLogger = logger
	kmu.Unlock()
	return kafkaLogger, nil
}

func (k *KafkaLoggerFactory) createLoggerProcuder() (*zerolog.Logger, error) {
	kafkaLoggerWriter, err := zero_logger_adapter.NewKafkaLoggerWriter(k.config.KafkaConfig)
	if err != nil {
		kafkaLoggerWriter.Close()
		return nil, fmt.Errorf("failed to create kafka logger: %w", err)
	}

	zeroLogger := setUpClientLToZeroL(kafkaLoggerWriter, k.config)

	return &zeroLogger, nil
}

var (
	kafkaLoggerAdapter *KafkaLoggerAdapter
	kamu               sync.RWMutex
)

// Kafka 工廠
type KafkaLoggerAdapterFactory struct {
	config *LoggerProducerConfig
}

// 紀錄zerolog logger的錯誤次數
type KafkaLoggerAdapter struct {
	*zerolog.Logger                                        //muti logger
	kafkaLogger     *zero_logger_adapter.KafkaLoggerWriter //紀錄指標，用來取得error count
}

func NewKafkaLoggerAdapter(l *zerolog.Logger, kafkaLogger *zero_logger_adapter.KafkaLoggerWriter) *KafkaLoggerAdapter {
	return &KafkaLoggerAdapter{
		Logger:      l,
		kafkaLogger: kafkaLogger,
	}
}

func (k *KafkaLoggerAdapter) GetErrorCount() int64 {
	return k.kafkaLogger.GetErrorCount()
}

func NewKafkaLoggerAdapterFactory(config *LoggerProducerConfig) (*KafkaLoggerAdapterFactory, error) {
	return &KafkaLoggerAdapterFactory{
		config: config,
	}, nil
}
func (k *KafkaLoggerAdapterFactory) GetLoggerProcuder() (*KafkaLoggerAdapter, error) {
	kamu.RLock()
	if kafkaLoggerAdapter != nil {
		return kafkaLoggerAdapter, nil
	}
	kamu.RUnlock()

	kamu.Lock()
	if kafkaLoggerAdapter != nil {
		kamu.Unlock()
		return kafkaLoggerAdapter, nil
	}
	logger, err := k.createLoggerProcuder()
	if err != nil {
		return nil, err
	}

	kafkaLoggerAdapter = logger
	kamu.Unlock()
	return kafkaLoggerAdapter, nil
}

func (k *KafkaLoggerAdapterFactory) createLoggerProcuder() (*KafkaLoggerAdapter, error) {
	kafkaLoggerWriter, err := zero_logger_adapter.NewKafkaLoggerWriter(k.config.KafkaConfig)
	if err != nil {
		kafkaLoggerWriter.Close()
		return nil, fmt.Errorf("failed to create kafka logger: %w", err)
	}

	zeroLogger := setUpClientLToZeroL(kafkaLoggerWriter, k.config)

	return NewKafkaLoggerAdapter(&zeroLogger, kafkaLoggerWriter), nil
}

// ILoggerProcuder 轉換成zero logger
func setUpClientLToZeroL(logger logger_producer.ILoggerProcuder, config *LoggerProducerConfig) zerolog.Logger {
	var multiLogger zerolog.LevelWriter
	if config.IsOutPutStd {
		multiLogger = zerolog.MultiLevelWriter(
			zerolog.ConsoleWriter{Out: os.Stdout},
			logger,
		)
	} else {
		multiLogger = zerolog.MultiLevelWriter(
			logger,
		)
	}

	l := zerolog.New(multiLogger).With()

	if config.Project != "" {
		l = l.Str("project", config.Project)
	}

	if config.Module != "" {
		l = l.Str("module", config.Module)
	}

	return l.
		Timestamp().
		Logger()
}
