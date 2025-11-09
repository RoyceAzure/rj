package pool

import "sync"

type IPool[T any] interface {
	Get() *[]T
	Put(buffer *[]T)
	Stop()
}

type BasicPool[T any] struct {
	mainBufferPool *sync.Pool
}

func NewBasicPool[T any](bufferSize int) *BasicPool[T] {
	return &BasicPool[T]{
		mainBufferPool: &sync.Pool{
			New: func() any {
				s := make([]T, 0, bufferSize)
				return &s
			},
		},
	}
}

func (p *BasicPool[T]) Get() *[]T {
	return p.mainBufferPool.Get().(*[]T)
}

func (p *BasicPool[T]) Put(buffer *[]T) {
	p.mainBufferPool.Put(buffer)
}

func (p *BasicPool[T]) Stop() {
	p.mainBufferPool = nil
}
