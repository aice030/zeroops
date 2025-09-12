import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { apiService } from '@/api'

// 服务节点类型定义
export interface ServiceNode {
  id: string
  name: string
  x: number
  y: number
  versions: ServiceVersion[]
}

// 服务版本类型定义
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

// 发布计划类型定义
export interface ReleasePlan {
  id: string
  version: string
  startTime: string
  creator: string
}

// 变更记录类型定义
export interface ChangeItem {
  id: string
  service: string
  version: string
  state: '发布中' | '灰度中' | '已完成'
  progress?: number
  ok?: boolean
  batches?: Batch[]
}

export interface Batch {
  name: string
  status: '正常' | '异常' | '进行中'
  start: string
  end: string
  anomaly?: string
  moduleRecords?: ModuleRecord[]
}

export interface ModuleRecord {
  id: string
  module: string
  action: string
  timestamp: string
  status: '成功' | '失败' | '告警' | '回滚'
  details?: string
  thoughts?: string
  eventData?: any
}

// 告警变更记录类型定义
export interface AlarmChangeItem {
  id: string
  service: string
  change: string
  timestamp: string
  details: string
}

export const useAppStore = defineStore('app', () => {
  // 状态
  const currentView = ref<'home' | 'changelog'>('home')
  const selectedNode = ref<ServiceNode | null>(null)
  const selectedSlice = ref<{ nodeId: string; label: string } | null>(null)
  const metricsOpen = ref(false)
  const selectedVersion = ref('v1.0.7')
  const scheduledStart = ref('')
  const scheduledReleases = ref<ReleasePlan[]>([])
  const editingRelease = ref<{id: string, newTime: string} | null>(null)
  const cancelConfirm = ref<{id: string, version: string} | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 模拟数据 - 服务拓扑
  const nodes = ref<ServiceNode[]>([
    { 
      id: "s3", 
      name: "s3", 
      x: 520, 
      y: 120, 
      versions: [
        { label: "v1.0.0", value: 55, eta: "~ 2h 30m", anomalous: false, observing: false },
        { label: "v1.0.1", value: 30, eta: "~ 1h 10m", anomalous: false, observing: true, rolling: true, elapsedMin: 30, remainingMin: 60 },
        { label: "v1.0.3", value: 15, eta: "~ 40m", anomalous: false, observing: false, rolling: true, elapsedMin: 10, remainingMin: 30 },
      ]
    },
    { 
      id: "stg", 
      name: "stg", 
      x: 340, 
      y: 200, 
      versions: [
        { label: "v1.0.0", value: 70, eta: "~ 3h 00m", anomalous: false, observing: false },
        { label: "v1.0.2", value: 30, eta: "~ 30m", anomalous: false, observing: true, rolling: true, elapsedMin: 15, remainingMin: 20 },
      ]
    },
    { 
      id: "meta", 
      name: "meta", 
      x: 520, 
      y: 260, 
      versions: [
        { label: "v1.0.3", value: 100, eta: "~ 25m", anomalous: false, observing: false },
      ]
    },
    { 
      id: "mq", 
      name: "mq", 
      x: 700, 
      y: 200, 
      versions: [
        { label: "v1.0.1", value: 100, eta: "~ 50m", anomalous: false, observing: false },
      ]
    },
    { 
      id: "worker", 
      name: "worker", 
      x: 820, 
      y: 300, 
      versions: [
        { label: "v1.0.1", value: 100, eta: "~ 20m", anomalous: false, observing: false },
      ]
    },
    { 
      id: "mongodb", 
      name: "mongodb", 
      x: 420, 
      y: 380, 
      versions: [
        { label: "v1.0.1", value: 100, eta: "~ 1h 10m", anomalous: true, observing: false },
      ]
    },
  ])

  const edges = ref([
    { source: "s3", target: "stg" },
    { source: "s3", target: "meta" },
    { source: "s3", target: "mq" },
    { source: "mq", target: "worker" },
    { source: "meta", target: "mongodb" },
  ])

  // 计算属性
  const statusColor = computed(() => ({
    healthy: "bg-emerald-500",
    abnormal: "bg-rose-500", 
    canary: "bg-amber-500"
  }))

  const statusStroke = computed(() => ({
    healthy: "#10b981",
    abnormal: "#f43f5e",
    canary: "#f59e0b"
  }))

  const statusFill = computed(() => ({
    healthy: "#10b981",
    abnormal: "#f43f5e", 
    canary: "#f59e0b"
  }))

  // 方法
  const statusFromVersions = (versions: ServiceVersion[]) => {
    return versions.some(v => v.anomalous) ? 'abnormal' : 
           (versions.some(v => v.observing) ? 'canary' : 'healthy')
  }

  const setView = (view: 'home' | 'changelog') => {
    currentView.value = view
  }

  const setSelectedNode = (node: ServiceNode | null) => {
    selectedNode.value = node
    if (!node) {
      selectedSlice.value = null
    }
  }

  const setSelectedSlice = (slice: { nodeId: string; label: string } | null) => {
    selectedSlice.value = slice
  }

  const setMetricsOpen = (open: boolean) => {
    metricsOpen.value = open
  }

  const setSelectedVersion = (version: string) => {
    selectedVersion.value = version
  }

  const setScheduledStart = (time: string) => {
    scheduledStart.value = time
  }


  return {
    // 状态
    currentView,
    selectedNode,
    selectedSlice,
    metricsOpen,
    selectedVersion,
    scheduledStart,
    scheduledReleases,
    editingRelease,
    cancelConfirm,
    loading,
    error,
    nodes,
    edges,
    
    // 计算属性
    statusColor,
    statusStroke,
    statusFill,
    
    // 方法
    statusFromVersions,
    setView,
    setSelectedNode,
    setSelectedSlice,
    setMetricsOpen,
    setSelectedVersion,
    setScheduledStart
  }
})
