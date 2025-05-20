package client

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/infra/mq/constant"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// 基礎的消費者與生產者
type BaseClient struct {
	id         string
	name       string
	channel    *amqp.Channel
	channelMu  sync.RWMutex
	status     atomic.Int32
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewBaseClient(name string) *BaseClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseClient{
		id:         uuid.New().String(),
		name:       name,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

func (b *BaseClient) getChanFromManger() error {
	b.status.Store(int32(constant.ClientInit))
	ma, err := mq.SelectConnFactory.GetManager()
	if err != nil {
		return err
	}

	channel, err := ma.GetChannel()
	if err != nil {
		return err
	}

	if channel.IsClosed() {
		return fmt.Errorf("channel is closed")
	}

	if err := channel.Confirm(false); err != nil {
		channel.Close()
		return fmt.Errorf("failed to set confirm mode: %v", err)
	}

	b.channelMu.Lock()
	b.channel = channel
	b.channelMu.Unlock()
	return nil
}

// 要等待Manager重連成功後 才會重置channel
// 以下情況發生時 將不再嘗試重連:
//  1. 收到Close指令
//  2. Manager本身有錯誤或者狀態為Closed
//  3. 從Manager取得Channel失敗
func (b *BaseClient) resetChannel() error {
	if b.status.Load() == int32(constant.ClientStop) {
		return nil
	}

	fmt.Print("consumer 重置channel")
	b.status.Store(int32(constant.ClientReset))
	ma, err := mq.SelectConnFactory.GetManager()
	if err != nil {
		return err
	}

	if ma.Status() == int32(constant.ManagerStatusClosed) {
		return fmt.Errorf("manager is closed")
	}
	//等待Manger重連成功
	maStatus := ma.Status()
	if maStatus == int32(constant.ManagerStatusDisconnected) || maStatus == int32(constant.ManagerStatusReconnecting) {
		//等待Manger重連成功
		ch := make(chan struct{})
		ma.Register(b.id, ch)
		select {
		case <-ch:
			break
		case <-b.ctx.Done():
			return fmt.Errorf("consumer 已經關閉，終止重置channel")
		}

	}

	fmt.Printf("conn manager 連線恢復, 重置channel")
	err = b.getChanFromManger()
	if err != nil {
		return err
	}

	return nil
}

// Close 關閉消費者
func (b *BaseClient) Close() error {
	return b.close()
}

// 設定狀態為Stop
// 關閉done channel
func (b *BaseClient) close() error {
	fmt.Print("開始關閉consumer")
	b.status.Store(int32(constant.ClientStop))
	b.cancelFunc()
	return nil
}

// 只有當consumer處於Stop狀態時 才會執行重啟
func (b *BaseClient) reStart() error {
	fmt.Print("開始重啟consumer")
	if b.status.Load() != int32(constant.ClientStop) {
		return fmt.Errorf("consumer is not in stop state")
	}

	err := b.reSet()
	if err != nil {
		return err
	}

	return nil
}

// BaseClient 狀態重置
func (b *BaseClient) reSet() error {
	b.status.Store(int32(constant.ClientInit))
	ctx, cancel := context.WithCancel(context.Background())
	b.ctx = ctx
	b.cancelFunc = cancel
	err := b.resetChannel()
	if err != nil {
		return err
	}

	return nil
}
