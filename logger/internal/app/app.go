package app

import (
	"context"
	"fmt"
	"time"

	"github.com/RoyceAzure/lab/rj_kafka/kafka/config"
	"github.com/RoyceAzure/lab/rj_kafka/kafka/consumer"
	"github.com/RoyceAzure/rj/infra/elsearch"
	"github.com/RoyceAzure/rj/logger/internal/infrastructure/kafka"
)

type KafkaLoggerApp struct {
	cfg       *config.Config
	consumers []*consumer.Consumer
	elDao     elsearch.IElSearchDao
}

func NewKafkaLoggerApp(cfg *config.Config, elHost, elPas string) (*KafkaLoggerApp, error) {
	elDao, err := setUpElDao(elHost, elPas)
	if err != nil {
		return nil, err
	}
	return &KafkaLoggerApp{
		cfg:       cfg,
		consumers: make([]*consumer.Consumer, 0, cfg.Partition),
		elDao:     elDao,
	}, nil
}

func setUpElDao(host, pas string) (*elsearch.ElSearchDao, error) {
	// 初始化 ES
	err := elsearch.InitELSearch(host, pas)
	if err != nil {
		return nil, err
	}

	dao, err := elsearch.GetInstance()
	if err != nil {
		return nil, err
	}
	return dao, nil
}

func (k *KafkaLoggerApp) Start() error {
	err := kafka.SetupKafkaTopic(k.cfg)
	if err != nil {
		return err
	}

	for i := 0; i < k.cfg.Partition; i++ {
		c, err := kafka.NewKafkaElConsumer(k.cfg, k.elDao)
		if err != nil {
			return err
		}
		c.Start()
		k.consumers = append(k.consumers, c)
	}
	return nil
}

// 停止所有消費者
//
// will block until all consumers are closed or timeout
func (k *KafkaLoggerApp) Stop(ctx context.Context) error {
	for _, c := range k.consumers {
		c.Close(time.Second * 10)
	}

	allClosed := make(chan struct{})
	go func() {
		defer close(allClosed)
		for _, c := range k.consumers {
			<-c.C()
		}
	}()

	select {
	case <-allClosed:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for consumers to close")
	}
}
