package scheduler

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
)

/*
cron func 簽名不回傳error  看來error要自己處理
*/
type ISchedulerTask interface {
	RunSchedulerTask()
	GetStatus() *SchedulerTaskStatus
	SetStatus(*SchedulerTaskStatus)
}

type BaseSchedulerTask struct {
	status SchedulerTaskStatus
}

func (b *BaseSchedulerTask) GetStatus() *SchedulerTaskStatus {
	return &b.status
}

func (b *BaseSchedulerTask) SetStatus(s *SchedulerTaskStatus) {
	b.status = *s
}

/*
可能有多個程序會來查詢Task狀態  使用atomic
todo memory只儲存最新一次  後續可能要寫入永久儲存
*/
type SchedulerTaskStatus struct {
	LastRun   atomic.Value // stores time.Time
	NextRun   atomic.Value // stores time.Time
	RunCount  atomic.Int64
	LastError atomic.Value // stores error
	IsRunning atomic.Bool
}

type IScheduler interface {
	AddTask(time string, task ISchedulerTask) (int, error)
	RemoveTask(taskId int) error
	UpdateTask(taskId int, time string, task ISchedulerTask) (int, error)
	Start()
	Stop()
}

type Scheduler struct {
	cron *cron.Cron
	mu   sync.Mutex
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(),
	}
}

/*
decorator pattern
add new task
*/
func (s *Scheduler) wrapTask(task ISchedulerTask) func() {
	return func() {
		status := task.GetStatus()
		status.LastRun.Store(time.Now().UTC())
		status.RunCount.Add(1)
		status.IsRunning.Store(true)
		fmt.Println("執行twse scheduler task")
		defer func() {
			status.IsRunning.Store(false)

			if r := recover(); r != nil {
				status.LastError.Store(fmt.Errorf("panic in task: %v", r))
			}
		}()
		task.RunSchedulerTask()
	}
}

func (s *Scheduler) addTask(time string, task ISchedulerTask) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entityId, err := s.cron.AddFunc(time, s.wrapTask(task))
	if err != nil {
		return 0, err
	}

	return int(entityId), nil
}

func (s *Scheduler) removeTask(taskId int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cron.Remove(cron.EntryID(taskId))

	return nil
}

func (s *Scheduler) AddTask(time string, task ISchedulerTask) (int, error) {
	return s.addTask(time, task)
}

func (s *Scheduler) RemoveTask(taskId int) error {
	return s.removeTask(taskId)
}

// 先移除，再新增
func (s *Scheduler) UpdateTask(taskId int, time string, task ISchedulerTask) (int, error) {
	err := s.removeTask(taskId)
	if err != nil {
		return 0, err
	}

	return s.addTask(time, task)
}

/*
used go routine inside
*/
func (s *Scheduler) Start() {
	go s.cron.Start()
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}
