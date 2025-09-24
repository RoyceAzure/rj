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

// WorkerPool 管理 Worker
type WorkerPool struct {
	pool *sync.Pool
}

// NewWorkerPool 創建 WorkerPool
func NewWorkerPool() *WorkerPool {
	return &WorkerPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return NewWorker(WorkerOption{})
			},
		},
	}
}

// Get 獲取 Worker
func (wp *WorkerPool) Get() *Worker {
	return wp.pool.Get().(*Worker)
}

// Put 回收 Worker 並清理狀態
func (wp *WorkerPool) Put(w *Worker) {
	w.Reset() // 在放回 Pool 前重置狀態
	wp.pool.Put(w)
}

// 分配策略要能動態調整
// worker 要能動態調整
// 活耀worker 與 worker pool 的互動關係?
// 什時需要從pool中取出  什時要回收
// 由誰來做這件事情

//事件分配應該只分配到活耀worker， 而活耀worker數量由一個模組來管理  控制活耀worker 與 worker pool 的互動關係
// 要有一個統一的SyncTaskSystem 指標  紀錄TaskSystem狀態  用來控制子模組動態協調

/*
 */
type SyncTaskSystem struct {
	ctx         context.Context
	cancleFunc  context.CancelFunc
	taskQueue   chan WorkerTask
	resultQueue chan *TaskResult
	workers     []IWoker //活耀worker
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
