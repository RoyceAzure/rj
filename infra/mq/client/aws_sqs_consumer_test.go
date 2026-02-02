package client

import (
	"log"
	"testing"
	"time"

	"github.com/RoyceAzure/rj/infra/mq/constant"
	"github.com/stretchr/testify/require"
)

func testAwsSQSClientConfig() AwsClientConfig {
	return AwsClientConfig{
		Endpoint:            "",
		Region:              "ap-northeast-1",
		AccessKey:           "",
		SecretKey:           "",
		FilterKey:           "routing_key",
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
		VisibilityTimeout:   30,
	}
}

func TestNewAWSsqsConsumer(t *testing.T) {
	cf := testAwsSQSClientConfig()
	name := "test-consumer"

	consumer, err := NewAWSsqsConsumer(name, cf)

	require.NoError(t, err)
	require.NotNil(t, consumer)
	require.Contains(t, consumer.Name, name) // Name 會被格式化為 "AWS SQS Consumer {name}_{uuid}"
	require.NotNil(t, consumer.BaseAWSSQSClient)
	require.NotNil(t, consumer.client)
	require.Equal(t, cf, consumer.CF)
	require.Equal(t, int32(constant.ClientStop), consumer.status.Load())
}

func TestAWSsqsConsumer_Close(t *testing.T) {
	cf := testAwsSQSClientConfig()
	consumer, err := NewAWSsqsConsumer("test-consumer", cf)
	require.NoError(t, err)

	// 初始狀態應該是 ClientStop
	require.Equal(t, int32(constant.ClientStop), consumer.status.Load())

	// 關閉應該成功
	err = consumer.Close()
	require.NoError(t, err)
	require.Equal(t, int32(constant.ClientStop), consumer.status.Load())
}

func TestAWSsqsConsumer_Consume(t *testing.T) {
	cf := testAwsSQSClientConfig()
	consumer, err := NewAWSsqsConsumer("test-consumer", cf)
	require.NoError(t, err)

	// 初始狀態應該是 ClientStop
	require.Equal(t, int32(constant.ClientStop), consumer.status.Load())

	// 測試 handler
	handler := func(message []byte) error {
		log.Printf("Consumer %s, 收到訊息: %s", consumer.Name, string(message))
		return nil
	}

	// 開始消費（會啟動 goroutine）
	err = consumer.Consume("", handler)
	require.NoError(t, err)

	// 狀態應該變為 ClientRunning
	require.Equal(t, int32(constant.ClientRunning), consumer.status.Load())

	// 等待一小段時間讓 goroutine 啟動
	time.Sleep(100 * time.Millisecond)

	// 關閉 consumer
	err = consumer.Close()
	require.NoError(t, err)
	require.Equal(t, int32(constant.ClientStop), consumer.status.Load())
}

func TestAWSsqsConsumer_ReStart(t *testing.T) {
	cf := testAwsSQSClientConfig()
	consumer, err := NewAWSsqsConsumer("test-consumer", cf)
	require.NoError(t, err)

	// 初始狀態應該是 ClientStop
	require.Equal(t, int32(constant.ClientStop), consumer.status.Load())

	// 測試 handler
	handler := func([]byte) error {
		return nil
	}

	// 重啟 consumer（會啟動 goroutine）
	err = consumer.ReStart("", handler)
	require.NoError(t, err)

	// 狀態應該變為 ClientRunning
	require.Equal(t, int32(constant.ClientRunning), consumer.status.Load())

	// 等待一小段時間讓 goroutine 啟動
	time.Sleep(100 * time.Millisecond)

	// 關閉 consumer
	err = consumer.Close()
	require.NoError(t, err)
	require.Equal(t, int32(constant.ClientStop), consumer.status.Load())
}

func TestAWSsqsConsumer_Consume_AlreadyRunning(t *testing.T) {
	cf := testAwsSQSClientConfig()
	consumer, err := NewAWSsqsConsumer("test-consumer", cf)
	require.NoError(t, err)

	handler := func([]byte) error {
		return nil
	}

	// 第一次 Consume，應該成功
	err = consumer.Consume("", handler)
	require.NoError(t, err)
	require.Equal(t, int32(constant.ClientRunning), consumer.status.Load())

	// 第二次 Consume，應該不會改變狀態（因為已經在運行）
	err = consumer.Consume("", handler)
	require.NoError(t, err)
	require.Equal(t, int32(constant.ClientRunning), consumer.status.Load())

	// 清理
	consumer.Close()
}
