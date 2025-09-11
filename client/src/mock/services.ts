// Mock数据 - 服务拓扑数据
// 根据我们商议的方案：使用dependencies数组记录依赖关系

export interface ServiceItem {
  name: string
  deployState: 'InDeploying' | 'AllDeployFinish'
  health: 'Normal' | 'Warning' | 'Error'
  dependencies: string[]
}

export interface ServicesResponse {
  items: ServiceItem[]
}

// Mock数据 - 根据当前写死的数据转换而来
export const mockServicesData: ServicesResponse = {
  items: [
    {
      name: "s3",
      deployState: "InDeploying",
      health: "Warning",
      dependencies: [] // s3是根节点，无依赖
    },
    {
      name: "stg",
      deployState: "InDeploying",
      health: "Warning",
      dependencies: ["s3"] // stg依赖s3
    },
    {
      name: "meta",
      deployState: "AllDeployFinish",
      health: "Normal",
      dependencies: ["s3"] // meta依赖s3
    },
    {
      name: "mq",
      deployState: "AllDeployFinish",
      health: "Normal",
      dependencies: ["s3"] // mq依赖s3
    },
    {
      name: "worker",
      deployState: "AllDeployFinish",
      health: "Normal",
      dependencies: ["mq"] // worker依赖mq
    },
    {
      name: "mongodb",
      deployState: "AllDeployFinish",
      health: "Error",
      dependencies: ["meta"] // mongodb依赖meta
    }
  ]
}

// Mock数据 - 服务版本详情（用于弹窗显示）
export interface ServiceVersion {
  label: string
  value: number
  eta: string
  anomalous: boolean
  observing: boolean
  rolling?: boolean
  elapsedMin?: number
  remainingMin?: number
}

export interface ServiceDetail {
  name: string
  deployState: 'InDeploying' | 'AllDeployFinish'
  health: 'Normal' | 'Warning' | 'Error'
  dependencies: string[]
  versions: ServiceVersion[]
}

export const mockServiceDetails: Record<string, ServiceDetail> = {
  "s3": {
    name: "s3",
    deployState: "InDeploying",
    health: "Normal",
    dependencies: [],
    versions: [
      { label: "v1.0.0", value: 55, eta: "~ 2h 30m", anomalous: false, observing: false },
      { label: "v1.0.1", value: 30, eta: "~ 1h 10m", anomalous: false, observing: true, rolling: true, elapsedMin: 30, remainingMin: 60 },
      { label: "v1.0.3", value: 15, eta: "~ 40m", anomalous: false, observing: false, rolling: true, elapsedMin: 10, remainingMin: 30 }
    ]
  },
  "stg": {
    name: "stg",
    deployState: "InDeploying",
    health: "Normal",
    dependencies: ["s3"],
    versions: [
      { label: "v1.0.0", value: 70, eta: "~ 3h 00m", anomalous: false, observing: false },
      { label: "v1.0.2", value: 30, eta: "~ 30m", anomalous: false, observing: true, rolling: true, elapsedMin: 15, remainingMin: 20 }
    ]
  },
  "meta": {
    name: "meta",
    deployState: "AllDeployFinish",
    health: "Normal",
    dependencies: ["s3"],
    versions: [
      { label: "v1.0.3", value: 100, eta: "~ 25m", anomalous: false, observing: false }
    ]
  },
  "mq": {
    name: "mq",
    deployState: "AllDeployFinish",
    health: "Normal",
    dependencies: ["s3"],
    versions: [
      { label: "v1.0.1", value: 100, eta: "~ 50m", anomalous: false, observing: false }
    ]
  },
  "worker": {
    name: "worker",
    deployState: "AllDeployFinish",
    health: "Normal",
    dependencies: ["mq"],
    versions: [
      { label: "v1.0.1", value: 100, eta: "~ 20m", anomalous: false, observing: false }
    ]
  },
  "mongodb": {
    name: "mongodb",
    deployState: "AllDeployFinish",
    health: "Error",
    dependencies: ["meta"],
    versions: [
      { label: "v1.0.1", value: 100, eta: "~ 1h 10m", anomalous: true, observing: false }
    ]
  }
}

