package client

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/RoyceAzure/lab/rj_kafka/pkg/admin"
	kafka_config "github.com/RoyceAzure/lab/rj_kafka/pkg/config"
	kafka_consumer "github.com/RoyceAzure/lab/rj_kafka/pkg/consumer"
	"github.com/RoyceAzure/lab/rj_kafka/pkg/model"
	"github.com/RoyceAzure/rj/logger/client/consumer"
	"github.com/RoyceAzure/rj/logger/client/producer"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

var (
	testClusterConfig *admin.ClusterConfig
	cfg               *kafka_config.Config
	ConsuemErrorGroup = "consumer-group"
	LogConsumerGropu  = fmt.Sprintf("%s-log-%d", ConsuemErrorGroup, time.Now().UnixNano())

	TopicPrefix  = "test-topic"
	LogTopicName = fmt.Sprintf("%s-log-%d", TopicPrefix, time.Now().UnixNano())
)

type TestMsg struct {
	Id      int    `json:"id"`
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

func generateTestMessage(n int) chan string {
	t := make(chan string, n)
	go func() {
		for i := 0; i < n; i++ {
			t <- fmt.Sprintf("this is test log %d", i)
		}
		close(t)
	}()
	return t
}

func setupTestEnvironment(t *testing.T, cfg *kafka_config.Config) func() {
	testClusterConfig = &admin.ClusterConfig{
		Cluster: struct {
			Name    string   `yaml:"name"`
			Brokers []string `yaml:"brokers"`
		}{
			Name:    "test-cluster",
			Brokers: []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		},
	}
	cfg.Brokers = testClusterConfig.Cluster.Brokers

	// 創建 admin client
	adminClient, err := admin.NewAdmin(testClusterConfig.Cluster.Brokers)
	require.NoError(t, err)

	// 創建 cart command topic，設定較短的retention時間以便自動清理
	err = adminClient.CreateTopic(context.Background(), admin.TopicConfig{
		Name:              LogTopicName,
		Partitions:        cfg.Partition,
		ReplicationFactor: 3,
		Configs: map[string]interface{}{
			"cleanup.policy":      "delete",
			"retention.ms":        "60000", // 1分鐘後自動刪除
			"min.insync.replicas": "2",
			"delete.retention.ms": "1000", // 標記刪除後1秒鐘清理
		},
	})
	require.NoError(t, err)

	// 等待 topic 創建完成
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = adminClient.WaitForTopics(ctx, []string{LogTopicName}, 30*time.Second)
	require.NoError(t, err)

	// 更新全局配置
	testClusterConfig.Topics = []admin.TopicConfig{
		{
			Name:              LogTopicName,
			Partitions:        cfg.Partition,
			ReplicationFactor: 3,
		},
	}

	// 返回清理函數，只需關閉adminClient
	return func() {
		adminClient.Close()
	}
}

// 似乎有訊息在zero logger階段中消失，並沒有進到kafka producer buffer裡面
// 測試前請先設定kafka topics，或者啟動server端，目前是用來測試server端的
func TestLoggerProduce(t *testing.T) {
	testCases := []struct {
		name                  string
		each_publish_num      int
		each_publish_duration time.Duration
		testMsgs              int
		loggerNum             int
		generateTestMsg       func(int) chan string
		earilyStop            time.Duration
	}{
		{
			name:                  "10 logger, 20000 messages, ",
			testMsgs:              100000,
			loggerNum:             3000,
			each_publish_num:      2,
			each_publish_duration: 500 * time.Millisecond,
			generateTestMsg:       generateTestMessage,
			earilyStop:            time.Second * 20,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cfg = kafka_config.DefaultConfig()
			cfg.Brokers = []string{"localhost:9092", "localhost:9093", "localhost:9094"}
			cfg.Partition = 9
			cfg.ConsumerGroup = "log_consumer_group"
			cfg.Timeout = time.Second
			cfg.BatchSize = 8000
			cfg.CommitInterval = time.Millisecond * 200
			cfg.RequiredAcks = 1

			cfg.Topic = "log"

			kafkaLoggerFactory, err := producer.NewKafkaLoggerFactory(&producer.LoggerProducerConfig{
				KafkaConfig: cfg,
				IsOutPutStd: false,
			})
			require.NoError(t, err)

			var loggers []*producer.KafkaLoggerAdapter
			for i := 0; i < testCase.loggerNum; i++ {
				logger, err := kafkaLoggerFactory.GetLoggerProcuder()
				require.NoError(t, err)
				loggers = append(loggers, logger)
			}

			t.Log("等待前置作業初始化完成...")
			time.Sleep(time.Second * 10) // 等待前置作業初始化完成

			g := new(errgroup.Group)
			MsgsCh := testCase.generateTestMsg(testCase.testMsgs)
			t.Log("開始發送訊息...")
			for i := 0; i < testCase.loggerNum; i++ {
				g.Go(func() error {
					l := loggers[i]
					for msg := range MsgsCh {
						time.Sleep(testCase.each_publish_duration)
						l.Info().Msg(msg)
					}
					return nil
				})
			}

			t.Log("等待發送訊息完成...")
			err = g.Wait()
			time.Sleep(time.Second * 20) // 等待producer處理完
			t.Log("發送訊息完成...")
			t.Log("error count: ", loggers[0].GetErrorCount())
			require.NoError(t, err)
			t.Log("發送訊息完成...")
		})
	}
}

