package mq

import (
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ConnectionSingleTon struct {
	conn      *amqp.Connection
	url       string
	mu        sync.Mutex
	closeChan chan *amqp.Error
}

var (
	instance *ConnectionSingleTon
	once     sync.Once
)

// GetInstance 返回單例連線物件
func GetInstance(url string) *ConnectionSingleTon {
	once.Do(func() {
		instance = &ConnectionSingleTon{url: url}
	})
	return instance
}

// GetConnection 取得連線，如果沒有則建立
func (c *ConnectionSingleTon) GetConnection() (*amqp.Connection, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 檢查現有連線
	if c.conn != nil && !c.conn.IsClosed() {
		return c.conn, nil
	}

	// 需要建立新連線
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	// 設定自動重連
	go c.handleReconnect()

	c.conn = conn
	return conn, nil
}

func (c *ConnectionSingleTon) handleReconnect() {
	for {
		c.mu.Lock()
		//case: conn沒有建立 或者已經正常關閉
		if c.conn == nil {
			c.mu.Unlock()
			return
		}

		if c.closeChan == nil {
			// channel 不存在
			c.closeChan = c.conn.NotifyClose(make(chan *amqp.Error))
		}

		c.mu.Unlock()
		// 等待連線關閉事件
		reason, ok := <-c.closeChan
		if !ok {
			// 連線正常關閉
			return
		}

		// case: 收到錯誤訊息，執行重連
		log.Printf("Connection closed: %v", reason)
		for {
			c.mu.Lock()
			conn, err := amqp.Dial(c.url)
			if err != nil {
				c.mu.Unlock()
				log.Printf("Failed to reconnect: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}
			c.closeChan = nil
			c.conn = conn
			c.mu.Unlock()
			log.Println("Reconnected successfully")
			break
		}
	}
}

// Channel 建立新的 channel
func (c *ConnectionSingleTon) Channel() (*amqp.Channel, error) {
	conn, err := c.GetConnection()
	if err != nil {
		return nil, err
	}

	return conn.Channel()
}

// 關閉連線
func (c *ConnectionSingleTon) Close() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %v", err)
	}
	c.conn = nil
	c.closeChan = nil
	return nil
}
