package logger

import (
	"encoding/json"
	"fmt"

	"github.com/RoyceAzure/rj/logger/internal/model"
	"github.com/RoyceAzure/rj/repo/file"
)

// for consumer  不需要接收來自zerolog的資訊
type FileLogger struct {
	fileDao file.FileDAO
}

func NewFileLogger(fileDao file.FileDAO) *FileLogger {
	return &FileLogger{
		fileDao: fileDao,
	}
}

func (fw *FileLogger) Write(p []byte) (n int, err error) {
	if fw == nil {
		return 0, fmt.Errorf("file logger is not init")
	}
	var logEntry model.MQLog

	err = json.Unmarshal(p, &logEntry)
	if err != nil {
		return 0, err
	}
	err = fw.fileDao.Append(logEntry.String())
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (fw *FileLogger) Close() error {
	return fw.fileDao.Close()
}
