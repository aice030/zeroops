package utils

import (
	"encoding/json"
	"net/http"
)

// SetJSONResponse 设置JSON响应
func SetJSONResponse(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// SetErrorResponse 设置错误响应
func SetErrorResponse(w http.ResponseWriter, statusCode int, message string) error {
	return SetJSONResponse(w, statusCode, map[string]any{
		"error":   message,
		"code":    statusCode,
		"success": false,
	})
}