// 可发布版本数据结构 - 匹配后端API返回格式
export interface AvailableVersion {
  version: string
  createTime: string
}

export interface AvailableVersionsResponse {
  items: AvailableVersion[]
}

// Mock数据 - 可发布版本列表
export const mockAvailableVersions: Record<string, AvailableVersionsResponse> = {
  "s3": {
    items: [
      { version: "v1.0.4", createTime: "2024-01-01T03:00:00Z" },
      { version: "v1.0.5", createTime: "2024-01-02T03:00:00Z" },
      { version: "v1.0.6", createTime: "2024-01-03T03:00:00Z" },
      { version: "v1.0.7", createTime: "2024-01-04T03:00:00Z" },
      { version: "v1.0.8", createTime: "2024-01-05T03:00:00Z" }
    ]
  },
  "stg": {
    items: [
      { version: "v1.0.3", createTime: "2024-01-01T02:00:00Z" },
      { version: "v1.0.4", createTime: "2024-01-02T02:00:00Z" },
      { version: "v1.0.5", createTime: "2024-01-03T02:00:00Z" },
      { version: "v1.0.6", createTime: "2024-01-04T02:00:00Z" }
    ]
  },
  "meta": {
    items: [
      { version: "v1.0.4", createTime: "2024-01-01T01:00:00Z" },
      { version: "v1.0.5", createTime: "2024-01-02T01:00:00Z" },
      { version: "v1.0.6", createTime: "2024-01-03T01:00:00Z" }
    ]
  },
  "mq": {
    items: [
      { version: "v1.0.2", createTime: "2024-01-01T00:30:00Z" },
      { version: "v1.0.3", createTime: "2024-01-02T00:30:00Z" },
      { version: "v1.0.4", createTime: "2024-01-03T00:30:00Z" }
    ]
  },
  "worker": {
    items: [
      { version: "v1.0.2", createTime: "2024-01-01T00:15:00Z" },
      { version: "v1.0.3", createTime: "2024-01-02T00:15:00Z" },
      { version: "v1.0.4", createTime: "2024-01-03T00:15:00Z" },
      { version: "v1.0.5", createTime: "2024-01-04T00:15:00Z" }
    ]
  },
  "mongodb": {
    items: [
      { version: "v1.0.2", createTime: "2024-01-01T00:10:00Z" },
      { version: "v1.0.3", createTime: "2024-01-02T00:10:00Z" }
    ]
  }
}

// 兼容性：保留原有的版本选项格式（用于全局版本选择）
export const mockVersionOptions = [
  { label: 'v1.0.4', value: 'v1.0.4' },
  { label: 'v1.0.5', value: 'v1.0.5' },
  { label: 'v1.0.6', value: 'v1.0.6' },
  { label: 'v1.0.7', value: 'v1.0.7' }
]

// 发布计划数据结构 - 匹配后端API返回格式
export interface DeploymentPlan {
  id: string
  service: string
  version: string
  status: 'Schedule' | 'InDeployment' | 'Finished'
  scheduleTime?: string
  finishTime?: string
  isPaused?: boolean
}

export interface DeploymentPlansResponse {
  items: DeploymentPlan[]
}

