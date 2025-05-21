package client

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/RoyceAzure/rj/infra/mq/constant"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ThreadSafeProducer struct {
	*BaseClient
	chanCloseFlag atomic.Bool
	msgChan       chan publishRequest
	errChan       chan error //for publish error chan
	confirms      chan amqp.Confirmation
	channelNotify chan *amqp.Error
}

type publishRequest struct {
	exchange   string
	routingKey string
	message    []byte
}

func NewThreadSafeProducer(name string) (*ThreadSafeProducer, error) {
	producer := &ThreadSafeProducer{
		BaseClient: NewBaseClient(name),
	}

	err := producer.setChanFromManger()
	if err != nil {
		return nil, err
	}

	err = producer.setProducerExtraChannel()
	if err != nil {
		return nil, err
	}

	return producer, nil
}

func (p *ThreadSafeProducer) setProducerExtraChannel() error {
	p.confirms = p.channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	p.channelNotify = p.channel.NotifyClose(make(chan *amqp.Error, 1))
	//避免清空原有資料
	if p.msgChan == nil {
		p.msgChan = make(chan publishRequest, 100)
	}
	//避免清空原有資料
	if p.errChan == nil {
		p.errChan = make(chan error, 100)
	}

	p.chanCloseFlag.Store(false)

	return nil
}

func (p *ThreadSafeProducer) reconnect() error {
	if err := p.resetChannel(); err != nil {
		return err
	}

	if p.setProducerExtraChannel() != nil {
		return fmt.Errorf("failed to set channel")
	}

	return nil
}

// Publish 發布訊息
func (p *ThreadSafeProducer) Publish(exchange, routingKey string, message []byte) error {
	select {
	case <-p.done:
		return fmt.Errorf("producer is closed")
	default:
	}

	if exchange == "" || routingKey == "" {
		return fmt.Errorf("invalid parameters: exchange and routingKey cannot be empty")
	}

	req := publishRequest{
		exchange:   exchange,
		routingKey: routingKey,
		message:    message,
	}

	select {
	case p.msgChan <- req:
	default:
		return fmt.Errorf("producer is full")
	}

	return nil
}

func (p *ThreadSafeProducer) publish(exchange, routingKey string, message []byte) error {
	// 檢查 channel 是否已關閉
	select {
	case <-p.done:
		return fmt.Errorf("producer is closed")
	default:
	}

	// 使用 context 控制超時
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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
	case confirm := <-p.confirms:
		if !confirm.Ack {
			return fmt.Errorf("producer %s_%s failed to receive confirmation for message", p.name, p.id)
		}
	case <-ctx.Done():
		return fmt.Errorf("producer %s_%s confirmation timeout", p.name, p.id)
	}

	return nil
}

func (p *ThreadSafeProducer) Start() {
	//error handler thread
	go func() {
		for {
			select {
			case <-p.done:
				log.Printf("producer %s_%s 接收到了關閉訊號，結束error handler", p.name, p.id)
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
		p.status.Store(int32(constant.ClientRunning))
		for {
			select {
			case <-p.done:
				log.Printf("producer %s_%s 接收到了關閉訊號，結束publish thread", p.name, p.id)
				return
			case <-p.channelNotify:
				//channel不正常關閉
				if err := p.reconnect(); err != nil {
					log.Printf("producer %s_%s 重連失敗，終止程序，error: %v", p.name, p.id, err)
					p.Close()
					return
				}
			case req, ok := <-p.msgChan:
				if !ok {
					log.Printf("producer %s_%s 內部發生錯誤，終止程序", p.name, p.id)
					p.Close()
					return
				}
				p.errChan <- p.publish(req.exchange, req.routingKey, req.message)
			}
		}
	}()
}

func (p *ThreadSafeProducer) handleError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return nil
}

func (p *ThreadSafeProducer) Close() error {
	if p.status.Load() == int32(constant.ClientStop) {
		return nil
	}
	go p.close()
	return nil
}

func (p *ThreadSafeProducer) close() error {
	log.Printf("producer %s_%s 開始關閉", p.name, p.id)
	p.BaseClient.close()
	p.flush(10 * time.Second)

	if p.chanCloseFlag.Load() {
		close(p.errChan)
		close(p.msgChan)
		close(p.channelNotify)
		close(p.confirms)
	}

	p.chanCloseFlag.Store(true)
	return nil
}

// 清理剩餘msg channel訊息
func (p *ThreadSafeProducer) flush(exitDuration time.Duration) error {
	log.Printf("producer %s_%s 開始清理剩餘訊息", p.name, p.id)
	timer := time.NewTimer(exitDuration)
	select {
	case <-timer.C:
		return nil
	case req := <-p.msgChan:
		p.errChan <- p.publish(req.exchange, req.routingKey, req.message)
	}
	log.Printf("producer %s_%s 清理剩餘訊息完成", p.name, p.id)

	return nil
}

func (p *ThreadSafeProducer) ReStart() error {
	if err := p.reStart(); err != nil {
		return err
	}

	if err := p.setProducerExtraChannel(); err != nil {
		return err
	}

	p.Start()
	return nil
}

var _ IProducer = (*ThreadSafeProducer)(nil)
