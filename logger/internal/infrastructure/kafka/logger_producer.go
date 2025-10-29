package kafka

import (
	"github.com/RoyceAzure/lab/rj_kafka/kafka/config"
	"github.com/RoyceAzure/lab/rj_kafka/kafka/producer"
)

func NewKafkaProducer(cfg *config.Config, opts ...producer.Option) (*producer.ConcurrencekafkaProducer, error) {
	return producer.NewConcurrencekafkaProducer(NewKafkaWriter(cfg), *cfg, opts...)
}
