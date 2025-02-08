package mq

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type IMQConnManager interface {
	Connect() error
	RegisterChannel() (int, *amqp.Channel, error)
	GetChannel(channelId int) (*amqp.Channel, error)
	ReleaseChannel(channelId int) error
}

type MQConnParams struct {
	MqHost  string
	MqPort  string
	MqUser  string
	MqPas   string
	MqVHost string
}

// 要能自己處理channel  透過connect factory
type MQSelectConnManager struct {
	params       MQConnParams
	conn         *amqp.Connection
	connMu       sync.RWMutex
	channelMu    sync.RWMutex
	channels     map[int]*amqp.Channel
	channelCount int
	closeChan    chan *amqp.Error
	done         atomic.Bool
}

func NewMQSelectConnManager(params MQConnParams) (*MQSelectConnManager, error) {
	manager := MQSelectConnManager{
		params:   params,
		channels: make(map[int]*amqp.Channel, 20),
	}
	err := manager.Connect()

	if err != nil {
		return nil, err
	}

	return &manager, nil
}

func (r *MQSelectConnManager) getUrl() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s%s", r.params.MqUser, r.params.MqPas, r.params.MqHost, r.params.MqPort, r.params.MqVHost)
}

func (r *MQSelectConnManager) Connect() error {
	if r.done.Load() {
		return nil
	}

	config := amqp.Config{
		Dial: amqp.DefaultDial(time.Second * 5),
		Properties: amqp.Table{
			"connection_name": "go_conn_manager",
		},
	}

	// 建立連線 (這是阻塞的，會等到連線建立完成或失敗)
	conn, err := amqp.DialConfig(r.getUrl(), config)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	r.conn = conn
	r.closeChan = r.conn.NotifyClose(make(chan *amqp.Error))

	go r.handleReconnect(time.Second * 5)

	return nil
}

func (r *MQSelectConnManager) createChannel() (*amqp.Channel, error) {
	if r.done.Load() {
		return nil, fmt.Errorf("conn manager is closed")
	}

	if err := r.checkConn(); err != nil {
		return nil, err
	}

	channel, err := r.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("create channel failed")
	}
	// 設定 Channel QoS
	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %v", err)
	}

	return channel, nil
}

func (r *MQSelectConnManager) RegisterChannel() (int, *amqp.Channel, error) {
	channel, err := r.createChannel()
	if err != nil {
		return 0, nil, err
	}

	r.channelMu.Lock()
	defer r.channelMu.Unlock()
	r.channelCount++
	r.channels[r.channelCount] = channel

	return r.channelCount, channel, nil
}

func (r *MQSelectConnManager) GetChannel(channelId int) (*amqp.Channel, error) {
	r.channelMu.RLock()
	ch, ok := r.channels[channelId]
	r.channelMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("channel id not exists, please call RegisterChannel first")
	}

	return ch, nil
}

func (r *MQSelectConnManager) ReleaseChannel(channelId int) error {
	r.channelMu.Lock()
	defer r.channelMu.Unlock()

	ch, ok := r.channels[channelId]
	if !ok {
		return fmt.Errorf("channel id not exists")
	}

	delete(r.channels, channelId)

	return ch.Close()
}

func (r *MQSelectConnManager) checkConn() error {
	r.connMu.RLock()
	defer r.connMu.RUnlock()
	if r.conn != nil && r.conn.IsClosed() {
		return fmt.Errorf("conn is closed")
	}
	return nil
}

func (r *MQSelectConnManager) reCreateChannel() error {
	if r.done.Load() {
		return fmt.Errorf("conn manager is closed")
	}

	if err := r.checkConn(); err != nil {
		return err
	}

	ers := []error{}
	r.channelMu.Lock()
	for k := range r.channels {
		channel, err := r.createChannel()
		if err != nil {
			fmt.Println(err)
			ers = append(ers, err)
		}
		r.channels[k] = channel
	}
	r.channelMu.Unlock()
	if len(ers) > 0 {
		return errors.Join(ers...)
	}

	return nil
}

func (r *MQSelectConnManager) handleReconnect(t time.Duration) {
	for {
		if r.done.Load() {
			return
		}

		if err := r.checkConn(); err != nil {
			return
		}

		// 等待連線關閉事件
		reason, ok := <-r.closeChan
		if !ok {
			// 連線正常關閉
			return
		}

		// case: 收到錯誤訊息，執行重連
		log.Printf("Connection closed, start reconnect: %v", reason)
		for {
			r.connMu.Lock()
			err := r.Connect()
			if err != nil {
				r.connMu.Unlock()
				log.Printf("Failed to reconnect: %v", err)
				time.Sleep(t)
				continue
			}
			r.connMu.Unlock()
			err = r.reCreateChannel()
			if err != nil {
				log.Printf("Failed to recreate channel: %v", err)
				break
			}
			log.Println("Reconnected successfully")
			break
		}
	}
}

func (r *MQSelectConnManager) Close() error {
	if r.done.Load() {
		return nil
	}
	r.done.Store(true)

	r.connMu.Lock()
	if r.conn != nil && !r.conn.IsClosed() {
		err := r.conn.Close()
		if err != nil {
			r.conn = nil
			return fmt.Errorf("failed to close rabbitmq conn, set conn = nil")
		}
	}
	return nil
}
