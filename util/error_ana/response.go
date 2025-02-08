package error_ana

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Status  int      `json:"status"`
	Message []string `json:"message,omitempty"`
	Data    any      `json:"data,omitempty"`
}

// jsonResponse 將資料以 jsonResponse 格式寫入 response
func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func ErrorJSON(w http.ResponseWriter, err error) {
	var messages []string
	status := 500

	if err != nil {
		switch v := err.(type) {
		case *AnaError:
			messages = []string{v.Error()}
			status = int(v.Code)
		default:
			messages = []string{"unknow error"}
		}
	}

	if messages == nil {
		messages = []string{}
	}

	jsonResponse(w, status, ErrorResponse{
		Status:  status,
		Message: messages,
	})
}
