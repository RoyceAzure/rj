package pool

import (
	"context"
	"sync/atomic"
)

type ShardChannelPool[T any] struct {
	mainBufferPool []*ChannelPool[T]
	poolSize       int
	bufferSize     int
	shardSize      int
	ctx            context.Context
	cancel         context.CancelFunc
	counter        uint64
}

func NewShardChannelPool[T any](shardSize, poolSize, bufferSize int) *ShardChannelPool[T] {
	ctx, cancel := context.WithCancel(context.Background())
	p := &ShardChannelPool[T]{
		mainBufferPool: make([]*ChannelPool[T], shardSize),
		poolSize:       poolSize,
		bufferSize:     bufferSize,
		shardSize:      shardSize,
		ctx:            ctx,
		cancel:         cancel,
	}
	for i := 0; i < shardSize; i++ {
		p.mainBufferPool[i] = NewChannelPool[T](poolSize, bufferSize)
	}
	return p
}

func (p *ShardChannelPool[T]) Get() *[]T {
	return p.mainBufferPool[atomic.AddUint64(&p.counter, 1)%uint64(p.shardSize)].Get()
}

func (p *ShardChannelPool[T]) Put(buffer *[]T) {
	p.mainBufferPool[atomic.AddUint64(&p.counter, 1)%uint64(p.shardSize)].Put(buffer)
}

func (p *ShardChannelPool[T]) Stop() {
	p.cancel()
	for _, pool := range p.mainBufferPool {
		pool.Stop()
	}
}
