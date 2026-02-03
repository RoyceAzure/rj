package client

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/RoyceAzure/rj/infra/mq/constant"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
)

// AWS SQS Consumer
type AWSsqsConsumer struct {
	*BaseAWSSQSClient
	Name   string
	status atomic.Int32 // 消費者狀態 只有使用ClientStop and ClientRunning 來控制
}

func NewAWSsqsConsumer(name string, cf AwsClientConfig) (*AWSsqsConsumer, error) {
	client, err := NewBaseAWSSQSClient(cf)
	if err != nil {
		return nil, err
	}
	consumer := &AWSsqsConsumer{
		BaseAWSSQSClient: client,
		Name:             fmt.Sprintf("AWS SQS Consumer %s", name+"_"+uuid.New().String()),
	}
	consumer.status.Store(int32(constant.ClientStop))
	return consumer, nil
}

// 非阻塞消費消息
// params:
//
//		queueName: SQS 佇列名稱
//		handler: 處理消息的函數
//
//	return:
//	 	1. error
func (c *AWSsqsConsumer) Consume(queueName, tag string, handler func([]byte) error) error {
	if c.status.CompareAndSwap(int32(constant.ClientStop), int32(constant.ClientRunning)) {
		go c.consume(queueName, handler)
		return nil
	}
	return nil
}

func (c *AWSsqsConsumer) consume(queueName string, handler func([]byte) error) error {
	for c.status.Load() == int32(constant.ClientRunning) {
		// 不用timeout, 因為aws 有 long polling 機制
		resp, err := c.client.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueName),
			MaxNumberOfMessages: c.CF.MaxNumberOfMessages,
			WaitTimeSeconds:     c.CF.WaitTimeSeconds,
			VisibilityTimeout:   c.CF.VisibilityTimeout,
		})

		if err != nil {
			log.Printf("Consumer %s, 接收失敗: %v", c.Name, err)
			continue
		}

		if len(resp.Messages) == 0 {
			continue
		}

		for _, msg := range resp.Messages {
			log.Printf("Consumer %s, 收到訊息: %s\n", c.Name, *msg.Body)
			err = handler([]byte(*msg.Body))
			if err != nil {
				log.Printf("Consumer %s, 處理失敗: %v", c.Name, err)
				continue
			}

			_, err = c.client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueName),
				ReceiptHandle: msg.ReceiptHandle,
			})

			if err != nil {
				log.Printf("Consumer %s, 刪除失敗: %v", c.Name, err)
			}
		}
	}
	return nil
}

func (c *AWSsqsConsumer) Close() error {
	if c.status.CompareAndSwap(int32(constant.ClientRunning), int32(constant.ClientStop)) {
		return nil
	}
	return nil
}

func (c *AWSsqsConsumer) ReStart(queueName, tag string, handler func([]byte) error) error {
	if c.status.CompareAndSwap(int32(constant.ClientStop), int32(constant.ClientRunning)) {
		go c.consume(queueName, handler)
		return nil
	}
	return nil
}
