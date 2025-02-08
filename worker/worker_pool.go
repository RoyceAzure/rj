package worker

import (
	"context"
	"fmt"
	"sync"
)

const (
	//一個Worker分配到一組工作數量有多少個
	WorkerProcessTaskNum = 100
)

var iterTaskPool = sync.Pool{
	New: func() any {
		return make([]WorkerTask, 0, 100)
	},
}

/*
worker啟動後  應該是一職存在  不斷從taskQueue取得task並執行
負責協調工作要給哪個worker???
*/
type SyncTaskSystem struct {
	ctx         context.Context
	cancleFunc  context.CancelFunc
	taskQueue   chan WorkerTask
	resultQueue chan *TaskResult
	workers     []IWoker
	mu          sync.RWMutex
}

/*
todo : woker應該要用註冊的方式

@parm

	qsize:taskQueue 大小
*/
func NewSyncTaskSystem(qsize int) (*SyncTaskSystem, context.CancelFunc) {
	ctx, cancle := context.WithCancel(context.Background())
	return &SyncTaskSystem{
		ctx:         ctx,
		cancleFunc:  cancle,
		taskQueue:   make(chan WorkerTask, qsize),
		resultQueue: make(chan *TaskResult, qsize),
	}, cancle
}

func (s *SyncTaskSystem) AddWorker(workers ...IWoker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workers = append(s.workers, workers...)
}

/*
used go routine inside

todo 分配工作部分可以優化
*/
func (s *SyncTaskSystem) Start() {
	for _, w := range s.workers {
		w.Run(s.ctx)
	}
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				fmt.Println("Stop SyncTaskSystem")
				return
			default:
				if len(s.taskQueue) != 0 {
					s.mu.RLock()
					for _, w := range s.workers {
						taskToFeed := s.taskDistributor(WorkerProcessTaskNum)
						w.FeedTask(taskToFeed)
					}
					s.mu.RUnlock()
				}
			}
		}
	}()
}

func (s *SyncTaskSystem) Stop() {
	s.cancleFunc()
}

/*
parm

	size: 每組任務數量
*/
func (s *SyncTaskSystem) taskDistributor(size int) []WorkerTask {
	bufferTaskPool := iterTaskPool.Get().([]WorkerTask)[:0]
	for i := 0; i < size; i++ {
		select {
		case task, ok := <-s.taskQueue:
			if !ok {
				break
			}
			bufferTaskPool = append(bufferTaskPool, task)
		default:
			return bufferTaskPool
		}
	}
	return bufferTaskPool
}

func (s *SyncTaskSystem) TaskInQueue(tasks ...WorkerTask) {
	go func() {
		for _, task := range tasks {
			s.taskQueue <- task
		}
	}()
}
