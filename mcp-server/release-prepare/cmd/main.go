package main

import (
	"encoding/json"
	"fmt"
	"os"
	"qiniu1024-mcp-server/release-prepare/internal/orchestrator"
)

func main() {
	issue := orchestrator.RunReleasePreparation("user-service", "v1.1.0")

	// 序列化成 JSON
	out, err := json.MarshalIndent(issue, "", "  ")
	if err != nil {
		fmt.Println("json marshal error:", err)
		return
	}

	// 写入文件
	err = os.WriteFile("issue.json", out, 0644)
	if err != nil {
		fmt.Println("write file error:", err)
		return
	}

	fmt.Println("JSON 数据已写入 issue.json")
}
