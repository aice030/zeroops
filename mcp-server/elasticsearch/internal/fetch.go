package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GetServiceByHostID 根据host_id获取对应的service名称
// hostID: 主机ID
// startTime: 开始时间，格式为 "2025-08-20T00:00:00Z"
// endTime: 结束时间，格式为 "2025-08-20T23:59:59Z"
func (e *ElasticsearchClient) GetServiceByHostID(hostID string, startTime string, endTime string) (string, error) {
	// 根据查询时间范围生成索引模式，而不是使用当前日期
	indexPatterns, err := e.IndexManager.GenerateIndexPatternsForTimeRange(startTime, endTime, "")
	if err != nil {
		return "", fmt.Errorf("生成索引模式失败: %w", err)
	}

	// 使用第一个索引模式进行查询（简化处理）
	if len(indexPatterns) == 0 {
		return "", fmt.Errorf("未生成有效的索引模式")
	}
	indexPattern := indexPatterns[0]

	// 构建查询，根据host_id查找对应的service
	query := map[string]interface{}{
		"bool": map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"term": map[string]interface{}{
						"host_id": hostID,
					},
				},
				{
					"range": map[string]interface{}{
						"@timestamp": map[string]interface{}{
							"gte": startTime,
							"lte": endTime,
						},
					},
				},
			},
		},
	}

	// 构建搜索请求
	searchRequest := map[string]interface{}{
		"query": query,
		"size":  1, // 只需要一条记录即可
	}

	// 执行查询
	result, err := e.executeSearch(indexPattern, searchRequest)
	if err != nil {
		return "", fmt.Errorf("查询service失败: %w", err)
	}

	// 从查询结果中提取service
	var service string
	if hits, ok := result["hits"].(map[string]interface{}); ok {
		if hitsList, ok := hits["hits"].([]interface{}); ok {
			if len(hitsList) > 0 {
				if hitMap, ok := hitsList[0].(map[string]interface{}); ok {
					if source, ok := hitMap["_source"].(map[string]interface{}); ok {
						if serviceValue, ok := source["service"].(string); ok {
							service = serviceValue
						}
					}
				}
			}
		}
	}

	if service == "" {
		return "", fmt.Errorf("未找到host_id为 %s 的service", hostID)
	}

	return service, nil
}

// FetchLogsByServiceAndHost 根据服务和host_id获取指定时间段内的所有日志
// service: 服务名称，如 "storage-service"，如果为空则自动根据host_id获取
// hostID: 主机ID
// startTime: 开始时间，格式为 "2025-08-20T00:00:00Z"
// endTime: 结束时间，格式为 "2025-08-20T23:59:59Z"
func (e *ElasticsearchClient) FetchLogsByServiceAndHost(service string, hostID string, startTime string, endTime string) (string, error) {
	// 如果service为空，先根据host_id获取service
	if service == "" {
		var err error
		service, err = e.GetServiceByHostID(hostID, startTime, endTime)
		if err != nil {
			return "", fmt.Errorf("根据host_id获取service失败: %w", err)
		}
	}

	// 根据时间范围生成具体的索引名
	indexPatterns, err := e.IndexManager.GenerateIndexPatternsForTimeRange(startTime, endTime, service)
	if err != nil {
		return "", fmt.Errorf("生成索引模式失败: %w", err)
	}

	// 使用第一个索引模式
	if len(indexPatterns) == 0 {
		return "", fmt.Errorf("未生成有效的索引模式")
	}
	indexName := indexPatterns[0]

	// 构建Elasticsearch查询
	query := map[string]interface{}{
		"bool": map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"term": map[string]interface{}{
						"host_id": hostID,
					},
				},
				{
					"range": map[string]interface{}{
						"@timestamp": map[string]interface{}{
							"gte": startTime,
							"lte": endTime,
						},
					},
				},
			},
		},
	}

	// 构建搜索请求
	searchRequest := map[string]interface{}{
		"query": query,
		"size":  1000, // 获取更多日志
		"sort": []map[string]interface{}{
			{
				"@timestamp": map[string]interface{}{
					"order": "asc",
				},
			},
		},
	}

	// 执行查询
	result, err := e.executeSearch(indexName, searchRequest)
	if err != nil {
		return "", fmt.Errorf("查询日志失败: %w", err)
	}

	// 格式化返回结果
	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"service":    service,
			"host_id":    hostID,
			"start_time": startTime,
			"end_time":   endTime,
			"index":      indexName,
			"total_logs": len(result["hits"].(map[string]interface{})["hits"].([]interface{})),
			"logs":       result["hits"].(map[string]interface{})["hits"].([]interface{}),
		},
	}

	responseData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", fmt.Errorf("响应JSON序列化失败: %w", err)
	}

	return string(responseData), nil
}

