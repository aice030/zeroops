# Prometheus MCP Server

## prometheus\_query

**描述：**&#x6267;行指定的 PromQL 查询语句，并获取对应地区的 Prometheus 监控指标数据

**参数列表：**

{

   `"promql"`(string, 必需)：PromQL 查询语句

   `"regionCode"`(string, 可选)：地区代码，用于映射查询的 Prometheus 地址。默认值为 "mock"，对应 "localhost"。

}

**返回值格式：**

```json
{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {
          "__name__": "process_cpu_seconds_total",
          "instance": "localhost:9090",
          "job": "prometheus"
        },
        "value": [1640995200, "0.001234"]
      }
    ]
  }
}
```

# ElasticSearch MCP Servier

## elasticsearch\_get\_service

**描述：**;询指定主机在给定时间区间内运行的服务

**参数列表：**

{

  ` "host_id"`(string, 必需)：主机ID

   `"start_time"`(string, 必需)：查询区间的起始时间，格式为 '2025-08-20T00:00:00Z'，工具支持格式转换

   `"end_time"`(string, 必需)：查询区间的结束时间，格式为 '2025-08-20T00:00:00Z'，工具支持格式转换

}

**返回值格式：**

```json
{
  "status": "success",
  "data": {
    "host_id": "server-001",
    "service": "storage-service",
    "start_time": "2025-08-20T00:00:00Z",
    "end_time": "2025-08-20T23:59:59Z"
  }
}
```

## elasticsearch\_fetch\_logs

**描述：**&#x6839;据服务名称和主机 ID，获取指定时间段内的运行日志

**参数列表：**

{

   `"service"`(string, 可选, 推荐)：服务名称，若未指定，则需额外根据主机查询服务。

   `"host_id"`(string, 必需)：主机ID

   `"start_time"`(string, 必需)：查询区间的起始时间，格式为 '2025-08-20T00:00:00Z'，工具支持格式转换

   `"end_time"`(string, 必需)：查询区间的结束时间，格式为 '2025-08-20T00:00:00Z'，工具支持格式转换

}

**返回值格式：**

```json
{
  "status": "success",
  "data": {
    "service": "storage-service",
    "host_id": "server-001",
    "start_time": "2025-08-20T00:00:00Z",
    "end_time": "2025-08-20T23:59:59Z",
    "index": "mock-storage-service-logs-2025.08.20",
    "total_logs": 150,
    "logs": [
      {
        "_index": "mock-storage-service-logs-2025.08.20",
        "_source": {
          "@timestamp": "2025-08-20T10:30:00Z",
          "host_id": "server-001",
          "service": "storage-service",
          "level": "INFO",
          "message": "Request processed successfully"
        }
      },
      ......
    ]
  }
}
```

## elasticsearch\_request\_trace

**描述：**;据请求 ID，追踪该请求在指定时间段内经过的所有服务，并获取相关运行日志，构建请求链路

**实现方式：**;具会基于请求ID和起止时间，查询请求在指定时间段内经过的所有服务及其日志记录。随后根据处理顺序对服务进行排序，从而得到该请求的上下游关系。

**参数列表：**

{

   `"request_id"`(string, 必需)：请求ID

  ` "start_time"`(string, 必需)：查询区间的结束时间，格式为 '2025-08-20T00:00:00Z'，工具支持格式转换

   `"end_time"`(string, 必需)：查询区间的结束时间，格式为 '2025-08-20T00:00:00Z'，工具支持格式转换

}

**返回值格式：**

```json
{
  "status": "success",
  "data": {
    "request_id": "req-12345",
    "start_time": "2025-08-20T00:00:00Z",
    "end_time": "2025-08-20T23:59:59Z",
    "index_pattern": "mock-storage-service-logs-2025.08.20",
    "total_services": 3,
    "services": ["api-gateway", "user-service", "database-service"],
    "total_logs": 25,
    "logs": [...]
  }
}
```