func TestIntergrationTest(t *testing.T) {
	testCases := []struct {
		name                  string
		each_publish_num      int
		each_publish_duration time.Duration
		testMsgs              int
		loggerNum             int
		generateTestMsg       func(int) chan string
		earilyStop            time.Duration
		handlerSuccessfunc    func(successMsgsChan chan kafka.Message, successDeposeCount *atomic.Uint32) func(kafka.Message)
		handlerErrorfunc      func(failedMsgsChan chan kafka.Message, failedDeposeCount *atomic.Uint32) func(model.ConsuemError)
	}{
		{
			name:                  "1 logger,  default consumer, 10000 messages",
			testMsgs:              100000,
			loggerNum:             3000,
			each_publish_num:      2,
			each_publish_duration: 500 * time.Millisecond,
			generateTestMsg:       generateTestMessage,
			earilyStop:            time.Second * 20,
			handlerSuccessfunc: func(successMsgsChan chan kafka.Message, successDeposeCount *atomic.Uint32) func(kafka.Message) {
				return func(msg kafka.Message) {
					select {
					case successMsgsChan <- msg:
					default:
						successDeposeCount.Add(1)
					}
				}
			},
			handlerErrorfunc: func(failedMsgsChan chan kafka.Message, failedDeposeCount *atomic.Uint32) func(model.ConsuemError) {
				return func(err model.ConsuemError) {
					select {
					case failedMsgsChan <- err.Message:
					default:
						failedDeposeCount.Add(1)
					}
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cfg = kafka_config.DefaultConfig()
			cfg.Partition = 6
			cfg.ConsumerGroup = LogConsumerGropu
			cfg.Timeout = time.Second
			cfg.BatchSize = 8000
			cfg.CommitInterval = time.Millisecond * 200
			cfg.RequiredAcks = 1
			cfg.Topic = LogTopicName
			closeEnv := setupTestEnvironment(t, cfg)
			defer closeEnv()

			successMsgsChan := make(chan kafka.Message, testCase.testMsgs)
			failedMsgsChan := make(chan kafka.Message, testCase.testMsgs)

			var successDeposeCount atomic.Uint32
			var failedDeposeCount atomic.Uint32
			successDeposeCount.Store(0)
			failedDeposeCount.Store(0)

			cfg.ConsumerHandlerSuccessfunc = testCase.handlerSuccessfunc(successMsgsChan, &successDeposeCount)
			cfg.ConsumerHandlerErrorfunc = testCase.handlerErrorfunc(failedMsgsChan, &failedDeposeCount)

			kafkaConsumerFactory, err := consumer.NewKafkaConsumerFactory(&consumer.LoggerConsumerConfig{
				ElUrl:       "http://localhost:9200",
				ElPas:       "password",
				KafkaConfig: cfg,
			})
			require.NoError(t, err)

			var consumers []*kafka_consumer.Consumer
			for i := 0; i < cfg.Partition; i++ {
				consumer, err := kafkaConsumerFactory.GetLoggerConsumer()
				require.NoError(t, err)
				consumer.Start()
				consumers = append(consumers, consumer)
			}

			kafkaLoggerFactory, err := producer.NewKafkaLoggerFactory(&producer.LoggerProducerConfig{
				KafkaConfig: cfg,
			})
			require.NoError(t, err)

			t.Log("等待前置作業初始化完成...")
			time.Sleep(time.Second * 10) // 等待前置作業初始化完成

			g := new(errgroup.Group)
			MsgsCh := testCase.generateTestMsg(testCase.testMsgs)
			t.Log("開始發送訊息...")
			var loggers []*producer.KafkaLoggerAdapter
			for i := 0; i < testCase.loggerNum; i++ {
				logger, err := kafkaLoggerFactory.GetLoggerProcuder()
				require.NoError(t, err)
				loggers = append(loggers, logger)
			}

			for i := 0; i < testCase.loggerNum; i++ {
				g.Go(func() error {
					l := loggers[i]
					for msg := range MsgsCh {
						time.Sleep(testCase.each_publish_duration)
						l.Info().Msg(msg)
					}
					return nil
				})
			}

			t.Log("等待發送訊息完成...")
			err = g.Wait()
			t.Log("發送訊息完成...")
			require.NoError(t, err)

			t.Log("等待所有消費者處理完...")
			time.Sleep(time.Second * 20) // 等待所有消費者處理完
			for _, c := range consumers {
				go c.Close(time.Second * 10)
			}

			for _, c := range consumers {
				<-c.C()
			}

			t.Log("所有消費者處理完且已經關閉...")

			//處理所有消費者處理完的結果
			t.Log("開始處理消費者結果...")
			close(successMsgsChan)
			close(failedMsgsChan)
			resultGroup := new(errgroup.Group)
			var allSuccessMsgs, allFailedMsgs []kafka.Message
			resultGroup.Go(func() error {
				for msg := range successMsgsChan {
					allSuccessMsgs = append(allSuccessMsgs, msg)
				}

				return nil
			})
			resultGroup.Go(func() error {
				for msg := range failedMsgsChan {
					allFailedMsgs = append(allFailedMsgs, msg)
				}
				return nil
			})
			t.Log("等待處理消費者結果完成...")
			resultGroup.Wait()
			t.Log("處理消費者結果完成...")

			t.Log("failed count: ", failedDeposeCount.Load())
			require.Equal(t, testCase.testMsgs, len(allSuccessMsgs)+len(allFailedMsgs))
		})
	}
}