// Mock数据 - 发布计划（按服务分组）
export const mockDeploymentPlans: Record<string, DeploymentPlansResponse> = {
  "s3": {
    items: [
      {
        id: "1001",
        service: "s3",
        version: "v1.0.4",
        status: "InDeployment",
        scheduleTime: "2024-01-15T14:00:00Z",
        isPaused: false
      },
      {
        id: "1002",
        service: "s3",
        version: "v1.0.5",
        status: "InDeployment",
        scheduleTime: "2024-01-16T10:00:00Z",
        isPaused: true
      },
      {
        id: "1003",
        service: "s3",
        version: "v1.0.3",
        status: "Finished",
        scheduleTime: "2024-01-14T09:00:00Z",
        finishTime: "2024-01-14T18:00:00Z"
      }
    ]
  },
  "stg": {
    items: [
      {
        id: "2001",
        service: "stg",
        version: "v1.0.1",
        status: "InDeployment",
        scheduleTime: "",
        isPaused: false
      },
      {
        id: "2002",
        service: "stg",
        version: "v1.0.2",
        status: "InDeployment",
        scheduleTime: "2024-01-03T05:00:00Z",
        isPaused: false
      },
      {
        id: "2003",
        service: "stg",
        version: "v1.0.3",
        status: "Finished",
        finishTime: "2024-01-03T05:00:00Z"
      }
    ]
  },
  "meta": {
    items: [
      {
        id: "3001",
        service: "meta",
        version: "v1.0.4",
        status: "InDeployment",
        scheduleTime: "2024-01-17T09:00:00Z"
      },
      {
        id: "3002",
        service: "meta",
        version: "v1.0.5",
        status: "Finished",
        finishTime: "2024-01-16T15:30:00Z"
      }
    ]
  },
  "mq": {
    items: [
      {
        id: "4001",
        service: "mq",
        version: "v1.0.2",
        status: "InDeployment",
        scheduleTime: "2024-01-18T12:00:00Z"
      },
      {
        id: "4002",
        service: "mq",
        version: "v1.0.3",
        status: "Finished",
        finishTime: "2024-01-17T20:00:00Z"
      }
    ]
  },
  "worker": {
    items: [
      {
        id: "5001",
        service: "worker",
        version: "v1.0.2",
        status: "InDeployment",
        scheduleTime: "2024-01-19T08:00:00Z"
      },
      {
        id: "5002",
        service: "worker",
        version: "v1.0.3",
        status: "Finished",
        finishTime: "2024-01-18T16:00:00Z"
      },
      {
        id: "5003",
        service: "worker",
        version: "v1.0.4",
        status: "InDeployment",
        scheduleTime: ""
      }
    ]
  },
  "mongodb": {
    items: [
      {
        id: "6001",
        service: "mongodb",
        version: "v1.0.2",
        status: "InDeployment",
        scheduleTime: "2024-01-20T14:00:00Z"
      },
      {
        id: "6002",
        service: "mongodb",
        version: "v1.0.3",
        status: "Finished",
        finishTime: "2024-01-19T11:00:00Z"
      }
    ]
  }
}

// 兼容性：保留原有的发布计划格式（用于全局发布计划）
export const mockScheduledReleases = [
  {
    id: 'release-1',
    version: 'v1.0.4',
    startTime: '2024-01-15 14:00:00',
    creator: '张三'
  },
  {
    id: 'release-2',
    version: 'v1.0.5',
    startTime: '2024-01-15 16:00:00',
    creator: '李四'
  }
]

// 指标API数据结构 - 匹配Prometheus query_range格式
export interface MetricData {
  metric: {
    __name__: string
    job?: string
    instance?: string
    version?: string
    service?: string
  }
  values: Array<[number, string]> // [timestamp, value]
}

export interface MetricsResponse {
  status: 'success' | 'error'
  data: {
    resultType: 'matrix' | 'vector' | 'scalar' | 'string'
    result: MetricData[]
  }
}

// 部署变更记录数据类型定义
export interface DeploymentChangelogItem {
  service: string
  version: string
  startTime: string
  endTime?: string
  instances: number
  totalInstances: number
  health: 'Normal' | 'Warning' | 'Error'
}

export interface DeploymentChangelogResponse {
  items: DeploymentChangelogItem[]
  next?: string
}

