package customerlogger

import (
	"encoding/json"
	"fmt"

	"github.com/RoyceAzure/rj/infra/elsearch"
	"github.com/RoyceAzure/rj/logger/server/domain/model"
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
	var logEntry model.MQLog

	err = json.Unmarshal(p, &logEntry)
	if err != nil {
		return 0, err
	}

	if logEntry.Project == "" {
		logEntry.Project = "default"
	}

	b, err := json.Marshal(logEntry)
	if err != nil {
		return 0, err
	}

	err = fw.elDao.Create(logEntry.Project, b)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (fw *ElLogger) Close() error {
	return fw.elDao.Close()
}
