package logger_producer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/RoyceAzure/rj/infra/mq/client"
	"github.com/RoyceAzure/rj/logger/internal/model"
)

// used with zero logger
type ILoggerProcuder interface {
	Write(p []byte) (n int, err error)
}

type BaseMQClientLogger struct {
	producer client.IProducer
}

func (bl *BaseMQClientLogger) extractLogEntity(p []byte) (*model.MQLog, error) {
	if p == nil {
		return nil, fmt.Errorf("message is empty")
	}

	var zeroLogData map[string]any
	err := json.Unmarshal(p, &zeroLogData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal log data: %v", err)
	}

	mqLog := &model.MQLog{}

	if t, ok := zeroLogData["time"]; ok {
		if tt, ok := t.(time.Time); ok {
			mqLog.Timestamp = tt
		} else {
			mqLog.Timestamp = time.Now().UTC()
		}
	}

	// 處理字串類型欄位
	if project, ok := zeroLogData["project"].(string); ok {
		mqLog.Project = project
	}
	if module, ok := zeroLogData["module"].(string); ok {
		mqLog.Module = module
	}
	if action, ok := zeroLogData["action"].(string); ok {
		mqLog.Action = action
	}
	if message, ok := zeroLogData["message"].(string); ok {
		mqLog.Message = message
	}
	if level, ok := zeroLogData["level"].(string); ok {
		mqLog.Level = level
	}

	// 處理可空的字串指針欄位
	if userId, ok := zeroLogData["user_id"].(string); ok {
		mqLog.UserID = &userId
	}
	if ipAddr, ok := zeroLogData["ip_address"].(string); ok {
		mqLog.IPAddress = &ipAddr
	}
	if userAgent, ok := zeroLogData["user_agent"].(string); ok {
		mqLog.UserAgent = &userAgent
	}
	if requestId, ok := zeroLogData["request_id"].(string); ok {
		mqLog.RequestID = &requestId
	}
	if errDetails, ok := zeroLogData["error_details"].(string); ok {
		mqLog.ErrorDetails = &errDetails
	}
	if reqMethod, ok := zeroLogData["request_method"].(string); ok {
		mqLog.RequestMethod = &reqMethod
	}
	if reqPath, ok := zeroLogData["request_path"].(string); ok {
		mqLog.RequestPath = &reqPath
	}

	// 處理 ID
	if id, ok := zeroLogData["id"].(float64); ok {
		mqLog.ID = int64(id)
	}

	// 處理 metadata
	if metadata, ok := zeroLogData["metadata"]; ok {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %v", err)
		}
		mqLog.Metadata = metadataBytes
	}

	return mqLog, nil

}
