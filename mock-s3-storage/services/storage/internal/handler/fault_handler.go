package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"storage-service/internal/service"
	"strings"
)

type FaultHandler struct {
	faultService service.FaultService
}

func NewFaultHandler(s service.FaultService) *FaultHandler {
	return &FaultHandler{faultService: s}
}

// StartFault 处理 /fault/start/{name} POST 请求
func (h *FaultHandler) StartFault(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}
	name := getNameFromPath(r.URL.Path, "/fault/start/")
	if name == "" {
		http.Error(w, "fault name required", http.StatusBadRequest)
		return
	}
	err := h.faultService.StartFault(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("start fault failed: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// StopFault 处理 /fault/stop/{name} POST 请求
func (h *FaultHandler) StopFault(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}
	name := getNameFromPath(r.URL.Path, "/fault/stop/")
	if name == "" {
		http.Error(w, "fault name required", http.StatusBadRequest)
		return
	}
	err := h.faultService.StopFault(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("stop fault failed: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

// GetFaultStatus 处理 /fault/status/{name} GET 请求
func (h *FaultHandler) GetFaultStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "只支持GET方法", http.StatusMethodNotAllowed)
		return
	}
	name := getNameFromPath(r.URL.Path, "/fault/status/")
	if name == "" {
		http.Error(w, "fault name required", http.StatusBadRequest)
		return
	}
	status, err := h.faultService.GetFaultStatus(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("get fault status failed: %v", err), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": status})
}

// ListFaults 处理 /fault/list GET 请求
func (h *FaultHandler) ListFaults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "只支持GET方法", http.StatusMethodNotAllowed)
		return
	}
	list, err := h.faultService.ListFaults()
	if err != nil {
		http.Error(w, fmt.Sprintf("list faults failed: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"faults": list})
}

// getNameFromPath 用于从路径中提取故障名称
func getNameFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	return strings.TrimPrefix(path, prefix)
}

// writeJSON 统一写json响应
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