// 扩展的部署变更记录数据类型，包含详细的分批次信息
export interface DetailedDeploymentChangelogItem extends DeploymentChangelogItem {
  batches?: {
    name: string
    status: '正常' | '异常' | '进行中'
    start: string
    end: string
    anomaly?: string
    moduleRecords?: {
      id: string
      module: string
      action: string
      timestamp: string
      status: '成功' | '失败' | '告警' | '回滚'
      details?: string
      eventData?: any
    }[]
  }[]
}

// Mock数据 - 部署变更记录（包含详细的分批次信息）
export const mockDeploymentChangelog: { items: DetailedDeploymentChangelogItem[], next?: string } = {
  items: [
    {
      service: "stg",
      version: "v1.0.3",
      startTime: "2024-01-15T14:00:00Z",
      endTime: "2024-01-15T16:00:00Z",
      instances: 50,
      totalInstances: 100,
      health: "Warning",
      batches: [
        { 
          name: '第一批', 
          status: '正常', 
          start: '2024-01-15 14:00:00', 
          end: '2024-01-15 14:10:00' 
        },
        { 
          name: '第二批', 
          status: '正常', 
          start: '2024-01-15 14:20:00', 
          end: '2024-01-15 14:35:00' 
        },
        {
          name: '第三批',
          status: '异常',
          start: '2024-01-15 14:50:00',
          end: '2024-01-15 15:10:00',
          anomaly: 'Stg服务指标异常。',
          moduleRecords: [
            {
              id: 'event-1',
              module: '发布系统',
              action: '发布到节点',
              timestamp: '2024-01-15 14:50:00',
              status: '成功',
              details: '发布到 bj1-node-002, qn1-node-002 节点'
            },
            {
              id: 'event-2',
              module: '发布系统',
              action: '部署新版本',
              timestamp: '2024-01-15 14:50:15',
              status: '成功',
              eventData: {
                deployment: {
                  service: "Stg服务v1.0.3",
                  environment: "生产环境",
                  healthCheck: "通过",
                  loadBalancer: "已更新路由规则",
                  trafficStatus: "新版本开始接收流量"
                }
              }
            },
            {
              id: 'event-3',
              module: '监控告警系统',
              action: '灰度发布监控',
              timestamp: '2024-01-15 14:55:20',
              status: '告警',
              eventData: {
                monitoringAlert: {
                  service: "Stg服务",
                  phase: "灰度过程中",
                  issue: "延迟异常"
                },
                metrics: {
                  p95Latency: {
                    before: "120ms",
                    after: "450ms",
                    increase: "275%"
                  },
                  errorRate: {
                    before: "0.1%",
                    after: "2.3%",
                    increase: "2200%"
                  }
                }
              }
            },
            {
              id: 'event-4',
              module: '发布系统',
              action: '自动回滚',
              timestamp: '2024-01-15 15:05:45',
              status: '回滚',
              eventData: {
                rollbackTrigger: {
                  trigger: "AI检测到异常后自动触发回滚流程",
                  aiJudgment: "基于延迟和错误率指标，AI判断当前版本存在严重问题"
                }
              }
            }
          ]
        }
      ]
    },
    {
      service: "stg", 
      version: "v1.0.2",
      startTime: "2024-01-14T10:00:00Z",
      endTime: "2024-01-14T12:00:00Z",
      instances: 30,
      totalInstances: 100,
      health: "Normal",
      batches: [
        { 
          name: '第一批', 
          status: '正常', 
          start: '2024-01-14 10:00:00', 
          end: '2024-01-14 10:15:00' 
        },
        { 
          name: '第二批', 
          status: '正常', 
          start: '2024-01-14 10:15:00', 
          end: '2024-01-14 10:30:00' 
        },
        { 
          name: '第三批', 
          status: '正常', 
          start: '2024-01-14 10:30:00', 
          end: '2024-01-14 10:45:00' 
        }
      ]
    },
    {
      service: "stg",
      version: "v1.0.1", 
      startTime: "2024-01-13T08:00:00Z",
      endTime: "2024-01-13T10:00:00Z",
      instances: 100,
      totalInstances: 100,
      health: "Normal",
      batches: [
        { 
          name: '第一批', 
          status: '正常', 
          start: '2024-01-13 08:00:00', 
          end: '2024-01-13 08:20:00' 
        },
        { 
          name: '第二批', 
          status: '正常', 
          start: '2024-01-13 08:20:00', 
          end: '2024-01-13 08:40:00' 
        },
        { 
          name: '第三批', 
          status: '正常', 
          start: '2024-01-13 08:40:00', 
          end: '2024-01-13 09:00:00' 
        }
      ]
    },
    {
      service: "s3",
      version: "v1.0.4",
      startTime: "2024-01-15T14:00:00Z",
      instances: 25,
      totalInstances: 50,
      health: "Normal",
      batches: [
        { 
          name: '第一批', 
          status: '正常', 
          start: '2024-01-15 14:00:00', 
          end: '2024-01-15 14:10:00' 
        },
        { 
          name: '第二批', 
          status: '进行中', 
          start: '2024-01-15 14:10:00', 
          end: '-' 
        }
      ]
    },
    {
      service: "s3",
      version: "v1.0.3",
      startTime: "2024-01-14T18:00:00Z",
      endTime: "2024-01-14T20:00:00Z", 
      instances: 50,
      totalInstances: 50,
      health: "Normal",
      batches: [
        { 
          name: '第一批', 
          status: '正常', 
          start: '2024-01-14 18:00:00', 
          end: '2024-01-14 18:15:00' 
        },
        { 
          name: '第二批', 
          status: '正常', 
          start: '2024-01-14 18:15:00', 
          end: '2024-01-14 18:30:00' 
        },
        { 
          name: '第三批', 
          status: '正常', 
          start: '2024-01-14 18:30:00', 
          end: '2024-01-14 18:45:00' 
        }
      ]
    }
  ],
  next: "2024-01-15T14:00:00Z"
}

