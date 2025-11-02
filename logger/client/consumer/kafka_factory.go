package consumer

import (
	"fmt"

	"github.com/RoyceAzure/lab/rj_kafka/pkg/consumer"
	"github.com/RoyceAzure/rj/infra/elsearch"
	"github.com/RoyceAzure/rj/logger/internal/infrastructure/kafka"
)

// KafkaFactory 工廠
type KafkaConsumerFactory struct {
	config *LoggerConsumerConfig
}

func NewKafkaConsumerFactory(config *LoggerConsumerConfig) (*KafkaConsumerFactory, error) {
	if config.ElUrl == "" || config.ElPas == "" {
		return nil, fmt.Errorf("must have elurl and elpas")
	}

	return &KafkaConsumerFactory{
		config: config,
	}, nil
}

func (k *KafkaConsumerFactory) GetLoggerConsumer() (*consumer.Consumer, error) {
	err := elsearch.InitELSearch(k.config.ElUrl, k.config.ElPas, k.config.ElOptions...)
	if err != nil {
		return nil, err
	}

	elDao, err := elsearch.GetInstance()
	if err != nil {
		return nil, err
	}

	return kafka.NewKafkaElConsumer(k.config.KafkaConfig, elDao)
}