// FetchRequestTrace 根据request_id追踪请求经过的服务
// requestID: 请求ID
// startTime: 开始时间，格式为 "2025-08-20T00:00:00Z"
// endTime: 结束时间，格式为 "2025-08-20T23:59:59Z"
func (e *ElasticsearchClient) FetchRequestTrace(requestID string, startTime string, endTime string) (string, error) {
	// 根据查询时间范围生成索引模式，查询所有服务的索引
	indexPatterns, err := e.IndexManager.GenerateIndexPatternsForTimeRange(startTime, endTime, "")
	if err != nil {
		return "", fmt.Errorf("生成索引模式失败: %w", err)
	}

	// 使用第一个索引模式进行查询（简化处理）
	if len(indexPatterns) == 0 {
		return "", fmt.Errorf("未生成有效的索引模式")
	}
	indexPattern := indexPatterns[0]

	// 构建跨索引查询
	query := map[string]interface{}{
		"bool": map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"term": map[string]interface{}{
						"request_id": requestID,
					},
				},
				{
					"range": map[string]interface{}{
						"@timestamp": map[string]interface{}{
							"gte": startTime,
							"lte": endTime,
						},
					},
				},
			},
		},
	}

	// 构建搜索请求
	searchRequest := map[string]interface{}{
		"query": query,
		"size":  1000,
		"sort": []map[string]interface{}{
			{
				"@timestamp": map[string]interface{}{
					"order": "asc",
				},
			},
		},
	}

	// 执行跨索引查询
	result, err := e.executeSearch(indexPattern, searchRequest)
	if err != nil {
		return "", fmt.Errorf("查询请求追踪失败: %w", err)
	}

	// 从查询结果中提取服务列表
	var services []string
	serviceMap := make(map[string]bool) // 用于去重

	if hits, ok := result["hits"].(map[string]interface{}); ok {
		if hitsList, ok := hits["hits"].([]interface{}); ok {
			for _, hit := range hitsList {
				if hitMap, ok := hit.(map[string]interface{}); ok {
					if source, ok := hitMap["_source"].(map[string]interface{}); ok {
						if service, ok := source["service"].(string); ok {
							if !serviceMap[service] {
								services = append(services, service)
								serviceMap[service] = true
							}
						}
					}
				}
			}
		}
	}

	// 格式化返回结果
	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"request_id":     requestID,
			"start_time":     startTime,
			"end_time":       endTime,
			"index_pattern":  indexPattern,
			"total_services": len(services),
			"services":       services,
			"total_logs":     len(result["hits"].(map[string]interface{})["hits"].([]interface{})),
			"logs":           result["hits"].(map[string]interface{})["hits"].([]interface{}),
		},
	}

	responseData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", fmt.Errorf("响应JSON序列化失败: %w", err)
	}

	return string(responseData), nil
}

// executeSearch 执行Elasticsearch搜索请求
func (e *ElasticsearchClient) executeSearch(index string, searchRequest map[string]interface{}) (map[string]interface{}, error) {
	// 将查询转换为JSON
	jsonData, err := json.Marshal(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}

	// 构建HTTP请求
	url := fmt.Sprintf("%s/%s/_search", e.BaseURL, index)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{Timeout: e.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查是否有错误
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Elasticsearch返回错误状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}
	return result, nil
}
