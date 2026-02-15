package api

import (
	"net/http"
	"strings"

	"github.com/goccy/go-json"
)

type Response struct {
	Success  bool      `json:"success"`
	MetaData *MetaData `json:"meta,omitempty"`
	Data     any       `json:"data"` //不要使用omitempty 否則空list也會回傳null
}

type MetaData struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int64 `json:"total_count"`
}

type FailedResponse struct {
	Success       bool          `json:"success"`
	ResponseError ResponseError `json:"error"`
}

type ResponseError struct {
	Code    int      `json:"code"`
	Message *string  `json:"message,omitempty"`
	Details []string `json:"details,omitempty"`
}

// JSONOption JSON 編碼選項（Option Pattern）
type JSONOption func(*jsonEncoderConfig)

type jsonEncoderConfig struct {
	escapeHTML bool
	onError    func(http.ResponseWriter, error)
}

// WithEscapeHTML 設定是否轉義 HTML（預設為 true，符合標準庫行為）
func WithEscapeHTML(escape bool) JSONOption {
	return func(cfg *jsonEncoderConfig) {
		cfg.escapeHTML = escape
	}
}

// WithErrorHandler 設定錯誤處理函數
func WithErrorHandler(handler func(http.ResponseWriter, error)) JSONOption {
	return func(cfg *jsonEncoderConfig) {
		cfg.onError = handler
	}
}

// JSON 將資料以 JSON 格式寫入 response
// 使用 Option Pattern 支援靈活的配置
func JSON(w http.ResponseWriter, status int, data any, options ...JSONOption) {
	// 預設配置
	cfg := &jsonEncoderConfig{
		escapeHTML: true, // 預設開啟 HTML 轉義（向後兼容）
		onError: func(w http.ResponseWriter, err error) {
			// 預設錯誤處理：嘗試設置錯誤狀態（如果 WriteHeader 尚未調用）
			// 注意：如果 WriteHeader 已經調用，這裡不會生效，只能記錄日誌
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		},
	}

	// 應用選項
	for _, opt := range options {
		opt(cfg)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(cfg.escapeHTML)

	if err := enc.Encode(data); err != nil {
		// 使用配置的錯誤處理函數
		// 注意：WriteHeader 已經調用，某些錯誤處理可能無法生效
		if cfg.onError != nil {
			cfg.onError(w, err)
		}
		return
	}
}

// JSONWithEscapeHTML 將資料以 JSON 格式寫入 response，關閉 HTML 轉義以提升效能
// 這是 JSON 函數的便捷包裝，向後兼容
// 錯誤處理：WriteHeader 已調用，錯誤時僅返回（不嘗試設置錯誤狀態）
func JSONWithEscapeHTML(w http.ResponseWriter, status int, data any) {
	JSON(w, status, data, 
		WithEscapeHTML(false),
		WithErrorHandler(func(w http.ResponseWriter, err error) {
			// 注意：WriteHeader 已經呼叫過，這裡若報錯通常只能記錄 Log
			// 不嘗試設置錯誤狀態，避免重複調用 WriteHeader
		}),
	)
}
func ErrorJSON(w http.ResponseWriter, status int, err error, errMsg string) {
	var details []string
	if err != nil {
		details = strings.Split(err.Error(), "\n")
	}

	JSON(w, status, FailedResponse{
		Success: false,
		ResponseError: ResponseError{
			Code:    status,
			Message: &errMsg,
			Details: details,
		},
	})
}

// 帶元數據的成功響應
func SuccessJSON(w http.ResponseWriter, data any, metaData *MetaData) {
	JSON(w, http.StatusOK, Response{
		Success:  true,
		MetaData: metaData,
		Data:     data,
	})
}

// 不帶元數據的成功響應
func SuccessJSONWithoutMeta(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}
