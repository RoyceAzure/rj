package jkafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaWriter interface {
	WriteMessages(context.Context, []kafka.Message) error
}

type JKafkaWriter struct {
	writer *kafka.Writer
}

func NewJKafkaWriter(addr ...string) KafkaWriter {
	return &JKafkaWriter{
		writer: newKafkaWriter(addr...),
	}
}

func newKafkaWriter(addr ...string) *kafka.Writer {
	return &kafka.Writer{
		Addr:                   kafka.TCP(addr...),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
}

func (jw *JKafkaWriter) WriteMessages(ctx context.Context, msgs []kafka.Message) error {
	var err error
	for i := 0; i < WRITE_RETRIES; i++ {
		err = jw.writer.WriteMessages(ctx, msgs...)
		if errors.Is(err, kafka.LeaderNotAvailable) || errors.Is(err, context.DeadlineExceeded) {
			time.Sleep(time.Millisecond * 250)
			continue
		}
		if err != nil {
			err = fmt.Errorf("kafka write messages get err : %w", err)
		}
		break
	}
	return err
}
