package asynq_worker

import (
	"context"

	repository "github.com/RoyceAzure/rj/repo/mongo"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	LogQueue = "logQueue"
)

/*
asynq有定義 processor 的interface
*/
type TaskProcessor interface {
	Start() error
	Stop()
	StartWithHandler(handler *asynq.ServeMux) error
}

type WriteLogProcessor struct {
	server   *asynq.Server
	mongoDao repository.IMongoDao
}

/*
server是processot自己建立的? 是host建立asynq server去遠端redis 拿取任務資料
*/
func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, mongoDao repository.IMongoDao) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				LogQueue: 10,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().
					Err(err).
					Str("type", task.Type()).
					Bytes("body", task.Payload()).
					Msg("process task failed")
			}),
			Logger: NewLoggerAdapter(),
		},
	)

	return &WriteLogProcessor{
		server:   server,
		mongoDao: mongoDao,
	}
}

/*
案照這個設計思維   所有processor應可以連到同一個redis server
且每個processor start 就是指start一個路由與handler的對應  且會永久執行?  對  就跟http server很像
一旦執行後就會被block  所以要用go routine執行
asynq 本身processor處理每個task就是使用go routine處理
error是指 sersver startup 過程的錯誤
*/
func (processor *WriteLogProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskWriteLog, processor.WriteLog)
	return processor.server.Start(mux)
}

func (processor *WriteLogProcessor) Stop() {
	processor.server.Shutdown()
}

func (processor *WriteLogProcessor) StartWithHandler(handler *asynq.ServeMux) error {
	return processor.server.Start(handler)
}
