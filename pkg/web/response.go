package web

import (
	"encoding/json"
	"net/http"
)

// ApiResponse: standard response yapısı
type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// RespondJSON: ham payload ile JSON response gönderir
func RespondJSON(w http.ResponseWriter, status int, payload ApiResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// JSON: daha kısa bir wrapper
func JSON(w http.ResponseWriter, status int, success bool, message string, data interface{}) {
	RespondJSON(w, status, ApiResponse{Success: success, Message: message, Data: data})
}

// Ok / Created / NoContent / Error shortcutlar
func Ok(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, true, "OK", data)
}

func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, true, "Created", data)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Error(w http.ResponseWriter, status int, err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	JSON(w, status, false, msg, nil)
}
