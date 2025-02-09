package asynq_worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
)

const TaskWriteLog = "task:write_log"

func (taskProcesser *WriteLogProcessor) WriteLog(ctx context.Context, task *asynq.Task) error {
	var payload bson.M

	if taskProcesser.mongoDao == nil {
		return fmt.Errorf("mongo dao has not init %w", asynq.SkipRetry)
	}

	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal task payload %w", asynq.SkipRetry)
	}

	err = taskProcesser.mongoDao.InsertBsonM(ctx, payload)
	if err != nil {
		return err
	}

	return nil
}
