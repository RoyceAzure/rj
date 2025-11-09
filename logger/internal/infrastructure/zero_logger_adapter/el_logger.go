package zero_logger_adapter

import (
	"encoding/json"
	"fmt"

	"github.com/RoyceAzure/rj/infra/elsearch"
)

type ElLogger struct {
	elDao elsearch.IElSearchDao
}

func NewElLogger(elDao elsearch.IElSearchDao) *ElLogger {
	return &ElLogger{
		elDao: elDao,
	}
}

func (fw *ElLogger) Write(p []byte) (n int, err error) {
	if fw == nil {
		return 0, fmt.Errorf("file logger is not init")
	}
	var logEntry map[string]any

	copyP := make([]byte, len(p))
	copy(copyP, p) //避免zero logger在寫入時的競爭問題

	err = json.Unmarshal(copyP, &logEntry)
	if err != nil {
		return 0, err
	}

	if logEntry["project"] == "" {
		logEntry["project"] = "default"
	}

	b, err := json.Marshal(logEntry)
	if err != nil {
		return 0, err
	}

	err = fw.elDao.Create(logEntry["project"].(string), b)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (fw *ElLogger) Close() error {
	return fw.elDao.Close()
}
