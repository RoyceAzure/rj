package client

import (
	"encoding/json"
	"fmt"

	"github.com/RoyceAzure/rj/infra/mq"
)

type ClientElLogger struct {
	BaseMQClientLogger
	exchange   string
	routingKey string
}

func NewClientELLogger(producer mq.IProducer, exchange, routingKey string) (*ClientElLogger, error) {
	// 加入參數驗證
	if producer == nil {
		return nil, fmt.Errorf("producer cannot be nil")
	}
	if exchange == "" {
		return nil, fmt.Errorf("exchange cannot be empty")
	}
	if routingKey == "" {
		return nil, fmt.Errorf("routingKey cannot be empty")
	}

	return &ClientElLogger{
		BaseMQClientLogger: BaseMQClientLogger{
			producer: producer,
		},
		exchange:   exchange,
		routingKey: routingKey,
	}, nil
}

func (mw *ClientElLogger) Write(p []byte) (n int, err error) {
	mqLog, err := mw.extractLogEntity(p)
	if err != nil {
		return 0, err
	}
	mqLogByte, err := json.Marshal(mqLog)
	if err != nil {
		return 0, err
	}
	err = mw.producer.Publish(mw.exchange, mw.routingKey, mqLogByte)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
