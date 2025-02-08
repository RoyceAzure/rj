package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

/*
應該要記錄本身worker工作狀態  是否啟動  工作多久...等等

*/

type WorkerStatus int

const (
	Processed WorkerStatus = iota
	Idle
	Sleep
	Stop
)

type WorkerOption func(*Worker)

type Worker struct {
	Id               uuid.UUID
	taskNum          int
	taskBuffer       chan []WorkerTask
	isRunning        bool
	startTime        time.Time
	closeTime        time.Time
	finishedJobCount atomic.Uint64
	status           atomic.Value
	logger           *zerolog.Logger
	WatitTime        time.Duration
}

type IWoker interface {
	Run(context.Context)
	FeedTask([]WorkerTask)
}

/*
@parm

	taskNum: worker buffer大小  表示可以接收幾組任務
	logger : *zerolog.Logger logger di 由worker這層處理
*/
func NewWorker(taskNum int, logger *zerolog.Logger, options ...WorkerOption) *Worker {
	id, _ := uuid.NewUUID()
	taskBuffer := make(chan []WorkerTask, taskNum)
	w := &Worker{
		Id:         id,
		taskNum:    taskNum,
		taskBuffer: taskBuffer,
		isRunning:  false,
		logger:     logger,
	}
	w.status.Store(Idle)

	for _, op := range options {
		op(w)
	}

	return w
}

/*
sleep 應該要有條件
todo HandleResult  寫入DB
*/
func (w *Worker) Run(ctx context.Context) {
	if w.isRunning {
		return
	}
	go func() {
		w.isRunning = true
		w.status.Store(Idle)
		w.startTime = time.Now().UTC()
		for {
			select {
			case <-ctx.Done():
				w.stop()
				return
			case taskList := <-w.taskBuffer:
				w.status.Store(Processed)
				for _, task := range taskList {
					b := task.GetTaskInfo()
					res := task.Excute(ctx)

					// parmInfo := map[string]any{
					// 	"metadata": b,
					// }
					// parmInfoByte, _ := json.Marshal(parmInfo)
					if res.Code == TaskSuccess {
						w.writeWithFileds(w.logger.Debug(), "excute worker task successed", b)
					} else {
						w.writeWithFileds(w.logger.Error(), fmt.Sprintf("excute worker task failed, err : %s", (*res).Error.Error()), b)
					}

					w.finishedJobCount.Add(1)
					task.HandleResult(res)
				}
				w.status.Store(Idle)
				// default:
				// 	w.status.Store(Sleep)
				// 	time.Sleep(time.Second * 10)
			}
			time.Sleep(w.WatitTime)
		}
	}()
}

func (w *Worker) stop() {
	// close(w.taskBuffer)
	fmt.Println("Stop Worker")
	w.status.Store(Stop)
	w.closeTime = time.Now().UTC()
	runDurSecond := w.closeTime.Sub(w.startTime).Seconds()
	fmt.Printf("worker id %s, runs %.2f second, finished job : %d", w.Id, runDurSecond, w.finishedJobCount.Load())
	fmt.Println()
}

/*
 */
func (w *Worker) FeedTask(tasks []WorkerTask) {
	w.taskBuffer <- tasks
}

// writeWithFileds 寫入key-value資訊，並附加上msg
// 參數:
//   - fileds : []byte json string，key value形式。將key-value附加到zero.logger裡面， ex : []byte({"name" : "royce"})
//   - msg : 額外訊息
func (w *Worker) writeWithFileds(l *zerolog.Event, msg string, fileds ...[]byte) {
	if w.logger == nil {
		return
	}
	var data map[string]any

	for _, filed := range fileds {
		err := json.Unmarshal(filed, &data)
		if err != nil {
			l = w.logger.Error()
			msg = "some info extract failed, " + msg
		}
	}

	l.Fields(data)
	l.Msg(msg)
}
