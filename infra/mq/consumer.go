package mq

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type IConsumer interface {
	Consume(queueName string, handler func([]byte) error) error
	Close() error
}

type ConsumerV2 struct {
	name      string
	channel   *amqp.Channel
	channelId int
	done      chan struct{}
}

func NewConsumerV2(name string) (*ConsumerV2, error) {
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
	return &ConsumerV2{
		name:      name,
		channel:   channel,
		channelId: channelId,
		done:      make(chan struct{}),
	}, nil
}

// Consume 開始消費訊息
func (c *ConsumerV2) Consume(queueName string, handler func([]byte) error) error {
	// 設置QoS
	if queueName == "" {
		return fmt.Errorf("invalid parameters: queueName cannot be empty")
	}

	err := c.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %v", err)
	}

	fmt.Printf("Consumer %s啟動\n", c.name)

	// 開始消費訊息
	msgs, err := c.channel.Consume(
		queueName, // 佇列名稱
		"",        // 消費者標籤
		false,     // 自動確認
		false,     // 獨佔
		false,     // no-local
		false,     // no-wait
		nil,       // 參數
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %v", err)
	}

	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					fmt.Printf("Consumer %s, channel連線出現問題 等待回復\n", c.name)
					time.Sleep(time.Second * 5)
					continue
				}
				// 處理訊息
				fmt.Printf("Consumer %s, 接收到消息，提交給handler處理\n", c.name)
				err := handler(msg.Body)
				if err != nil {
					fmt.Printf("Consumer %s, 處理消息失敗, 拒絕訊息\n", err)
					continue
				}
				// 確認訊息
				msg.Ack(false)
			case <-c.done:
				fmt.Printf("Consumer %s, 結束消費程序\n", c.name)
				return
			}
		}
	}()

	return nil
}

// Close 關閉消費者
func (c *ConsumerV2) Close() error {
	fmt.Print("開始關閉consumer")
	close(c.done)
	ma, err := SelectConnFactory.GetManager()
	if err != nil {
		return err
	}

	return ma.ReleaseChannel(c.channelId)
}
