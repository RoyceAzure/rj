package worker

import (
	"context"
	"encoding/json"
)

type TaskResultCode int

const (
	TaskSuccess TaskResultCode = iota
	TaskFailed
)

type BaseTaskInfo struct {
	TaskName string `json:"task_name"`
	Module   string `json:"module"`
	Action   string `json:"action"`
}

type TaskResult struct {
	TaskName string `json:"task_name"`
	Code     TaskResultCode
	Response any
	Error    error
	Parm     any
}

type TaskInfo struct {
	BaseTaskInfo
	Metadata json.RawMessage `json:"metadata"` // 使用 json.RawMessage 來保持原始的 JSON 編碼
}

type WorkerTask interface {
	Excute(context.Context) *TaskResult
	HandleResult(*TaskResult)
	GetTaskInfo() []byte
}
