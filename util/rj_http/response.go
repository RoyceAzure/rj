package rj_http

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data"` //不要使用omitempty  否則空list也會回傳null
}

type ErrorResponse struct {
	Status  int      `json:"status"`
	Message []string `json:"message,omitempty"`
	Data    any      `json:"data,omitempty"`
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

func ErrorJSON(w http.ResponseWriter, status int, data any) {
	var messages []string

	if data != nil {
		switch v := data.(type) {
		case error:
			messages = strings.Split(v.Error(), "\n")
		case string:
			messages = []string{v}
		}
	}

	if messages == nil {
		messages = []string{}
	}

	JSON(w, status, ErrorResponse{
		Status:  status,
		Message: messages,
	})
}

func SuccessJSON(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, Response{
		Status: http.StatusOK,
		Data:   data,
	})
}