// Mock指标数据 - 为每个服务的每个指标生成数据
export const mockMetricsData: Record<string, Record<string, MetricsResponse>> = {
  "s3": {
    "latency": {
      status: "success",
      data: {
        resultType: "matrix",
        result: [
          {
            metric: {
              __name__: "latency",
              service: "s3",
              version: "v1.0.1",
              instance: "s3-01"
            },
            values: [
              [1704268800, "45"], // 2024-01-03 06:00:00
              [1704269100, "42"], // 2024-01-03 06:05:00
              [1704269400, "48"], // 2024-01-03 06:10:00
              [1704269700, "41"], // 2024-01-03 06:15:00
              [1704270000, "44"], // 2024-01-03 06:20:00
              [1704270300, "46"]  // 2024-01-03 06:25:00
            ]
          }
        ]
      }
    },
    "traffic": {
      status: "success",
      data: {
        resultType: "matrix",
        result: [
          {
            metric: {
              __name__: "traffic",
              service: "s3",
              version: "v1.0.1",
              instance: "s3-01"
            },
            values: [
              [1704268800, "1200"],
              [1704269100, "1180"],
              [1704269400, "1250"],
              [1704269700, "1190"],
              [1704270000, "1210"],
              [1704270300, "1230"]
            ]
          }
        ]
      }
    },
    "errors": {
      status: "success",
      data: {
        resultType: "matrix",
        result: [
          {
            metric: {
              __name__: "errors",
              service: "s3",
              version: "v1.0.1",
              instance: "s3-01"
            },
            values: [
              [1704268800, "2.5"],
              [1704269100, "2.1"],
              [1704269400, "2.8"],
              [1704269700, "2.3"],
              [1704270000, "2.4"],
              [1704270300, "2.6"]
            ]
          }
        ]
      }
    },
    "saturation": {
      status: "success",
      data: {
        resultType: "matrix",
        result: [
          {
            metric: {
              __name__: "saturation",
              service: "s3",
              version: "v1.0.1",
              instance: "s3-01"
            },
            values: [
              [1704268800, "75"],
              [1704269100, "72"],
              [1704269400, "78"],
              [1704269700, "74"],
              [1704270000, "76"],
              [1704270300, "77"]
            ]
          }
        ]
      }
    }
  },
  "stg": {
    "latency": {
      status: "success",
      data: {
        resultType: "matrix",
        result: [
          {
            metric: {
              __name__: "latency",
              service: "stg",
              version: "v1.0.1",
              instance: "stg-01"
            },
            values: [
              [1704268800, "38"],
              [1704269100, "35"],
              [1704269400, "41"],
              [1704269700, "36"],
              [1704270000, "39"],
              [1704270300, "40"]
            ]
          }
        ]
      }
    },
    "traffic": {
      status: "success",
      data: {
        resultType: "matrix",
        result: [
          {
            metric: {
              __name__: "traffic",
              service: "stg",
              version: "v1.0.1",
              instance: "stg-01"
            },
            values: [
              [1704268800, "950"],
              [1704269100, "920"],
              [1704269400, "980"],
              [1704269700, "940"],
              [1704270000, "960"],
              [1704270300, "970"]
            ]
          }
        ]
      }
    },
    "errors": {
      status: "success",
      data: {
        resultType: "matrix",
        result: [
          {
            metric: {
              __name__: "errors",
              service: "stg",
              version: "v1.0.1",
              instance: "stg-01"
            },
            values: [
              [1704268800, "1.8"],
              [1704269100, "1.5"],
              [1704269400, "2.1"],
              [1704269700, "1.7"],
              [1704270000, "1.9"],
              [1704270300, "2.0"]
            ]
          }
        ]
      }
    },
    "saturation": {
      status: "success",
      data: {
        resultType: "matrix",
        result: [
          {
            metric: {
              __name__: "saturation",
              service: "stg",
              version: "v1.0.1",
              instance: "stg-01"
            },
            values: [
              [1704268800, "68"],
              [1704269100, "65"],
              [1704269400, "71"],
              [1704269700, "67"],
              [1704270000, "69"],
              [1704270300, "70"]
            ]
          }
        ]
      }
    }
  },
  "meta": {
    "latency": {
      status: "success",
      data: {
        resultType: "matrix",
        result: [
          {
            metric: {
              __name__: "latency",
              service: "meta",
              version: "v1.0.1",
              instance: "meta-01"
            },
            values: [
              [1704268800, "52"],
              [1704269100, "49"],
              [1704269400, "55"],
              [1704269700, "50"],
              [1704270000, "53"],
              [1704270300, "54"]
            ]
          }
        ]
      }
    },
    "traffic": {
      status: "success",
      data: {
        resultType: "matrix",
        result: [
          {
            metric: {
              __name__: "traffic",
              service: "meta",
              version: "v1.0.1",
              instance: "meta-01"
            },
            values: [
              [1704268800, "800"],
              [1704269100, "780"],
              [1704269400, "820"],
              [1704269700, "790"],
              [1704270000, "810"],
              [1704270300, "815"]
            ]
          }
        ]
      }
    },
    "errors": {
      status: "success",
      data: {
        resultType: "matrix",
        result: [
          {
            metric: {
              __name__: "errors",
              service: "meta",
              version: "v1.0.1",
              instance: "meta-01"
            },
            values: [
              [1704268800, "3.2"],
              [1704269100, "2.9"],
              [1704269400, "3.5"],
              [1704269700, "3.0"],
              [1704270000, "3.3"],
              [1704270300, "3.4"]
            ]
          }
        ]
      }
    },
    "saturation": {
      status: "success",
      data: {
        resultType: "matrix",
        result: [
          {
            metric: {
              __name__: "saturation",
              service: "meta",
              version: "v1.0.1",
              instance: "meta-01"
            },
            values: [
              [1704268800, "82"],
              [1704269100, "79"],
              [1704269400, "85"],
              [1704269700, "81"],
              [1704270000, "83"],
              [1704270300, "84"]
            ]
          }
        ]
      }
    }
  }
}

