package jkafka

import (
	"github.com/segmentio/kafka-go"
)

const (
	WRITE_RETRIES = 3
)

type JKafka struct {
	conn *kafka.Conn
}

func NewJKafka(conn *kafka.Conn) *JKafka {
	return &JKafka{
		conn: conn,
	}
}
