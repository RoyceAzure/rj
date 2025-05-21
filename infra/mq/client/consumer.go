package client

import (
	"fmt"

	"github.com/RoyceAzure/rj/infra/mq/constant"
	amqp "github.com/rabbitmq/amqp091-go"
)

type IConsumer interface {
	//非阻塞消費消息
	Consume(queueName string, handler func([]byte) error) error
	Close() error
	ReStart(queueName string, handler func([]byte) error) error
}

type ConsumerV2 struct {
	*BaseClient
}

func NewConsumerV2(name string) (*ConsumerV2, error) {
	consumer := &ConsumerV2{
		BaseClient: NewBaseClient(name),
	}

	err := consumer.setChanFromManger()
	if err != nil {
		return nil, err
	}

	consumer.status.Store(int32(constant.ClientInit))
	return consumer, nil
}

// 非阻塞消費消息
func (c *ConsumerV2) Consume(queueName string, handler func([]byte) error) error {
	go func() {
		for {
			msgs, err := c.setChannel(queueName)
			if err != nil {
				fmt.Printf("Consumer %s_%s, 設置channel失敗, 錯誤訊息: %v", c.name, c.id, err)
				c.close()
				return
			}
			resultCode, _ := c.consume(msgs, handler)
			if resultCode == 0 {
				fmt.Printf("Consumer %s_%s, 結束消費動作", c.name, c.id)
				return
			}
			//重置channel
			if resultCode == 1 {
				err := c.resetChannel()
				//重置channel失敗 會關閉消費者
				if err != nil {
					fmt.Printf("Consumer %s_%s, 消費訊息失敗, 錯誤訊息: %v", c.name, c.id, err)
					c.close()
					return
				}
			}
		}
	}()
	return nil
}

func (c *ConsumerV2) setChannel(queueName string) (<-chan amqp.Delivery, error) {
	if queueName == "" {
		return nil, fmt.Errorf("invalid parameters: queueName cannot be empty")
	}

	err := c.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %v", err)
	}

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
		return nil, fmt.Errorf("failed to register a consumer: %v", err)
	}
	return msgs, nil
}

// 消費訊息
//
//		return:
//	 	1 : 重置channel
//	 	0 : 結束消費程序
func (c *ConsumerV2) consume(msgs <-chan amqp.Delivery, handler func([]byte) error) (int, error) {
	c.status.Store(int32(constant.ClientRunning))
	fmt.Printf("Consumer %s_%s, 開始消費訊息", c.name, c.id)
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return 1, fmt.Errorf("consumer %s_%s, channel連線出現問題 等待回復", c.name, c.id)
			}
			// 處理訊息
			fmt.Printf("Consumer %s_%s, 接收到消息，提交給handler處理\n", c.name, c.id)
			err := handler(msg.Body)
			if err != nil {
				fmt.Printf("Consumer %s_%s, 處理消息失敗, 拒絕訊息\n", c.name, c.id)
			}
			// 確認訊息
			msg.Ack(false)

		case <-c.done:
			fmt.Printf("Consumer %s_%s, 結束消費程序\n", c.name, c.id)
			return 0, nil
		}
	}
}

// 只有當consumer處於Stop狀態時 才會執行重啟
func (c *ConsumerV2) ReStart(queueName string, handler func([]byte) error) error {
	err := c.reStart()
	if err != nil {
		return err
	}

	c.Consume(queueName, handler)
	return nil
}

func (c *ConsumerV2) Close() error {
	fmt.Printf("consumer %s_%s 開始關閉 ", c.name, c.id)
	return c.close()
}

var _ IConsumer = (*ConsumerV2)(nil)
