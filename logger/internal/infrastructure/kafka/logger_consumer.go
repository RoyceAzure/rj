package kafka

import (
	"time"

	"github.com/RoyceAzure/lab/rj_kafka/pkg/config"
	"github.com/RoyceAzure/lab/rj_kafka/pkg/consumer"
	"github.com/RoyceAzure/rj/infra/elsearch"
)

func NewKafkaElConsumer(cfg *config.Config, elDao elsearch.IElSearchDao, opts ...consumer.Option) (*consumer.Consumer, error) {
	processer, err := consumer.NewKafkaElProcesser(elDao, time.Millisecond*200)
	if err != nil {
		return nil, err
	}

	return consumer.NewConsumer(NewKafkaReader(cfg), processer, *cfg, opts...), nil
}
