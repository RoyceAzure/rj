package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

type StdWriterStrategy struct{}

func (s *StdWriterStrategy) CreateWriter(cfg Config) (io.Writer, io.Closer, error) {
	asyncWriter := diode.NewWriter(os.Stdout, cfg.BufSize, 10*time.Millisecond, func(missed int) {
		fmt.Printf("Logger Dropped %d messages\n", missed)
	})

	return asyncWriter, asyncWriter, nil
}

// StdLogFactory 標準輸出 Logger Factory（使用組合而非繼承）
type StdLogFactory struct {
	*BaseLogFactory
}

// 1. 定義全域變數 (不導出，強迫透過 Get 存取)
var (
	stdLogGlobalFactory     *StdLogFactory
	stdLogGlobalFactoryOnce sync.Once
)

func InitStdLogGlobal(cfg Config) {
	stdLogGlobalFactoryOnce.Do(func() {
		baseFactory, err := NewBaseLogFactory(cfg, &StdWriterStrategy{})
		if err != nil {
			panic(fmt.Sprintf("failed to create std log factory: %v", err))
		}
		stdLogGlobalFactory = &StdLogFactory{BaseLogFactory: baseFactory}
	})
}

func GetStdLogFactory() *StdLogFactory {
	if stdLogGlobalFactory == nil {
		InitStdLogGlobal(DefaultConfig())
	}
	return stdLogGlobalFactory
}

func GetStdLogLogger(serviceName string) zerolog.Logger {
	return GetStdLogFactory().Create(serviceName)
}
