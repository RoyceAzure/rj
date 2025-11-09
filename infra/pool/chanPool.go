package pool

import (
	"context"
	"sync"
)

// 使用chan 當作保底池，保底數量poolSize，避免每次都重新分配記憶體
type ChannelPool[T any] struct {
	backupBufferPool sync.Pool
	mainBufferPool   chan *[]T
	poolSize         int
	bufferSize       int
	ctx              context.Context
	cancel           context.CancelFunc
}

// param poolSize: 池子大小，設定預期峰值QPS得1.5~2倍
// param bufferSize: 每個池子的大小
func NewChannelPool[T any](poolSize, bufferSize int) *ChannelPool[T] {
	ctx, cancel := context.WithCancel(context.Background())

	p := ChannelPool[T]{
		poolSize:       poolSize,
		bufferSize:     bufferSize,
		mainBufferPool: make(chan *[]T, poolSize),
		backupBufferPool: sync.Pool{
			New: func() any {
				s := make([]T, 0, bufferSize)
				return &s
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}

	go p.background()

	return &p
}

func (p *ChannelPool[T]) Get() *[]T {
	select {
	case buffer := <-p.mainBufferPool:
		return buffer
	default:
		return p.backupBufferPool.Get().(*[]T)
	}
}

func (p *ChannelPool[T]) Put(buffer *[]T) {
	select {
	case p.mainBufferPool <- buffer:
	default:
		p.backupBufferPool.Put(buffer)
	}
}

func (p *ChannelPool[T]) background() {
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			newBuf := make([]T, 0, p.bufferSize)
			p.mainBufferPool <- &newBuf
		}
	}
}

func (p *ChannelPool[T]) Stop() {
	p.cancel()
}
