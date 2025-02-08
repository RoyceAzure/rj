package jkafka

import (
	"context"
	"errors"
	"sync"

	"github.com/segmentio/kafka-go"
)

type KafkaReader interface {
	ReadMessageAsync(ctx context.Context, ch chan<- kafka.Message, errch chan<- error, wg *sync.WaitGroup)
}

type JKafkaReader struct {
	reader *kafka.Reader
}

/*
獨立消費者
*/
func NewJKafkaReader(config *kafka.ReaderConfig) KafkaReader {
	return &JKafkaReader{
		reader: newKafkaReader(config),
	}
}

func newKafkaReader(config *kafka.ReaderConfig) *kafka.Reader {
	return kafka.NewReader(*config)
}

func (jr *JKafkaReader) ReadMessageAsync(ctx context.Context, ch chan<- kafka.Message, errch chan<- error, wg *sync.WaitGroup) {
	defer close(ch)
	defer close(errch)
	defer wg.Done()
	for {
		msg, err := jr.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			errch <- err
			continue
		}
		ch <- msg
		// err = jr.reader.CommitMessages(ctx, msg)
		// if err != nil {
		// 	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		// 		return
		// 	}
		// 	errch <- err
		// }
	}
}
