package api

import (
	"encoding/json"
	"net/http"
	"strings"
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

// JSON 將資料以 JSON 格式寫入 response
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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
