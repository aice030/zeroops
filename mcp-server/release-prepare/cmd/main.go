package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"qiniu1024-mcp-server/release-prepare/internal/orchestrator"
)

// ServiceInfo 定义接收的结构体
type ServiceInfo struct {
	ServiceName string `json:"serviceName"`
	Version     string `json:"version"`
}

// CORS中间件函数
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// 处理预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 调用下一个处理器
		next(w, r)
	}
}

// 接受服务名称与版本
func validateServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var info ServiceInfo
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&info)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("收到服务信息: %+v\n", info)

	// 调用其他方法
	issue := orchestrator.RunReleasePreparation(info.ServiceName, info.Version)

	// 序列化成 JSON
	issueJSON, err := json.MarshalIndent(issue, "", "  ")
	if err != nil {
		fmt.Println("json marshal error:", err)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"status":      "ok",
		"serviceName": info.ServiceName,
		"version":     info.Version,
		"result":      string(issueJSON),
	}
	json.NewEncoder(w).Encode(resp)
}

func Run() {
	http.HandleFunc("/validate-service", enableCORS(validateServiceHandler))
	fmt.Println("Server running on http://localhost:8070")
	log.Fatal(http.ListenAndServe(":8070", nil))
}
