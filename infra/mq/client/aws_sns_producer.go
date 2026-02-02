package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/google/uuid"
)

// AWS SNS Producer
type AWSSNSProducer struct {
	*BaseAWSSNSClient
	Name string
}

func NewAWSProducer(name string, cf AwsClientConfig) (*AWSSNSProducer, error) {
	client, err := NewBaseAWSClient(cf)
	if err != nil {
		return nil, err
	}
	return &AWSSNSProducer{
		Name:             fmt.Sprintf("AWS SNS Producer %s", name+"_"+uuid.New().String()),
		BaseAWSSNSClient: client,
	}, nil
}

//	發送消息到AWS SNS
//
// params:
//
//	exchange: 交換機 等同於AWS SNS topic, 這裡會使用CF.Endpoint
//	routingKey: 路由鍵 等同於AWS SNS topic的routing key, 使用MessageAttributes
//	message: 消息
//
// returns:
//
//	error: 錯誤
func (p *AWSSNSProducer) Publish(exchange, routingKey string, message []byte) error {
	input := &sns.PublishInput{
		Message:  aws.String(string(message)),
		TopicArn: aws.String(p.CF.Endpoint),
		MessageAttributes: map[string]types.MessageAttributeValue{
			p.CF.FilterKey: {
				DataType:    aws.String("String"),
				StringValue: aws.String(routingKey),
			},
		},
	}

	// 5. 發送
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	result, err := p.client.Publish(ctx, input)
	if err != nil {
		log.Fatalf("發送失敗: %v", err)
		return err
	}
	log.Printf("[Success] send message to AWS SNS %s, Message ID: %s, routingKey: %s, routingValue: %s", p.Name, *result.MessageId, p.CF.FilterKey, routingKey)
	return nil
}

func (p *AWSSNSProducer) Close() error {
	return nil
}
