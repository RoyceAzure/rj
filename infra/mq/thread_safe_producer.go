package mq

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ThreadSafeProducer struct {
	channel   *amqp.Channel
	channelId int
	msgChan   chan publishRequest
	errChan   chan error //for publish error chan
	done      chan struct{}
}

type publishRequest struct {
	exchange   string
	routingKey string
	message    []byte
}

func NewThreadSafeProducer() (*ThreadSafeProducer, error) {
	ma, err := SelectConnFactory.GetManager()
	if err != nil {
		return nil, err
	}

	channelId, channel, err := ma.RegisterChannel()
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

	return &ThreadSafeProducer{
		channel:   channel,
		channelId: channelId,
		msgChan:   make(chan publishRequest, 100),
		errChan:   make(chan error, 100),
		done:      make(chan struct{}),
	}, nil
}

// Publish 發布訊息
func (p *ThreadSafeProducer) Publish(exchange, routingKey string, message []byte) error {
	if exchange == "" || routingKey == "" {
		return fmt.Errorf("invalid parameters: exchange and routingKey cannot be empty")
	}

	req := publishRequest{
		exchange:   exchange,
		routingKey: routingKey,
		message:    message,
	}

	p.msgChan <- req

	return nil
}

func (p *ThreadSafeProducer) publish(exchange, routingKey string, message []byte) error {
	// 檢查 channel 是否已關閉
	select {
	case <-p.done:
		return fmt.Errorf("producer is closed")
	default:
	}

	// 獲取確認通道
	confirms := p.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	// 使用 context 控制超時
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := p.channel.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
			Timestamp:   time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return fmt.Errorf("failed to receive confirmation for message")
		}
	case <-ctx.Done():
		return fmt.Errorf("confirmation timeout")
	}

	return nil
}

func (p *ThreadSafeProducer) Start() {
	//error handler thread
	go func() {
		for {
			select {
			case <-p.done:
				log.Printf("thread safe producer is closed!!")
				return
			case err, ok := <-p.errChan:
				if !ok {
					return
				}
				p.handleError(err)
			}
		}
	}()

	//publish thread
	go func() {
		for {
			select {
			case <-p.done:
				log.Printf("thread safe producer is closed!!")
				return
			case req, ok := <-p.msgChan:
				if !ok {
					return
				}
				p.errChan <- p.publish(req.exchange, req.routingKey, req.message)
			}
		}
	}()
}

func (p *ThreadSafeProducer) handleError(err error) error {
	log.Print(err)
	return nil
}

func (p *ThreadSafeProducer) Close() error {
	close(p.done)
	ma, err := SelectConnFactory.GetManager()
	if err != nil {
		return err
	}

	err = ma.ReleaseChannel(p.channelId)

	close(p.errChan)
	close(p.msgChan)
	return err
}
