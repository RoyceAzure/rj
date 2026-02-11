package logger

import (
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog"
)

type ILogFactory interface {
	Create(serviceName string) zerolog.Logger
	Close()
}

// WriterStrategy 定義 Writer 策略介面（Strategy Pattern）
type WriterStrategy interface {
	CreateWriter(cfg Config) (io.Writer, io.Closer, error)
}

// Config 維持不變
type Config struct {
	Level      zerolog.Level
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	BufSize    int
	TimeFormat string
	ModuleName string
}

// DefaultConfig 維持不變
func DefaultConfig() Config {
	return Config{
		Level:      zerolog.InfoLevel,
		Filename:   "logs/app.log",
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   true,
		BufSize:    1000,
		TimeFormat: time.RFC3339,
	}
}

// LoggerType 定義 Logger 類型
type LoggerType string

const (
	LoggerTypeFile LoggerType = "file"
	LoggerTypeStd  LoggerType = "std"
)

// NewLogFactory 統一的 Factory 創建函數（Factory Pattern）
// 根據類型創建對應的 Factory
func NewLogFactory(loggerType LoggerType, cfg Config) (ILogFactory, error) {
	var strategy WriterStrategy

	switch loggerType {
	case LoggerTypeFile:
		strategy = &FileWriterStrategy{}
	case LoggerTypeStd:
		strategy = &StdWriterStrategy{}
	default:
		return nil, fmt.Errorf("unsupported logger type: %s", loggerType)
	}

	return NewBaseLogFactory(cfg, strategy)
}

// BaseLogFactory 基礎 Factory 結構（Template Method Pattern）
// 封裝共同的初始化邏輯，子類只需要實現 WriterStrategy
type BaseLogFactory struct {
	baseLogger zerolog.Logger
	baseConfig Config
	closer     io.Closer
	writer     WriterStrategy
}

// NewBaseLogFactory 創建基礎 Factory（Template Method）
func NewBaseLogFactory(cfg Config, writer WriterStrategy) (*BaseLogFactory, error) {
	// 1. 使用策略創建 Writer
	writerInstance, closer, err := writer.CreateWriter(cfg)
	if err != nil {
		return nil, err
	}

	// 2. 設定全域參數（共同邏輯）
	zerolog.SetGlobalLevel(cfg.Level)
	if cfg.TimeFormat != "" {
		zerolog.TimeFieldFormat = cfg.TimeFormat
	}

	// 3. 建立 Base Logger（共同邏輯）
	baseLogger := zerolog.New(writerInstance).
		With().
		Timestamp().
		Str("module", cfg.ModuleName).
		Logger()

	return &BaseLogFactory{
		baseLogger: baseLogger,
		baseConfig: cfg,
		closer:     closer,
		writer:     writer,
	}, nil
}

// Create 產生一個新的 Logger 實例（共同邏輯）
func (f *BaseLogFactory) Create(serviceName string) zerolog.Logger {
	return f.baseLogger.With().
		Str("func", serviceName).
		Logger()
}

// Close 關閉資源
func (f *BaseLogFactory) Close() {
	if f.closer != nil {
		f.closer.Close()
	}
}
