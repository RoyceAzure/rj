package loggerconsumer

import (
	"fmt"

	"github.com/RoyceAzure/rj/infra/elsearch"
	customerlogger "github.com/RoyceAzure/rj/logger/server/infrastructure/customer_logger"
	"github.com/olivere/elastic/v7"
)

type IFactory interface {
	GetLoggerConsummer() ILoggerConsumer
}

type LogConsumerConfig struct {
	ElUrl       string                     // for el
	ElPas       string                     // for el
	ElOptions   []elastic.ClientOptionFunc // for el
	LogFilePath string                     // for filelog
}

// Elasticsearch 工廠
type ElasticFactory struct {
	config *LogConsumerConfig
}

func NewElasticFactory(config *LogConsumerConfig) (*ElasticFactory, error) {
	if config.ElUrl == "" || config.ElPas == "" {
		return nil, fmt.Errorf("must have elurl and elpas")
	}

	return &ElasticFactory{
		config: config,
	}, nil
}

func (e *ElasticFactory) GetLoggerConsumer() (ILoggerConsumer, error) {
	err := elsearch.InitELSearch(e.config.ElUrl, e.config.ElPas, e.config.ElOptions...)
	if err != nil {
		return nil, err
	}

	elDao, err := elsearch.GetInstance()
	if err != nil {
		return nil, err
	}

	elLogger := customerlogger.NewElLogger(elDao)
	return nil, nil
}
