package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testAwsClientConfig() AwsClientConfig {
	return AwsClientConfig{
		Endpoint:  "", // LocalStack 預設端點
		Region:    "",
		AccessKey: "",
		SecretKey: "",
		FilterKey: "routing_key",
	}
}

func TestNewAWSProducer(t *testing.T) {
	cf := testAwsClientConfig()
	name := "test-producer"

	producer, err := NewAWSProducer(name, cf)

	require.NoError(t, err)
	require.NotNil(t, producer)
	require.Contains(t, producer.Name, name) // Name 會被格式化為 "AWS SNS Producer {name}_{uuid}"
	require.NotNil(t, producer.BaseAWSSNSClient)
	require.NotNil(t, producer.client)
	require.Equal(t, cf, producer.CF)
}

func TestAWSSNSProducer_Close(t *testing.T) {
	cf := testAwsClientConfig()
	producer, err := NewAWSProducer("test-producer", cf)
	require.NoError(t, err)

	err = producer.Close()
	require.NoError(t, err)
}

func TestAWSSNSProducer_Publish(t *testing.T) {
	cf := testAwsClientConfig()
	producer, err := NewAWSProducer("test-producer", cf)
	require.NoError(t, err)
	require.NotNil(t, producer)

	// 測試參數
	message := []byte("test message content")

	// 執行 Publish
	err = producer.Publish("data_processer", "back_testing", message)
	require.NoError(t, err)
}
