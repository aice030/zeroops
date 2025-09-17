# API接口文档

## Model层 API

### 获取所有服务列表

**请求：**
```http
GET /v1/services
```

**响应：**
```json
{
    "items": {
        {
            "name": "stg", // 服务名称
            "deployState": "InDeploying", // 发布状态：InDeploying|AllDeployFinish
            "health": "Normal", // 健康状态：Normal/Warning/Error
            "deps": ["stg","meta","mq"]
        },
        {
            "name": "meta", // 服务名称
            "deployState": "InDeploying", // 服务发布状态：InDeploying|AllDeployFinish
            "health": "Normal"  // 服务健康状态：Normal/Warning/Error
            "deps": []
        },
    },
    "relation": { // 树形关系描述，有向无环图  
        "s3": ["stg","meta","mq"], // 服务节点，下流服务节点
        "meta": ["mongodb"],
        "mq": ["worker"]
    }
}
```



### 获取服务详情

**请求：**
```http
GET /v1/services/:service/activeVersions
```

**响应：**
```json
{
    "items": [
        {
            "version": "v1.0.1",
            "deployID": "1001", 
            "startTime": "2024-01-01T00:00:00Z", // 开始时间
            "estimatedCompletionTime": "2024-01-01T03:00:00Z", // 预估完成时间
            "instances": 10, // 实例个数（比例由所有items的instance加起来做除法)
            "health": "Normal" // 健康状态：Normal/Warning/Error
        }
    ]
}
```
### 获取服务metrics指标值

**请求：**
```http
GET /v1/metricStats/:service
```
> version对应的状态，通过"获取服务详情"接口获取，前端根据version进行

**响应：**
```json
{
    "summary": { // 所有实例的聚合值
        "metrics": [ // 此版本发布的metric指标内容
            {
                "name": "latency",
                "value": 10 // 单位: ms
            },
            {
                "name": "traffic", // 流量
                "value": 1000 // Qps
            },
            {
                "name": "errorRatio", // 错误率
                "value": 10 
            },
            {
                "name": "saturation", // 饱和度
                "value": 50 // 百分比
            }
        ]
    },
    "items": [
        {
            "version": "v1.0.1",
            "metrics": [ // 此版本发布的metric指标内容
                {
                    "name": "latency",
                    "value": 10 // 单位: ms
                },
                {
                    "name": "traffic", // 流量
                    "value": 1000 // Qps
                },
                {
                    "name": "errorRatio", // 错误率
                    "value": 10 
                },
                {
                    "name": "saturation", // 饱和度
                    "value": 50 // 百分比
                }
            ]
        }
    ]
}
```


### 获取可用服务版本（已发布的排除）

**请求：**
```http
GET /v1/services/:service/availableVersions
```
> 指定服务的可用版本包。接口只返回所有未发布的版本包列表，已发布的版本包不下发。

**响应：**
```json
{
    "items": {
        {
            "version": "v1.0.1", // 版本包对应的版本号
            "createTime": "2024-01-01T03:00:00Z" // 版本包创建时间，RFC3339 格式
        },
        {
            "version": "v1.0.2",
            "createTime": "2024-01-02T03:00:00Z"
        }
    }
}
```
### 新建发布任务

**请求：**
```http
POST /v1/deployments
```

```json
{
    "service": "stg",
    "version": "v1.0.1", // 版本包对应的版本号
    "schedueTime": "2024-01-02T04:00:00Z" // 可选参数，不填为立即发布
}
```
> 接口只返回所有未发布的版本包列表，已发布的版本包不下发。

**响应：**
```json
{
    "id": "1001" // 发布id
}
```
> 同一个版本拒绝多次发布

**错误码：**
- 409: AlreadyInDeployment

### 获取待发布计划列表

**请求：**
```http
GET /v1/deployments?type=Schedule&service=stg  // InDeployment/Schedule/Finished
```
> 获取某个服务的所有未开始的发布任务列表。接口只返回所有未开始/进行中的发布任务列表，已发布的不包含在列表中

**响应：**
```json
{
    "items": {
        {
            "id": "1001",
            "service": "stg",
            "version": "v1.0.1",
            "status": "InDeployment",
            "scheduleTime": "" // 已经发了
        },
        {
            "id": "1002",
            "service": "stg",
            "version": "v1.0.2",
            "status": "InDeployment",
            "scheduleTime": "2024-01-03T05:00:00Z"
        },
        {
            "id": "1003",
            "service": "stg",
            "version": "v1.0.3",
            "status": "InDeployment",
            "finishTime": "2024-01-03T05:00:00Z"
        },
        {
            "id": "1003",
            "service": "stg",
            "version": "v1.0.3",
            "status": "rollbacked",
            "finishTime": "2024-01-03T05:00:00Z"
        }
    }
}
```

### 修改未开始的发布任务

**请求：**
```http
POST /v1/deployments/:deployID
```

```json
{
    "version": "v1.0.2",
    "scheduleTime": "2024-01-03T06:00:00Z" // 新的计划发布时间（当前只能修改此字段）
}
```
> 只能修改还未开始的发布任务

**响应：** 无响应体。200状态码，或异常码

### 删除未开始的发布任务

**请求：**
```http
DELETE /v1/deployments/:deployID
```
> 发布任务未开始的可以直接删除，否则只能暂停或终止。只能删除计划中的发布。已完成、进行中的都不能删除。

**响应：** 无响应体。200状态码，或异常码

### 暂停正在灰度的发布任务

**请求：**
```http
POST /v1/deployments/:deployID/pause
```
> 暂停发布任务。只能处理已经开始灰度，且未完成100%灰度的发布任务

