package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Logger interface {
	Info(ctx context.Context, msg string, args ...interface{})
	Error(ctx context.Context, msg string, args ...interface{})
	Warn(ctx context.Context, msg string, args ...interface{})
}

type DefaultLogger struct {
	logger *log.Logger
}

func (l *DefaultLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.logger.Printf("[INFO] "+msg, args...)
}

func (l *DefaultLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+msg, args...)
}

func (l *DefaultLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.logger.Printf("[WARN] "+msg, args...)
}

type WorkerStatus int

const (
	Running WorkerStatus = iota //啟用狀態
	Idle                        //啟用狀態但是無任務在處理
	Stop                        //停止狀態
)

type WorkerOption struct {
	TaskNum int
	Logger  Logger
}

//需要能回傳當前worker執行狀態與各種數據，讓distributor能夠決定要分配給哪個worker
// 要有清理功能

// 數據:
// 平均執行時間
// 當前待處理任務數，用於附載均衡
// 完成任務數
// 啟動時間，關閉時間，運行時間

type WorkerStaticsData struct {
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	CurrentTaskCount     int           `json:"current_task_count"`
	FinishedJobCount     uint64        `json:"finished_job_count"`
	StartTime            time.Time     `json:"start_time"`
	CloseTime            time.Time     `json:"close_time"`
	RunTime              time.Duration `json:"run_time"`
	Status               WorkerStatus  `json:"status"`
}

type IWoker interface {
	Run(context.Context)
	FeedTask([]WorkerTask)
	GetStatusData() WorkerStaticsData
	Reset()
}

type Worker struct {
	statics      *WorkerStaticsData
	Id           uuid.UUID
	option       WorkerOption
	taskBuffer   chan []WorkerTask
	isProcessing atomic.Bool  //標記內部迴圈是否執行
	status       atomic.Value //worker本身狀態
	logger       Logger
}

/*
@parm

	taskNum: worker buffer大小  表示可以接收幾組任務
	logger : *zerolog.Logger logger di 由worker這層處理
*/
func NewWorker(option WorkerOption) *Worker {
	id, _ := uuid.NewUUID()
	w := &Worker{
		Id:      id,
		statics: &WorkerStaticsData{},
		option:  option,
	}
	w.status.Store(Stop)
	w.SetOption(option)
	return w
}

func (w *Worker) SetOption(options WorkerOption) {
	if options.TaskNum == 0 {
		options.TaskNum = 10
	}
	w.taskBuffer = make(chan []WorkerTask, options.TaskNum)

	if options.Logger == nil {
		w.logger = &DefaultLogger{
			logger: log.New(os.Stdout, "", log.LstdFlags),
		}
	} else {
		w.logger = options.Logger
	}
}

func (w *Worker) GetStatusData() WorkerStaticsData {
	w.updateStaticData()
	return *w.statics
}

// update status FinishedJobCount and AverageExecutionTime
//
//	  param:
//		  dur 當前完成任務的執行時間
func (w *Worker) finishedWork(dur time.Duration) {
	w.statics.FinishedJobCount++
	if w.statics.AverageExecutionTime == 0 {
		w.statics.AverageExecutionTime = dur
	} else {
		w.statics.AverageExecutionTime = (w.statics.AverageExecutionTime*time.Duration(w.statics.FinishedJobCount) + dur) / time.Duration(w.statics.FinishedJobCount)
	}
}

func (w *Worker) updateStaticData() {
	w.statics.CurrentTaskCount = len(w.taskBuffer)
	if w.statics.CloseTime.IsZero() {
		w.statics.RunTime = time.Since(w.statics.StartTime)
	} else {
		w.statics.RunTime = w.statics.CloseTime.Sub(w.statics.StartTime)
	}
	w.statics.Status = w.status.Load().(WorkerStatus)
}

func (w *Worker) Reset() {
	w.statics = &WorkerStaticsData{}
	w.status.Store(Stop)
	w.isProcessing.Store(false)
	w.SetOption(w.option)
	w.taskBuffer = make(chan []WorkerTask, w.option.TaskNum)
}

func (w *Worker) GetTaskBuffer() chan []WorkerTask {
	return w.taskBuffer
}

/*
統一交由外部由ctx控制是否終止
或者內部發生致命錯誤而中止
*/
func (w *Worker) Run(ctx context.Context) {
	//只有當前狀態為Stop或者沒有任務在處理時，才會啟動
	if w.status.Load().(WorkerStatus) != Stop || w.isProcessing.Load() {
		return
	}

	w.statics.StartTime = time.Now().UTC()
	w.status.Store(Running)

	go func() {
		w.isProcessing.Store(true)
		defer w.isProcessing.Store(false)

		for {
			select {
			case <-ctx.Done():
				w.stop()
				return
			case taskList := <-w.taskBuffer:
				if w.status.Load().(WorkerStatus) == Stop {
					return
				}
				w.status.Store(Running)
				for _, task := range taskList {
					err := w.processTask(ctx, task)
					if errors.Is(err, context.Canceled) {
						//任務執行到一半中斷，將任務放回buffer
						w.taskBuffer <- []WorkerTask{task}
						break
					}
				}
			case <-time.After(time.Second * 5):
				w.status.CompareAndSwap(Running, Idle)
			}
		}
	}()
}

// 執行一個任務，並更新worker狀態
func (w *Worker) processTask(ctx context.Context, task WorkerTask) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	start := time.Now().UTC()
	b := task.GetTaskInfo()
	res := task.Excute(ctx)

	if res.Code == TaskSuccess {
		w.logger.Info(ctx, "excute worker task successed", b)
	} else {
		w.logger.Error(ctx, fmt.Sprintf("excute worker task failed, err : %s", (*res).Error.Error()), b)
	}

	w.finishedWork(time.Since(start))
	task.HandleResult(res)
	return nil
}

// worker 結束時，將buffer裡面的任務全部執行完
// 任務若沒有在10秒內執行完，則會保留在buffer裡面
func (w *Worker) flush() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			w.logger.Warn(ctx, "worker flush task buffer timeout, task buffer size: "+strconv.Itoa(len(w.taskBuffer)))
			return
		case taskList := <-w.taskBuffer:
			for _, task := range taskList {
				w.processTask(context.Background(), task)
			}
		default:
			w.logger.Info(ctx, "worker flush task buffer done")
			return
		}
	}
}

func (w *Worker) stop() {
	w.status.Store(Stop)
	close(w.taskBuffer)
	w.statics.CloseTime = time.Now().UTC()
	w.flush()
}

/*
 */
func (w *Worker) FeedTask(tasks []WorkerTask) {
	w.taskBuffer <- tasks
}
