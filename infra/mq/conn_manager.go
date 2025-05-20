package mq

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/RoyceAzure/rj/infra/mq/constant"
	amqp "github.com/rabbitmq/amqp091-go"
)

type IMQConnManager interface {
	Connect() error
	GetChannel() (*amqp.Channel, error)
	Close() error
	Status() int32
	// 註冊監聽Manager重連成功,有避免重複註冊功能
	Register(id string, ch chan struct{})
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
	params        MQConnParams
	conn          *amqp.Connection
	connMu        sync.RWMutex
	closeChan     chan *amqp.Error //監聽rabbitmq server正常關閉與否
	subscribers   sync.Map
	reConnRunning atomic.Bool
	// subscribers   []chan struct{}
	// subscribersMu sync.RWMutex
	status atomic.Int32
	done   atomic.Bool
}

func NewMQSelectConnManager(params MQConnParams) (*MQSelectConnManager, error) {
	manager := MQSelectConnManager{
		params: params,
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

// return :
//
//	StatusConnected 1
//	StatusDisconnected 2
//	StatusReconnecting 3
//	StatusClosed 4
func (r *MQSelectConnManager) Status() int32 {
	return r.status.Load()
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

	r.connMu.Lock()
	r.conn = conn
	r.closeChan = r.conn.NotifyClose(make(chan *amqp.Error))
	r.connMu.Unlock()

	r.status.Store(int32(constant.ManagerStatusConnected))
	if r.reConnRunning.CompareAndSwap(false, true) {
		go r.handleReconnect(time.Second * 5)
	}

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

// 建立並返回新的channel
//
//		error:
//	 	1. conn is closed
//	 	2. create channel failed
func (r *MQSelectConnManager) GetChannel() (*amqp.Channel, error) {
	return r.createChannel()
}

// 檢查連線是否正常
//
//		error:
//	 	1. conn is closed
//	 	2. conn is not init
func (r *MQSelectConnManager) checkConn() error {
	r.connMu.RLock()
	defer r.connMu.RUnlock()

	if r.conn == nil || r.conn.IsClosed() {
		r.status.Store(int32(constant.ManagerStatusDisconnected))
		return fmt.Errorf("conn is closed")
	}

	r.status.Store(int32(constant.ManagerStatusConnected))
	return nil
}

// ConnectManager在沒有收到Close()的情況下  任何形式的與rabbitmq server的連線斷開都會觸發重連
func (r *MQSelectConnManager) handleReconnect(t time.Duration) {
	for {
		if r.done.Load() {
			return
		}

		// 等待連線關閉事件
		reason := <-r.closeChan
		r.status.Store(int32(constant.ManagerStatusReconnecting))

		// case: 收到錯誤訊息，執行重連
		log.Printf("Connection closed, start reconnect: %v", reason)
		for {
			if r.done.Load() {
				return
			}
			//開始重連
			err := r.Connect()
			if err != nil {
				log.Printf("Failed to reconnect: %v", err)
				time.Sleep(t)
				continue
			}
			log.Printf("reconnect success")
			break
		}
		r.broadcast()
	}
}

// 註冊監聽Manager重連成功
func (r *MQSelectConnManager) Register(id string, ch chan struct{}) {
	if r.status.Load() == int32(constant.ManagerStatusConnected) {
		ch <- struct{}{}
	}
	r.subscribers.LoadOrStore(id, ch)
}

// 重連成功後 廣播訊號，通知所有註冊的消費者
func (r *MQSelectConnManager) broadcast() {
	r.subscribers.Range(func(key, value any) bool {
		select {
		case value.(chan struct{}) <- struct{}{}:
		default:
			// 如果channel已經滿了，則不會再發送
		}
		return true
	})
	r.subscribers.Clear()
}

func (r *MQSelectConnManager) Close() error {
	if r.done.Load() {
		return nil
	}
	r.done.Store(true)
	r.status.Store(int32(constant.ManagerStatusClosed))
	r.reConnRunning.Store(false)
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