**响应：** 无响应体。200状态码，或异常码

### 继续发布

**请求：**
```http
POST /v1/deployments/:deployID/continue
```

**响应：** 无响应体。200状态码，或异常码

### 回滚发布任务

**请求：**
```http
POST /v1/deployments/:deployID/rollback
```
> 回滚发布任务，它会将此版本已灰度的实例回滚至上一个版本，即使是已完成的发布也能回滚。

**响应：** 无响应体。200状态码，或异常码



### 获取服务指标数据

**请求：**
```http
GET /v1/metrics/:service/:name?version=v1.0.1&start=2024-01-03T06:00:00Z&end=2024-01-03T06:00:00Z&granule=5m(1m/1h)
```
> 按对齐后的时间返回。如果没有指定version，则取的是所有实例的聚合。:name由"获取服务metrics指标值"接口返回的指标列表指定。

**响应：**
> 参考 Prometheus query_range 返回结构体。参考链接：https://prometheus.io/docs/prometheus/latest/querying/api/

```json
{
   "status" : "success",
   "data" : {
      "resultType" : "matrix",
      "result" : [
         {
            "metric" : {
               "__name__" : "up",
               "job" : "prometheus",
               "instance" : "localhost:9090"
            },
            "values" : [
               [ 1435781430.781, "1" ],
               [ 1435781445.781, "1" ],
               [ 1435781460.781, "1" ]
            ]
         },
         {
            "metric" : {
               "__name__" : "up",
               "job" : "node",
               "instance" : "localhost:9091"
            },
            "values" : [
               [ 1435781430.781, "0" ],
               [ 1435781445.781, "0" ],
               [ 1435781460.781, "1" ]
            ]
         }
      ]
   }
}
```


### 获取服务变更记录

**请求：**
```http
GET /v1/changelog/deployment?start=xxxx&limit=10
```
> 从最新的变更记录往前排序。

**响应：**
```json
{
    "items": [
       {
        "serivce": "stg",
        "version": "v1.0.1", // 版本发布id
        "startTime": "2024-01-03T03:00:00Z",
        "endTime": "2024-01-03T06:00:00Z", // 可选参数
        "instances": 50, // 灰度实例个数
        "totalInstances": 100, // 灰度时，总实例个数
        "health": "Normal", // 健康状态：Normal/Warning/Error
      }
    ],
    "next": "xxxx"
}
```



### 获取告警变更记录

**请求：**
```http
GET /v1/changelog/alertrules?start=xxxx&limit=10
```

**响应：**
> 需要参考promethues的rule修改结构

```json
{
    "items": [
        {
            "name": "p98_latency_too_high", // 统一化告警规则的名称 p98_latency_too_high_<:service>
            "editTime": "2024-01-03T03:00:00Z",
            //scope:  serivce:<:service> api:<service1.api1>  idc:<idc> node:<node>
            "scope": "", // 空代表修改所有服务；指定了服务名代表修改指定服务；不确定是否有服务名+version
            "values": [
                {
                    "name": "threshold",
                    "old": "10", // 允许是数字字符串，也可以是其它字符串
                    "new": "15"
                },
                {
                    "name": "watchTimeDuration",
                    "old": "3min",
                    "new": "5min"
                }
            ],
            "reason": "xxx", // AI生成的reason
        }
    ],
    "next": "xxxx"
}
```
> 这里面有继承概念：如果某个服务，它有配置，则用自己的规则；如果没有配置，使用默认统一的规则。版本tag重要，需要指定


### 获取告警事件列表

**请求：**
```http
GET /v1/issues?start=xxxx&limit=10
```
> 可带额外一个query条件state=Closed，不传表示查询所有

**响应：**
```json
{
    "items": [ 
        {
            "id": "xxx", // 告警 issue ID
            "state": "Closed", // 告警条目的状态。Closed处理完成、Open处理中
            "level": "P0", // 枚举值：P0严重、P1重要、P2、Warning需要关注但不是线上异常
            "alertState": "Restored", // 告警处理状态。Restored 已恢复、AutoRestored 系统自动恢复、InProcessing 处理中
            "title": "yzh S3APIV2s3apiv2.putobject 0_64K上传响应时间95值:50012ms > 450ms", // 告警标题
            "labels": [
                {
                    "key": "api",
                    "value": "s3apiv2.putobject"
                },
                {
                    "key": "idc",
                    "value": "yzh"
                }
            ],
            "alertSince": "2025-05-05 11:00:00.0000Z",
            "resolved_at": "" // 告警已恢复, 记录恢复时间
        }
    ],
    "next": "xxxx"
}
```



### 获取某一个告警的处理

**请求：**
```http
GET /v1/issues/:issueID
```

**响应：**
```json
{
    "id": "xxx", // 告警 issue ID
    "state": "Closed", // 告警条目的状态。Closed处理完成、Open处理中
    "level": "P0", // 枚举值：P0严重、P1重要、P2、Warning需要关注但不是线上异常
    "alertState": "Restored", // 告警处理状态。Restored 已恢复、AutoRestored 系统自动恢复、InProcessing 处理中
    "title": "yzh S3APIV2s3apiv2.putobject 0_64K上传响应时间95值:50012ms > 450ms", // 告警标题
    "labels": [
        {
            "key": "api",
            "value": "s3apiv2.putobject"
        },
        {
            "key": "idc",
            "value": "yzh"
        }
    ],
    "alertSince": "2025-05-05 11:00:00.0000Z",
    "comments": [
        {
            "createdAt": "2024-01-03T03:00:00Z",
            "content": "markdown content" // 里面为一个整体的markdown，记录了AI的行为
        }
    ]
}
```
