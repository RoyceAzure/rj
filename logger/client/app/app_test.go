package app

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/RoyceAzure/lab/rj_kafka/pkg/admin"
	"github.com/RoyceAzure/lab/rj_kafka/pkg/config"
	kafka_config "github.com/RoyceAzure/lab/rj_kafka/pkg/config"
	"github.com/RoyceAzure/lab/rj_kafka/pkg/model"
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

func TestIntergrationApp(t *testing.T) {
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
			name:                  "app test",
			testMsgs:              100000,
			loggerNum:             3000,
			each_publish_num:      1,
			each_publish_duration: 1 * time.Millisecond,
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
			cfg.Brokers = []string{"localhost:9092", "localhost:9093", "localhost:9094"}
			cfg.Partition = 3
			cfg.ConsumerGroup = LogConsumerGropu
			cfg.Timeout = time.Second
			cfg.BatchSize = 8000
			cfg.CommitInterval = time.Millisecond * 200
			cfg.RequiredAcks = 1
			cfg.Topic = LogTopicName
			cfg.LogLevel = config.DebugLevel

			successMsgsChan := make(chan kafka.Message, testCase.testMsgs)
			failedMsgsChan := make(chan kafka.Message, testCase.testMsgs)

			var successDeposeCount atomic.Uint32
			var failedDeposeCount atomic.Uint32
			successDeposeCount.Store(0)
			failedDeposeCount.Store(0)

			cfg.ConsumerHandlerSuccessfunc = testCase.handlerSuccessfunc(successMsgsChan, &successDeposeCount)
			cfg.ConsumerHandlerErrorfunc = testCase.handlerErrorfunc(failedMsgsChan, &failedDeposeCount)

			app, err := NewKafkaLoggerApp(cfg, "http://localhost:9200", "password", map[string]interface{}{
				"cleanup.policy":      "delete",
				"retention.ms":        "60000", // 1分鐘後自動刪除
				"min.insync.replicas": "1",
				"delete.retention.ms": "1000", // 標記刪除後1秒鐘清理
			})
			require.NoError(t, err)
			if err := app.Start(); err != nil {
				t.Fatalf("Failed to start app: %v", err)
			}

			kafkaLoggerFactory, err := producer.NewKafkaLoggerFactory(&producer.LoggerProducerConfig{
				KafkaConfig: cfg,
				IsOutPutStd: false,
			})
			require.NoError(t, err)

			t.Log("等待前置作業初始化完成...")
			time.Sleep(time.Second * 10) // 等待前置作業初始化完成

			g := new(errgroup.Group)
			MsgsCh := testCase.generateTestMsg(testCase.testMsgs)
			var loggers []*producer.KafkaLoggerAdapter
			for i := 0; i < testCase.loggerNum; i++ {
				logger, err := kafkaLoggerFactory.GetLoggerProcuder()
				require.NoError(t, err)
				loggers = append(loggers, logger)
			}

			t.Log("等待所有logger初始化完成...")
			time.Sleep(time.Second * 10) // 等待所有消費者處理完

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
			t.Log("發送訊息完成...")
			require.NoError(t, err)

			t.Log("等待所有消費者處理完...")
			time.Sleep(time.Second * 10) // 等待所有消費者處理完

			err = app.Stop(context.Background())
			require.NoError(t, err)

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

			shouldConsumerMsgs := len(allSuccessMsgs) + len(allFailedMsgs) + int(failedDeposeCount.Load())
			missingConsumerMsgs := testCase.testMsgs - shouldConsumerMsgs

			zeroFailedCount := 0
			for i := 0; i < testCase.loggerNum; i++ {
				zeroFailedCount += int(loggers[i].GetErrorCount())
			}

			t.Log("consumer總共處理的 msgs數量: ", len(allSuccessMsgs)+len(allFailedMsgs))
			t.Log("失敗發送的 msgs數量: ", failedDeposeCount.Load())
			t.Logf("zero logger失敗發送的 msgs數量: %d", zeroFailedCount)
			t.Log("損失的訊息數量: ", missingConsumerMsgs)
			require.Less(t, missingConsumerMsgs, int(float64(testCase.testMsgs)*0.05), "損失的訊息數量不應該超過5%")
		})
	}
}
