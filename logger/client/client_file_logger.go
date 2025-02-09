package client

import (
	"encoding/json"
	"fmt"

	"github.com/RoyceAzure/rj/infra/mq"
)

// 接收zerologger資料
type ClientFileLogger struct {
	BaseMQClientLogger
	exchange    string
	routingKey  string
	logFilePath string
}

func NewClientFileLogger(producer mq.IProducer, exchange, routingKey, logFilePath string) (*ClientFileLogger, error) {
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
	if logFilePath == "" {
		return nil, fmt.Errorf("logFilePath cannot be empty")
	}

	return &ClientFileLogger{
		BaseMQClientLogger: BaseMQClientLogger{
			producer: producer,
		},
		exchange:   exchange,
		routingKey: routingKey,
	}, nil
}

func (mw *ClientFileLogger) Write(p []byte) (n int, err error) {
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
