package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// mqLog 系統日誌結構
type MQLog struct {
	ID            int64     `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	Module        string    `json:"module"`
	Action        string    `json:"action"`
	Message       string    `json:"message"`
	Level         string    `json:"level"`
	LogFilePath   string    `json:"log_file_path"` //for file log
	UserID        *string   `json:"user_id,omitempty"`
	IPAddress     *string   `json:"ip_address,omitempty"`
	UserAgent     *string   `json:"user_agent,omitempty"`
	RequestID     *string   `json:"request_id,omitempty"`
	Duration      *int32    `json:"duration,omitempty"`
	StatusCode    *int32    `json:"status_code,omitempty"`
	ErrorDetails  *string   `json:"error_details,omitempty"`
	RequestMethod *string   `json:"request_method,omitempty"`
	RequestPath   *string   `json:"request_path,omitempty"`
	ResponseSize  *int32    `json:"response_size,omitempty"`
	Metadata      []byte    `json:"metadata,omitempty"`
}

// LogLevel constants
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

// String 實現 Stringer 接口，返回格式化的日誌字串
func (l MQLog) String() string {
	var sb strings.Builder

	// 基本信息
	sb.WriteString(fmt.Sprintf("%s [%s] ", l.Timestamp.Format("2006-01-02 15:04:05.000"), strings.ToUpper(l.Level)))

	// 模組和動作
	if l.Module != "" {
		sb.WriteString(fmt.Sprintf("[%s] ", l.Module))
	}
	if l.Action != "" {
		sb.WriteString(fmt.Sprintf("[%s] ", l.Action))
	}

	// 主要訊息
	sb.WriteString(l.Message)

	// 請求相關信息
	var details []string

	if l.RequestID != nil {
		details = append(details, fmt.Sprintf("request_id=%s", *l.RequestID))
	}
	if l.RequestMethod != nil && l.RequestPath != nil {
		details = append(details, fmt.Sprintf("method=%s path=%s", *l.RequestMethod, *l.RequestPath))
	}
	if l.StatusCode != nil {
		details = append(details, fmt.Sprintf("status=%d", *l.StatusCode))
	}
	if l.Duration != nil {
		details = append(details, fmt.Sprintf("duration=%dms", *l.Duration))
	}
	if l.ResponseSize != nil {
		details = append(details, fmt.Sprintf("size=%d", *l.ResponseSize))
	}

	// 用戶相關信息
	if l.UserID != nil {
		details = append(details, fmt.Sprintf("user=%s", *l.UserID))
	}
	if l.IPAddress != nil {
		details = append(details, fmt.Sprintf("ip=%s", *l.IPAddress))
	}

	// 錯誤詳情
	if l.ErrorDetails != nil && *l.ErrorDetails != "" {
		details = append(details, fmt.Sprintf("error=%s", *l.ErrorDetails))
	}

	// Metadata
	if l.Metadata != nil {
		metadataJSON, err := json.Marshal(l.Metadata)
		if err == nil {
			details = append(details, fmt.Sprintf("metadata=%s", string(metadataJSON)))
		}
	}

	// 添加詳細信息
	if len(details) > 0 {
		sb.WriteString(" | ")
		sb.WriteString(strings.Join(details, " "))
	}

	return sb.String()
}