// 服务活跃版本数据结构 - 匹配后端API返回格式
export interface ServiceActiveVersion {
  version: string
  deployID: string
  startTime: string
  estimatedCompletionTime: string
  instances: number
  health: 'Normal' | 'Warning' | 'Error'
}

export interface ServiceActiveVersionsResponse {
  items: ServiceActiveVersion[]
}

// Mock数据 - 服务活跃版本数据
export const mockServiceActiveVersions: Record<string, ServiceActiveVersionsResponse> = {
  "s3": {
    items: [
      {
        version: "v1.0.0",
        deployID: "deploy-001",
        startTime: "2024-01-01T00:00:00Z",
        estimatedCompletionTime: "2024-01-01T02:30:00Z",
        instances: 55,
        health: "Normal"
      },
      {
        version: "v1.0.1",
        deployID: "deploy-002",
        startTime: "2024-01-01T01:00:00Z",
        estimatedCompletionTime: "2024-01-01T02:10:00Z",
        instances: 30,
        health: "Warning"
      },
      {
        version: "v1.0.3",
        deployID: "deploy-003",
        startTime: "2024-01-01T01:30:00Z",
        estimatedCompletionTime: "2024-01-01T02:10:00Z",
        instances: 15,
        health: "Normal"
      }
    ]
  },
  "stg": {
    items: [
      {
        version: "v1.0.0",
        deployID: "deploy-004",
        startTime: "2024-01-01T00:00:00Z",
        estimatedCompletionTime: "2024-01-01T03:00:00Z",
        instances: 70,
        health: "Normal"
      },
      {
        version: "v1.0.2",
        deployID: "deploy-005",
        startTime: "2024-01-01T01:15:00Z",
        estimatedCompletionTime: "2024-01-01T01:45:00Z",
        instances: 30,
        health: "Warning"
      }
    ]
  },
  "meta": {
    items: [
      {
        version: "v1.0.3",
        deployID: "deploy-006",
        startTime: "2024-01-01T00:00:00Z",
        estimatedCompletionTime: "2024-01-01T00:25:00Z",
        instances: 100,
        health: "Normal"
      }
    ]
  },
  "mq": {
    items: [
      {
        version: "v1.0.1",
        deployID: "deploy-007",
        startTime: "2024-01-01T00:00:00Z",
        estimatedCompletionTime: "2024-01-01T00:50:00Z",
        instances: 100,
        health: "Normal"
      }
    ]
  },
  "worker": {
    items: [
      {
        version: "v1.0.1",
        deployID: "deploy-008",
        startTime: "2024-01-01T00:00:00Z",
        estimatedCompletionTime: "2024-01-01T00:20:00Z",
        instances: 100,
        health: "Normal"
      }
    ]
  },
  "mongodb": {
    items: [
      {
        version: "v1.0.1",
        deployID: "deploy-009",
        startTime: "2024-01-01T00:00:00Z",
        estimatedCompletionTime: "2024-01-01T01:10:00Z",
        instances: 100,
        health: "Error"
      }
    ]
  }
}

