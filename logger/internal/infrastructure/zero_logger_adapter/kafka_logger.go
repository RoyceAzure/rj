package zero_logger_adapter

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	lab_config "github.com/RoyceAzure/lab/rj_kafka/pkg/config"
	"github.com/RoyceAzure/lab/rj_kafka/pkg/model"
	"github.com/RoyceAzure/lab/rj_kafka/pkg/producer"
	"github.com/segmentio/kafka-go"
)

var (
	instance *kafka.Writer
	mu       sync.RWMutex
)

// for consumer  不需要接收來自zerolog的資訊
type KafkaLoggerWriter struct {
	cfg        *lab_config.Config
	p          producer.Producer
	errorCount atomic.Int64
	logId      atomic.Int64
}

// 用傳入的Config替換單例的kafka writer
func SetKafkaWriterSingleton(cfg *lab_config.Config) {
	mu.Lock()
	defer mu.Unlock()
	instance = createKafkaLoggerWriter(cfg)
}

// 取得單例的kafka writer
// 若沒有則使用傳入的Config建立
func getKafkaWriterSingleton(cfg *lab_config.Config) (*kafka.Writer, error) {
	mu.RLock()
	if instance != nil {
		mu.RUnlock()
		return instance, nil
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()
	if instance == nil {
		instance = createKafkaLoggerWriter(cfg)
	}
	return instance, nil
}

func createKafkaLoggerWriter(cfg *lab_config.Config) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     cfg.GetBalancer(),
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.CommitInterval + time.Millisecond*100,
		RequiredAcks: kafka.RequiredAcks(cfg.RequiredAcks),
		// 設置較短的超時時間以快速發現問題
		WriteTimeout: 5 * time.Second,
		// 設置重試
		MaxAttempts: 3,
		Compression: kafka.Lz4,
	}
}

// 建立kafka logger writer
// 注意KafkaLoggerWriter 內部使用的kafka writer是單例的
func NewKafkaLoggerWriter(cfg *lab_config.Config) (*KafkaLoggerWriter, error) {
	instance, err := getKafkaWriterSingleton(cfg)
	if err != nil {
		return nil, err
	}

	p, err := producer.NewConcurrencekafkaProducer(instance, *cfg)
	if err != nil {
		return nil, err
	}

	w := &KafkaLoggerWriter{
		cfg: cfg,
		p:   p,
	}
	w.logId.Store(0)
	w.errorCount.Store(0)

	p.Start()

	return w, nil
}

func (kw *KafkaLoggerWriter) GetErrorCount() int64 {
	return kw.errorCount.Load()
}

func (kw *KafkaLoggerWriter) Write(p []byte) (n int, err error) {
	if kw == nil {
		return 0, fmt.Errorf("kafka logger is not init")
	}

	kw.logId.Add(int64(1))
	kbuf := make([]byte, 8)
	binary.BigEndian.PutUint64(kbuf, uint64(kw.logId.Load()))

	copyP := make([]byte, len(p))
	copy(copyP, p) //避免zero logger在寫入時的競爭問題

	_, err = kw.p.Produce(context.Background(), []model.Message{
		{
			Key:   kbuf,
			Value: copyP,
		},
	})

	if err != nil {
		kw.errorCount.Add(1)
		return 0, err
	}

	return len(p), nil
}

func (fw *KafkaLoggerWriter) Close() error {
	return fw.p.Close(time.Second * 15)
}
