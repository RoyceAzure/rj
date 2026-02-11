package logger

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"gopkg.in/natefinch/lumberjack.v2"
)

type FileWriterStrategy struct{}

func (s *FileWriterStrategy) CreateWriter(cfg Config) (io.Writer, io.Closer, error) {
	fileWriter := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	asyncWriter := diode.NewWriter(fileWriter, cfg.BufSize, 10*time.Millisecond, func(missed int) {
		fmt.Printf("Logger Dropped %d messages\n", missed)
	})

	return asyncWriter, asyncWriter, nil
}

type FileLogFactory struct {
	*BaseLogFactory
}

var (
	fileLogGlobalFactory     *FileLogFactory
	fileLogGlobalFactoryOnce sync.Once
)

func InitFileLogGlobal(cfg Config) {
	fileLogGlobalFactoryOnce.Do(func() {
		baseFactory, err := NewBaseLogFactory(cfg, &FileWriterStrategy{})
		if err != nil {
			panic(fmt.Sprintf("failed to create file log factory: %v", err))
		}
		fileLogGlobalFactory = &FileLogFactory{BaseLogFactory: baseFactory}
	})
}

func GetFileLogFactory() *FileLogFactory {
	if fileLogGlobalFactory == nil {
		// 防呆：如果使用者忘記 Init，給一個預設的，或是 Panic
		InitFileLogGlobal(DefaultConfig())
	}
	return fileLogGlobalFactory
}

func GetFileLogLogger(serviceName string) zerolog.Logger {
	return GetFileLogFactory().Create(serviceName)
}
