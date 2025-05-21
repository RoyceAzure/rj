package client

import (
	"fmt"
	"time"

	"github.com/RoyceAzure/rj/infra/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// IProducer 定義生產者介面
type IProducer interface {
	Publish(exchange, routingKey string, message []byte) error
	Close() error
}

type Producer struct {
	channel  *amqp.Channel
	done     chan struct{}
	confirms chan amqp.Confirmation
}

func NewProducer() (*Producer, error) {
	ma, err := mq.SelectConnFactory.GetManager()
	if err != nil {
		return nil, err
	}

	channel, err := ma.GetChannel()
	if err != nil {
		return nil, err
	}

	if channel.IsClosed() {
		return nil, fmt.Errorf("channel is closed")
	}

	if err := channel.Confirm(false); err != nil {
		channel.Close()
		return nil, fmt.Errorf("failed to set confirm mode: %v", err)
	}

	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	return &Producer{
		channel:  channel,
		done:     make(chan struct{}),
		confirms: confirms,
	}, nil
}

// Publish 發布訊息
func (p *Producer) Publish(exchange, routingKey string, message []byte) error {
	if exchange == "" || routingKey == "" {
		return fmt.Errorf("invalid parameters: exchange and routingKey cannot be empty")
	}

	// 檢查 channel 是否已關閉
	select {
	case <-p.done:
		return fmt.Errorf("producer is closed")
	default:
	}

	err := p.channel.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	// 等待發布確認
	select {
	case confirm := <-p.confirms:
		if !confirm.Ack {
			return fmt.Errorf("failed to receive confirmation for message")
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("confirmation timeout")
	}

	return nil
}

// Close 關閉生產者
func (p *Producer) Close() error {
	if p.done != nil {
		close(p.done)
	}
	return nil
}