// 指标数据结构 - 匹配后端返回格式
export interface Metric {
  name: string
  value: number
}

export interface ServiceMetricsSummary {
  metrics: Metric[]
}

export interface ServiceMetricsItem {
  version: string
  metrics: Metric[]
}

export interface ServiceMetricsResponse {
  summary: ServiceMetricsSummary
  items: ServiceMetricsItem[]
}

// Mock数据 - 服务指标数据
export const mockServiceMetrics: Record<string, ServiceMetricsResponse> = {
  "s3": {
    summary: {
      metrics: [
        { name: "latency", value: 15 },
        { name: "traffic", value: 1200 },
        { name: "errorRatio", value: 2 },
        { name: "saturation", value: 65 }
      ]
    },
    items: [
      {
        version: "v1.0.0",
        metrics: [
          { name: "latency", value: 13 },
          { name: "traffic", value: 800 },
          { name: "errorRatio", value: 1 },
          { name: "saturation", value: 55 }
        ]
      },
      {
        version: "v1.0.1",
        metrics: [
          { name: "latency", value: 18 },
          { name: "traffic", value: 300 },
          { name: "errorRatio", value: 3 },
          { name: "saturation", value: 70 }
        ]
      },
      {
        version: "v1.0.3",
        metrics: [
          { name: "latency", value: 20 },
          { name: "traffic", value: 100 },
          { name: "errorRatio", value: 5 },
          { name: "saturation", value: 80 }
        ]
      }
    ]
  },
  "stg": {
    summary: {
      metrics: [
        { name: "latency", value: 8 },
        { name: "traffic", value: 2000 },
        { name: "errorRatio", value: 0.5 },
        { name: "saturation", value: 45 }
      ]
    },
    items: [
      {
        version: "v1.0.0",
        metrics: [
          { name: "latency", value: 7 },
          { name: "traffic", value: 1400 },
          { name: "errorRatio", value: 0.3 },
          { name: "saturation", value: 40 }
        ]
      },
      {
        version: "v1.0.2",
        metrics: [
          { name: "latency", value: 10 },
          { name: "traffic", value: 600 },
          { name: "errorRatio", value: 0.8 },
          { name: "saturation", value: 55 }
        ]
      }
    ]
  },
  "meta": {
    summary: {
      metrics: [
        { name: "latency", value: 5 },
        { name: "traffic", value: 5000 },
        { name: "errorRatio", value: 0.1 },
        { name: "saturation", value: 30 }
      ]
    },
    items: [
      {
        version: "v1.0.3",
        metrics: [
          { name: "latency", value: 5 },
          { name: "traffic", value: 5000 },
          { name: "errorRatio", value: 0.1 },
          { name: "saturation", value: 30 }
        ]
      }
    ]
  },
  "mq": {
    summary: {
      metrics: [
        { name: "latency", value: 3 },
        { name: "traffic", value: 8000 },
        { name: "errorRatio", value: 0.05 },
        { name: "saturation", value: 25 }
      ]
    },
    items: [
      {
        version: "v1.0.1",
        metrics: [
          { name: "latency", value: 3 },
          { name: "traffic", value: 8000 },
          { name: "errorRatio", value: 0.05 },
          { name: "saturation", value: 25 }
        ]
      }
    ]
  },
  "worker": {
    summary: {
      metrics: [
        { name: "latency", value: 25 },
        { name: "traffic", value: 500 },
        { name: "errorRatio", value: 1.5 },
        { name: "saturation", value: 75 }
      ]
    },
    items: [
      {
        version: "v1.0.1",
        metrics: [
          { name: "latency", value: 25 },
          { name: "traffic", value: 500 },
          { name: "errorRatio", value: 1.5 },
          { name: "saturation", value: 75 }
        ]
      }
    ]
  },
  "mongodb": {
    summary: {
      metrics: [
        { name: "latency", value: 50 },
        { name: "traffic", value: 200 },
        { name: "errorRatio", value: 15 },
        { name: "saturation", value: 90 }
      ]
    },
    items: [
      {
        version: "v1.0.1",
        metrics: [
          { name: "latency", value: 50 },
          { name: "traffic", value: 200 },
          { name: "errorRatio", value: 15 },
          { name: "saturation", value: 90 }
        ]
      }
    ]
  }
}
