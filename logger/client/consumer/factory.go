package consumer

import (
	"fmt"

	"github.com/RoyceAzure/rj/infra/elsearch"
	"github.com/RoyceAzure/rj/logger/internal/infrastructure/mq/logger_consumer"
	"github.com/olivere/elastic/v7"
)

type ILoggerConsumerFactory interface {
	GetLoggerConsumer() (logger_consumer.ILoggerConsumer, error)
}

type LoggerConsumerConfig struct {
	ElUrl       string                     // for el
	ElPas       string                     // for el
	ElOptions   []elastic.ClientOptionFunc // for el
	LogFilePath string                     // for filelog
}

// Elasticsearch 工廠
type ElasticFactory struct {
	config *LoggerConsumerConfig
}

func NewElasticFactory(config *LoggerConsumerConfig) (*ElasticFactory, error) {
	if config.ElUrl == "" || config.ElPas == "" {
		return nil, fmt.Errorf("must have elurl and elpas")
	}

	return &ElasticFactory{
		config: config,
	}, nil
}

func (e *ElasticFactory) GetLoggerConsumer() (logger_consumer.ILoggerConsumer, error) {
	err := elsearch.InitELSearch(e.config.ElUrl, e.config.ElPas, e.config.ElOptions...)
	if err != nil {
		return nil, err
	}

	elDao, err := elsearch.GetInstance()
	if err != nil {
		return nil, err
	}

	return logger_consumer.NewElLoggerConsumer(elDao)
}
