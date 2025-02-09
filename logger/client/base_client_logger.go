package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/RoyceAzure/rj/infra/mq"
	"github.com/RoyceAzure/rj/logger/server/domain/model"
	"github.com/go-playground/validator/v10"
)

type BaseMQClientLoggerParams struct {
	Exchange   string `validate:"required" json:"exchange"`    // required 表示必須有值
	RoutingKey string `validate:"required" json:"routing_key"` // required 表示必須有值
	Module     string `json:"module"`                          // 非必填
}

func (p *BaseMQClientLoggerParams) Validate() error {
	validate := validator.New()

	if err := validate.Struct(p); err != nil {
		// 處理驗證錯誤
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			for _, e := range validationErrors {
				switch e.Field() {
				case "Exchange":
					return fmt.Errorf("exchange is required")
				case "RoutingKey":
					return fmt.Errorf("routing key is required")
				}
			}
		}
		return err
	}
	return nil
}

type BaseMQClientLogger struct {
	producer mq.IProducer
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
