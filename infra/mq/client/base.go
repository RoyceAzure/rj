package client

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/infra/mq/constant"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// 基礎的消費者與生產者
type BaseClient struct {
	id        string
	name      string
	channel   *amqp.Channel
	channelMu sync.RWMutex
	status    atomic.Int32
	done      chan struct{}
}

func NewBaseClient(name string) *BaseClient {
	return &BaseClient{
		id:   uuid.New().String(),
		name: name,
		done: make(chan struct{}),
	}
}

func (b *BaseClient) setChanFromManger() error {
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

	fmt.Printf("client %s_%s 重置channel", b.name, b.id)
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
		case <-b.done:
			return fmt.Errorf("client 已經關閉，終止重置channel")
		}

	}

	fmt.Printf("conn manager 連線恢復, 重置channel")
	err = b.setChanFromManger()
	if err != nil {
		return err
	}

	return nil
}

// 設置狀態與發送結束訊號
func (b *BaseClient) close() error {
	if b.status.Load() == int32(constant.ClientStop) {
		return nil
	}

	fmt.Printf("開始關閉 client %s_%s", b.name, b.id)
	close(b.done)
	b.status.Store(int32(constant.ClientStop))
	return nil
}

// 只有當consumer處於Stop狀態時 才會執行重啟
func (b *BaseClient) reStart() error {
	fmt.Printf("開始重啟 client %s_%s", b.name, b.id)
	if b.status.Load() != int32(constant.ClientStop) {
		return fmt.Errorf("client is not in stop state")
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
	b.done = make(chan struct{})
	err := b.resetChannel()
	if err != nil {
		return err
	}

	return nil
}

type AwsClientConfig struct {
	Endpoint  string // ARN or URL
	Region    string
	AccessKey string
	SecretKey string
	FilterKey string // SNS topic的filter key
}

// AWS SNS Client
type BaseAWSSNSClient struct {
	CF     AwsClientConfig
	client *sns.Client
}

func NewBaseAWSClient(cf AwsClientConfig) (*BaseAWSSNSClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cf.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cf.AccessKey, cf.SecretKey, "")),
	)
	if err != nil {
		log.Fatalf("無法載入 SDK 設定: %v", err)
	}
	client := sns.NewFromConfig(cfg)
	return &BaseAWSSNSClient{
		CF:     cf,
		client: client,
	}, nil
}
