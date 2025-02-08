package mq

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// IProducer 定義生產者介面
type IProducer interface {
	Publish(exchange, routingKey string, message []byte) error
	Close() error
}

type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan struct{}
}

func NewProducer(conn *amqp.Connection) (*Producer, error) {
	if conn == nil {
		return nil, fmt.Errorf("connection is nil")
	}

	if conn.IsClosed() {
		return nil, fmt.Errorf("connection is closed")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	// 設置 channel 的 confirm mode
	if err := ch.Confirm(false); err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to set confirm mode: %v", err)
	}

	return &Producer{
		conn:    conn,
		channel: ch,
		done:    make(chan struct{}),
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

	// 獲取確認通道
	confirms := p.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

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
	case confirm := <-confirms:
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
	close(p.done)
	if err := p.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %v", err)
	}
	return nil
}
