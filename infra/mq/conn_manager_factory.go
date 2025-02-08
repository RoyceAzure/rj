package mq

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	SelectConnFactory MQConnManagerSingleTonFactory
)

type MQConnManagerFactory interface {
	Init(params MQConnParams) error
	GetManager() (IMQConnManager, error)
}

type MQConnManagerSingleTonFactory struct {
	mu       sync.Mutex
	inited   atomic.Bool
	instance IMQConnManager
}

func (f *MQConnManagerSingleTonFactory) Init(params MQConnParams) error {
	if f.inited.Load() {
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	manager, err := NewMQSelectConnManager(params)
	if err != nil {
		return err
	}

	f.instance = manager
	f.inited.Store(true)

	return nil
}

func (f *MQConnManagerSingleTonFactory) GetManager() (IMQConnManager, error) {
	if !f.inited.Load() {
		return nil, fmt.Errorf("mq conn manager is not init yet, please call init first")
	}

	return f.instance, nil
}
