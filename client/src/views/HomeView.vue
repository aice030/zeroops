<template>
  <div class="home-container">
    <!-- 顶部导航 -->
    <div class="header">
      <div class="title">Zero Ops</div>
      <el-button type="primary" @click="goToChangelog">
        系统变更记录
      </el-button>
    </div>

    <div class="subtitle">整体服务状态</div>

    <!-- 服务关系图 -->
    <el-card class="topology-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon><Connection /></el-icon>
          <span>服务关系图</span>
        </div>
      </template>
      
      <div class="topology-container">
        <!-- 加载状态 -->
        <div v-if="loading" class="loading-container">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>加载服务数据中...</span>
        </div>
        
        <!-- 错误状态 -->
        <div v-else-if="error" class="error-container">
          <el-icon><Warning /></el-icon>
          <span>{{ error }}</span>
          <el-button size="small" @click="loadServicesData">重试</el-button>
        </div>
        
        <!-- SVG 连接线 -->
        <svg class="edges-svg">
          <defs>
            <template v-for="(edge, idx) in edges" :key="`arrow-${idx}`">
              <marker
                :id="`arrow-${idx}`"
                markerWidth="8"
                markerHeight="8"
                refX="8"
                refY="4"
                orient="auto"
              >
                <path d="M0,0 L0,8 L8,4 z" fill="#94a3b8" />
              </marker>
            </template>
          </defs>
          <template v-for="(edge, idx) in edges" :key="idx">
            <line
              :x1="getNodePosition(edge.source).x"
              :y1="getNodePosition(edge.source).y"
              :x2="getNodePosition(edge.target).x"
              :y2="getNodePosition(edge.target).y"
              stroke="#cbd5e1"
              stroke-width="2"
              :marker-end="`url(#arrow-${idx})`"
            />
          </template>
        </svg>

        <!-- 服务节点 -->
        <div
          v-for="node in nodes"
          :key="node.id"
          class="service-node"
          :style="{
            left: `${node.x}px`,
            top: `${node.y}px`
          }"
          @click="handleNodeClick(node)"
        >
          <div 
            class="node-circle"
            :style="{ backgroundColor: getNodeStatusColor(node) }"
          >
            <span class="node-name">{{ node.name }}</span>
            <!-- 灰度发布指示器 -->
            <div 
              v-if="hasRollingVersion(node)"
              class="rolling-indicator"
            >
              <svg viewBox="0 0 24 24">
                <circle
                  cx="12"
                  cy="12"
                  r="10"
                  fill="none"
                  :stroke="getNodeStatusStroke(node)"
                  stroke-width="2"
                />
                <path
                  d="M 12 2 A 10 10 0 0 1 12 22"
                  :fill="getNodeStatusFill(node)"
                  opacity="0.3"
                />
              </svg>
            </div>
          </div>
        </div>

        <!-- 图例说明 -->
        <div class="legend">
          <div class="legend-title">图例说明</div>
          <div class="legend-item">
            <span class="legend-dot" style="background-color: #10b981;"></span>
            <span>：服务正常</span>
          </div>
          <div class="legend-item">
            <span class="legend-dot" style="background-color: #f59e0b;"></span>
            <span>：有异常，AI正在观察和分析</span>
          </div>
          <div class="legend-item">
            <span class="legend-dot" style="background-color: #f43f5e;"></span>
            <span>：服务有异常</span>
          </div>
          <div class="legend-divider"></div>
          <div class="legend-item">
            <div class="rolling-example">
              <svg viewBox="0 0 24 24">
                <circle cx="12" cy="12" r="10" fill="none" stroke="#10b981" stroke-width="2" />
                <path d="M 12 2 A 10 10 0 0 1 12 22" fill="#10b981" opacity="0.3" />
              </svg>
            </div>
            <span>：服务正在灰度发布中</span>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 服务详情弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      title="服务概览"
      width="90%"
      :fullscreen="false"
      @close="handleCloseDialog"
    >
      <div v-if="selectedNode" class="service-detail">
        <div class="detail-header">
          <span>{{ selectedNode.name }} 服务概览</span>
          <el-tag :type="getStatusType(getNodeStatus(selectedNode))">
            {{ getNodeStatusText(getNodeStatus(selectedNode)) }}
          </el-tag>
        </div>
        <p class="detail-desc">点击饼图扇区进入指标看板（四大黄金指标）。</p>
        
        <div class="detail-content">
          <!-- 饼图 -->
          <div class="pie-chart-container">
            <div ref="pieChartRef" class="pie-chart"></div>
          </div>
          
          <!-- 版本信息表格 -->
          <div class="version-info">
            <div class="version-title">各版本：延迟、流量、错误、饱和度（示例数据）</div>
            <el-table :data="getVersionTableData(selectedNode.versions)" size="small" class="version-table">
              <el-table-column prop="version" label="版本" width="80" />
              <el-table-column prop="latency" label="延迟" width="80" />
              <el-table-column prop="traffic" label="流量" width="80" />
              <el-table-column prop="errors" label="错误" width="80" />
              <el-table-column prop="saturation" label="饱和度" width="80" />
              <el-table-column prop="status" label="状态" width="80">
                <template #default="{ row }">
                  <el-tag 
                    :type="getStatusType(row.status)"
                    size="small"
                  >
                    {{ row.status }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
            
            <!-- 发布管理 -->
            <div class="release-controls">
              <el-select v-model="selectedVersion" placeholder="选择目标版本" style="width: 250px">
                <el-option
                  v-for="option in availableVersionOptions"
                  :key="option.value"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
              
              <el-date-picker
                v-model="scheduledStart"
                type="datetime"
                placeholder="发布起始时间"
                format="YYYY-MM-DD HH:mm:ss"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width: 200px"
              />
              
              <el-button type="primary" @click="createRelease">新建发布</el-button>
              
              <el-popover placement="bottom" width="300" trigger="click">
                <template #reference>
                  <el-button>
                    <el-icon><Clock /></el-icon>
                    计划
                  </el-button>
                </template>
                <div class="release-plans">
                  <div class="plans-header">{{ selectedNode.name }} 发布计划列表</div>
                  <div v-if="deploymentPlansForDisplay.length > 0" class="plans-list">
                    <div
                      v-for="plan in deploymentPlansForDisplay"
                      :key="plan.id"
                      class="plan-item"
                    >
                      <div class="plan-header">
                        <span class="plan-version">{{ plan.version }}</span>
                        <div class="plan-actions" v-if="plan.originalStatus === 'InDeployment'">
                          <el-button size="small" @click="editRelease(plan)">编辑</el-button>
                          <el-button size="small" type="danger" @click="confirmCancel(plan)">取消</el-button>
                        </div>
                      </div>
                      <div class="plan-details">
                        <div>时间: {{ plan.time }}</div>
                        <div>状态: 
                          <el-tag 
                            :type="plan.originalStatus === 'completed' ? 'success' : 
                                   plan.originalStatus === 'rollbacked' ? 'danger' : 'warning'"
                            size="small"
                          >
                            {{ plan.status }}
                          </el-tag>
                        </div>
                      </div>
                    </div>
                  </div>
                  <div v-else class="no-plans">暂无发布计划</div>
                </div>
              </el-popover>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Loading, Warning } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { apiService } from '@/api'
import { mockApi } from '@/mock/api'
import type { ServicesResponse, ServiceDetail, ServiceActiveVersionsResponse, ServiceMetricsResponse, AvailableVersionsResponse, DeploymentPlansResponse } from '@/mock/services'

const router = useRouter()

// 响应式数据
const dialogVisible = ref(false)
const selectedNode = ref<any>(null)
const selectedVersion = ref('v1.0.7')
const scheduledStart = ref('')
const scheduledReleases = ref<any[]>([])
const pieChartRef = ref<HTMLElement>()

// 服务数据
const servicesData = ref<ServicesResponse | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)

// 存储当前服务的指标数据
const currentServiceMetrics = ref<ServiceMetricsResponse | null>(null)

// 存储当前服务的可发布版本列表
const currentServiceAvailableVersions = ref<AvailableVersionsResponse | null>(null)

// 存储当前服务的发布计划列表
const currentServiceDeploymentPlans = ref<DeploymentPlansResponse | null>(null)

// 自动布局算法
const calculateAutoLayout = (services: any[]) => {
  // 布局配置
  const layoutConfig = {
    levelHeight: 150,      // 层级间距
    nodeSpacing: 200,      // 同层节点间距
    startX: 400,           // 起始X坐标
    startY: 100,           // 起始Y坐标
    maxNodesPerLevel: 6    // 每层最大节点数
  }
  
  // 1. 构建依赖图
  const dependencyGraph = new Map<string, string[]>()
  const reverseGraph = new Map<string, string[]>()
  
  services.forEach(service => {
    dependencyGraph.set(service.name, service.dependencies || [])
    reverseGraph.set(service.name, [])
  })
  
  // 构建反向图
  services.forEach(service => {
    service.dependencies?.forEach((dep: string) => {
      if (reverseGraph.has(dep)) {
        reverseGraph.get(dep)!.push(service.name)
      }
    })
  })
  
  // 2. 拓扑排序确定层级
  const levels: string[][] = []
  const visited = new Set<string>()
  const inDegree = new Map<string, number>()
  
  // 计算入度
  services.forEach(service => {
    inDegree.set(service.name, service.dependencies?.length || 0)
  })
  
  // 找到所有入度为0的节点（根节点）
  let currentLevel: string[] = []
  inDegree.forEach((degree, serviceName) => {
    if (degree === 0) {
      currentLevel.push(serviceName)
    }
  })
  
  // 分层处理
  while (currentLevel.length > 0) {
    levels.push([...currentLevel])
    console.log(`层级 ${levels.length - 1}:`, currentLevel)
    const nextLevel: string[] = []
    
    currentLevel.forEach(serviceName => {
      visited.add(serviceName)
      // 找到依赖当前服务的所有服务
      const dependents = reverseGraph.get(serviceName) || []
      dependents.forEach(dependent => {
        if (!visited.has(dependent)) {
          const currentDegree = inDegree.get(dependent) || 0
          inDegree.set(dependent, currentDegree - 1)
          if (inDegree.get(dependent) === 0) {
            nextLevel.push(dependent)
          }
        }
      })
    })
    
    currentLevel = nextLevel
  }
  
  console.log('自动布局层级结构:', levels)
  
  // 3. 计算位置
  const positions = new Map<string, {x: number, y: number}>()
  
  levels.forEach((level, levelIndex) => {
    const levelY = layoutConfig.startY + levelIndex * layoutConfig.levelHeight
    const levelWidth = (level.length - 1) * layoutConfig.nodeSpacing
    const startX = layoutConfig.startX - levelWidth / 2
    
    console.log(`层级 ${levelIndex} 布局:`, {
      level,
      levelY,
      levelWidth,
      startX
    })
    
    level.forEach((serviceName, nodeIndex) => {
      const x = startX + nodeIndex * layoutConfig.nodeSpacing
      positions.set(serviceName, { x, y: levelY })
      console.log(`  ${serviceName}: (${x}, ${levelY})`)
    })
  })
  
  console.log('最终位置映射:', positions)
  return positions
}

// 数据转换函数
const transformServiceData = (data: ServicesResponse) => {
  const nodes: any[] = []
  const edges: any[] = []
  
  // 使用自动布局算法计算位置
  const positions = calculateAutoLayout(data.items)
  
  // 转换服务节点
  data.items.forEach((service) => {
    const position = positions.get(service.name) || { x: 400, y: 100 }
    
    const node = {
      id: service.name,
      name: service.name,
      x: position.x,
      y: position.y,
      health: service.health,
      deployState: service.deployState,
      dependencies: service.dependencies,
      // 根据发布状态生成版本信息
      versions: generateVersionsFromDeployState(service)
    }
    nodes.push(node)
    
    // 生成依赖关系边
    service.dependencies.forEach(dep => {
      edges.push({
        source: service.name,
        target: dep
      })
    })
  })
  
  return { nodes, edges }
}

// 根据发布状态生成版本信息
const generateVersionsFromDeployState = (service: any) => {
  if (service.deployState === 'InDeploying') {
    // 发布中：生成多个版本，其中一个在发布
    return [
      { label: "v1.0.0", value: 70, eta: "~ 2h 30m", anomalous: false, observing: false },
      { label: "v1.0.1", value: 30, eta: "~ 1h 10m", anomalous: false, observing: true, rolling: true, elapsedMin: 30, remainingMin: 60 }
    ]
  } else {
    // 发布完成：只有一个稳定版本
    const isError = service.health === 'Error'
    return [
      { 
        label: "v1.0.0", 
        value: 100, 
        eta: "~ 2h 30m", 
        anomalous: isError, 
        observing: false 
      }
    ]
  }
}

// API调用函数
const loadServicesData = async () => {
  loading.value = true
  error.value = null
  
  try {
    // 并行加载服务数据和版本选项
    const [servicesResponse, versionsResponse] = await Promise.all([
      mockApi.getServices(),
      mockApi.getVersionOptions()
    ])
    
    servicesData.value = servicesResponse
    versionOptions.value = versionsResponse
    
    // 转换数据
    const { nodes: transformedNodes, edges: transformedEdges } = transformServiceData(servicesResponse)
    nodes.value = transformedNodes
    edges.value = transformedEdges
    
    console.log('服务数据加载成功:', servicesResponse)
    console.log('版本选项加载成功:', versionsResponse)
  } catch (err) {
    error.value = '加载服务数据失败'
    console.error('加载服务数据失败:', err)
    ElMessage.error('加载服务数据失败')
  } finally {
    loading.value = false
  }
}

// 数据转换函数：将后端返回的活跃版本数据转换为前端需要的格式
const transformActiveVersionsToFrontend = (activeVersionsResponse: ServiceActiveVersionsResponse) => {
  const totalInstances = activeVersionsResponse.items.reduce((sum, item) => sum + item.instances, 0)
  
  return activeVersionsResponse.items.map(item => {
    // 计算百分比
    const percentage = totalInstances > 0 ? Math.round((item.instances / totalInstances) * 100) : 0
    
    // 计算ETA（预估剩余时间）
    const startTime = new Date(item.startTime)
    const estimatedCompletion = new Date(item.estimatedCompletionTime)
    const now = new Date()
    const remainingMs = estimatedCompletion.getTime() - now.getTime()
    const remainingMinutes = Math.max(0, Math.round(remainingMs / (1000 * 60)))
    const eta = remainingMinutes > 60 ? `~ ${Math.round(remainingMinutes / 60)}h ${remainingMinutes % 60}m` : `~ ${remainingMinutes}m`
    
    // 计算已用时间
    const elapsedMs = now.getTime() - startTime.getTime()
    const elapsedMinutes = Math.max(0, Math.round(elapsedMs / (1000 * 60)))
    
    // 判断是否在发布中（基于时间）
    const isRolling = remainingMinutes > 0 && elapsedMinutes > 0
    
    // 状态映射
    const isAnomalous = item.health === 'Error'
    const isObserving = item.health === 'Warning'
    
    return {
      label: item.version,
      value: percentage,
      eta: eta,
      anomalous: isAnomalous,
      observing: isObserving,
      rolling: isRolling,
      elapsedMin: elapsedMinutes,
      remainingMin: remainingMinutes,
      deployID: item.deployID,
      startTime: item.startTime,
      estimatedCompletionTime: item.estimatedCompletionTime,
      instances: item.instances,
      health: item.health
    }
  })
}

// 获取服务详情 - 使用新的API接口
const loadServiceDetail = async (serviceName: string) => {
  try {
    // 调用新的活跃版本API
    const activeVersionsResponse = await mockApi.getServiceActiveVersions(serviceName)
    
    // 转换数据格式
    const transformedVersions = transformActiveVersionsToFrontend(activeVersionsResponse)
    
    return {
      name: serviceName,
      versions: transformedVersions
    }
  } catch (err) {
    console.error('获取服务活跃版本失败:', err)
    ElMessage.error('获取服务详情失败')
    return null
  }
}

// 获取服务指标数据 - 使用新的API接口
const loadServiceMetrics = async (serviceName: string) => {
  try {
    // 调用新的指标API
    const metricsResponse = await mockApi.getServiceMetrics(serviceName)
    return metricsResponse
  } catch (err) {
    console.error('获取服务指标数据失败:', err)
    ElMessage.error('获取服务指标数据失败')
    return null
  }
}

// 获取服务可发布版本列表 - 使用新的API接口
const loadServiceAvailableVersions = async (serviceName: string) => {
  try {
    // 调用新的可发布版本API
    const availableVersionsResponse = await mockApi.getServiceAvailableVersions(serviceName)
    return availableVersionsResponse
  } catch (err) {
    console.error('获取服务可发布版本失败:', err)
    ElMessage.error('获取服务可发布版本失败')
    return null
  }
}

// 获取服务发布计划列表 - 使用新的API接口
const loadServiceDeploymentPlans = async (serviceName: string) => {
  try {
    // 调用新的发布计划API
    const deploymentPlansResponse = await mockApi.getServiceDeploymentPlans(serviceName)
    return deploymentPlansResponse
  } catch (err) {
    console.error('获取服务发布计划失败:', err)
    ElMessage.error('获取服务发布计划失败')
    return null
  }
}

// 生命周期
onMounted(() => {
  loadServicesData()
  loadScheduledReleases()
})

// 服务拓扑数据（通过API获取）
const nodes = ref<any[]>([])
const edges = ref<any[]>([])

// 版本选项（通过API获取）
const versionOptions = ref<Array<{label: string, value: string}>>([])

// 计算属性：将可发布版本数据转换为下拉框格式
const availableVersionOptions = computed(() => {
  if (!currentServiceAvailableVersions.value) {
    return []
  }
  
  return currentServiceAvailableVersions.value.items.map(item => ({
    label: `${item.version} (${new Date(item.createTime).toLocaleDateString('zh-CN')})`,
    value: item.version
  }))
})

// 计算属性：将发布计划数据转换为前端显示格式
const deploymentPlansForDisplay = computed(() => {
  if (!currentServiceDeploymentPlans.value) {
    return []
  }
  
  return currentServiceDeploymentPlans.value.items.map(plan => {
    // 状态映射
    const statusMap = {
      'InDeployment': '部署中',
      'completed': '已完成',
      'rollbacked': '已回滚'
    }
    
    // 时间格式化
    const formatTime = (timeStr?: string) => {
      if (!timeStr) return '已开始'
      return new Date(timeStr).toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      })
    }
    
    return {
      id: plan.id,
      version: plan.version,
      status: statusMap[plan.status] || plan.status,
      time: plan.scheduleTime ? formatTime(plan.scheduleTime) : 
            plan.finishTime ? formatTime(plan.finishTime) : '已开始',
      originalStatus: plan.status
    }
  })
})

// 方法
const goToChangelog = () => {
  router.push('/changelog')
}

const getNodeStatus = (node: any) => {
  // 直接使用后端返回的health状态
  const healthMap: Record<string, string> = {
    'Normal': 'healthy',
    'Warning': 'canary',
    'Error': 'abnormal'
  }
  return healthMap[node.health] || 'healthy'
}

const getNodeStatusColor = (node: any) => {
  const status = getNodeStatus(node)
  const statusMap: Record<string, string> = {
    healthy: "#10b981",    // 绿色
    abnormal: "#f43f5e",   // 红色
    canary: "#f59e0b"      // 黄色
  }
  return statusMap[status] || "#6b7280"
}

const getNodeStatusStroke = (node: any) => {
  const status = getNodeStatus(node)
  const statusMap: Record<string, string> = {
    healthy: "#10b981",
    abnormal: "#f43f5e",
    canary: "#f59e0b"
  }
  return statusMap[status] || "#6b7280"
}

const getNodeStatusFill = (node: any) => {
  const status = getNodeStatus(node)
  const statusMap: Record<string, string> = {
    healthy: "#10b981",
    abnormal: "#f43f5e", 
    canary: "#f59e0b"
  }
  return statusMap[status] || "#6b7280"
}

const hasRollingVersion = (node: any) => {
  // 根据deployState判断是否显示灰度发布指示器
  return node.deployState === 'InDeploying'
}

const getNodePosition = (nodeId: string) => {
  const node = nodes.value.find(n => n.id === nodeId)
  return node ? { x: node.x, y: node.y } : { x: 0, y: 0 }
}

const handleNodeClick = async (node: any) => {
  selectedNode.value = { ...node, status: getNodeStatus(node) }
  dialogVisible.value = true
  
  // 并行加载服务详情、指标数据、可发布版本和发布计划
  const [serviceDetail, metricsData, availableVersionsData, deploymentPlansData] = await Promise.all([
    loadServiceDetail(node.name),
    loadServiceMetrics(node.name),
    loadServiceAvailableVersions(node.name),
    loadServiceDeploymentPlans(node.name)
  ])
  
  if (serviceDetail) {
    // 更新节点的版本信息
    selectedNode.value.versions = serviceDetail.versions
  }
  
  if (metricsData) {
    // 存储指标数据
    currentServiceMetrics.value = metricsData
  }
  
  if (availableVersionsData) {
    // 存储可发布版本数据
    currentServiceAvailableVersions.value = availableVersionsData
    // 重置选中的版本为第一个可用版本
    if (availableVersionsData.items.length > 0) {
      selectedVersion.value = availableVersionsData.items[0].version
    }
  }
  
  if (deploymentPlansData) {
    // 存储发布计划数据
    currentServiceDeploymentPlans.value = deploymentPlansData
  }
  
  nextTick(() => {
    initPieChart()
  })
}

const getStatusType = (status: string) => {
  switch (status) {
    case 'healthy': return 'success'
    case 'canary': return 'warning'
    case 'abnormal': return 'danger'
    default: return 'info'
  }
}

const getNodeStatusText = (status: string) => {
  switch (status) {
    case 'healthy': return '服务正常'
    case 'canary': return '有异常，AI正在观察和分析'
    case 'abnormal': return '服务有异常'
    default: return '未知状态'
  }
}

// 数据转换函数：将后端返回的指标数据转换为前端表格需要的格式
const transformMetricsToTableData = (versions: any[], metricsResponse: ServiceMetricsResponse | null) => {
  if (!metricsResponse) {
    // 如果没有指标数据，使用随机数据作为fallback
    return versions.map(v => ({
      version: v.label,
      latency: Math.floor(Math.random() * 40 + 8),
      traffic: Math.floor(Math.random() * 800 + 400),
      errors: (Math.random() * 4.7 + 0.3).toFixed(1),
      saturation: Math.floor(Math.random() * 57 + 35),
      status: v.anomalous ? '异常' : (v.observing ? '观察中' : '正常')
    }))
  }

  // 使用真实的指标数据
  return versions.map(v => {
    // 找到对应版本的指标数据
    const versionMetrics = metricsResponse.items.find(item => item.version === v.label)
    
    if (versionMetrics) {
      // 从指标数据中提取对应的值
      const latency = versionMetrics.metrics.find(m => m.name === 'latency')?.value || 0
      const traffic = versionMetrics.metrics.find(m => m.name === 'traffic')?.value || 0
      const errorRatio = versionMetrics.metrics.find(m => m.name === 'errorRatio')?.value || 0
      const saturation = versionMetrics.metrics.find(m => m.name === 'saturation')?.value || 0
      
      return {
        version: v.label,
        latency: latency,
        traffic: traffic,
        errors: errorRatio.toFixed(1),
        saturation: saturation,
        status: v.anomalous ? '异常' : (v.observing ? '观察中' : '正常')
      }
    } else {
      // 如果找不到对应版本的指标，使用summary数据
      const latency = metricsResponse.summary.metrics.find(m => m.name === 'latency')?.value || 0
      const traffic = metricsResponse.summary.metrics.find(m => m.name === 'traffic')?.value || 0
      const errorRatio = metricsResponse.summary.metrics.find(m => m.name === 'errorRatio')?.value || 0
      const saturation = metricsResponse.summary.metrics.find(m => m.name === 'saturation')?.value || 0
      
      return {
        version: v.label,
        latency: latency,
        traffic: traffic,
        errors: errorRatio.toFixed(1),
        saturation: saturation,
        status: v.anomalous ? '异常' : (v.observing ? '观察中' : '正常')
      }
    }
  })
}

const getVersionTableData = (versions: any[]) => {
  return transformMetricsToTableData(versions, currentServiceMetrics.value)
}

const handleCloseDialog = () => {
  selectedNode.value = null
  dialogVisible.value = false
  // 清理当前服务的数据
  currentServiceMetrics.value = null
  currentServiceAvailableVersions.value = null
  currentServiceDeploymentPlans.value = null
}

const createRelease = async () => {
  try {
    ElMessage.success('发布计划创建成功')
  } catch (error) {
    ElMessage.error('创建发布计划失败')
  }
}

const editRelease = (release: any) => {
  ElMessage.info('编辑发布功能待实现')
}

const confirmCancel = (release: any) => {
  ElMessage.info('取消发布功能待实现')
}

// 初始化饼图
let pieChart: echarts.ECharts | null = null

const initPieChart = () => {
  if (pieChartRef.value && selectedNode.value) {
    pieChart = echarts.init(pieChartRef.value)
    
    const option = {
      tooltip: {
        trigger: 'item',
        formatter: (params: any) => {
          const data = params.data
          let html = `<div style="padding: 10px;">
            <div style="font-weight: bold; margin-bottom: 8px;">灰度详情 · ${data.name}</div>
            <div style="display: flex; justify-content: space-between;">
              <span>占比</span>
              <span style="color: #666;">${data.value}%</span>
            </div>`
          
          // 显示开始时间和预估完成时间
          if (data.startTime && data.estimatedCompletionTime) {
            const startTime = new Date(data.startTime).toLocaleString('zh-CN', {
              year: 'numeric',
              month: '2-digit',
              day: '2-digit',
              hour: '2-digit',
              minute: '2-digit'
            })
            const estimatedTime = new Date(data.estimatedCompletionTime).toLocaleString('zh-CN', {
              year: 'numeric',
              month: '2-digit',
              day: '2-digit',
              hour: '2-digit',
              minute: '2-digit'
            })
            
            html += `<div style="margin-top: 6px; padding-top: 6px; border-top: 1px solid #eee;">
              <div style="margin-bottom: 4px;">
                <span style="color: #666;">开始时间:</span>
                <span style="margin-left: 8px; font-weight: 500;">${startTime}</span>
              </div>
              <div>
                <span style="color: #666;">预估完成:</span>
                <span style="margin-left: 8px; font-weight: 500;">${estimatedTime}</span>
              </div>
            </div>`
          }
          
          if (data.rolling) {
            html += `<div style="margin-top: 4px; color: #666;">
              发布持续时间 <b>${data.elapsedMin}</b> 分钟，预计剩余时间 <b>${data.remainingMin}</b> 分钟
            </div>`
            html += `<div style="margin-top: 4px;">ETA <span style="color: #666;">${data.eta}</span></div>`
          }
          
          // 根据版本的具体状态显示
          const versionStatus = data.anomalous ? 'abnormal' : (data.observing ? 'canary' : 'healthy')
          const statusText = data.anomalous ? '异常' : (data.observing ? '有异常点，AI正在观察和分析' : '正常')
          const statusColor = data.anomalous ? '#f43f5e' : (data.observing ? '#f59e0b' : '#10b981')
          html += `<div style="margin-top: 6px; padding-top: 6px; border-top: 1px solid #eee; color: ${statusColor}; font-size: 12px;">${statusText}</div>`
          html += '</div>'
          return html
        }
      },
      series: [{
        type: 'pie',
        radius: ['40%', '70%'],
        center: ['50%', '50%'],
        label: {
          show: true,
          position: 'outside',
          formatter: '{b}: {d}%',
          fontSize: 12,
          fontWeight: 'bold'
        },
        labelLine: {
          show: true,
          length: 15,
          length2: 10
        },
        data: selectedNode.value.versions.map((v: any, index: number) => {
          // 根据每个版本的具体状态确定颜色
          const versionStatus = v.anomalous ? 'abnormal' : (v.observing ? 'canary' : 'healthy')
          const statusMap: Record<string, string> = {
            healthy: "#10b981",    // 绿色
            abnormal: "#f43f5e",   // 红色
            canary: "#f59e0b"      // 黄色
          }
          
          return {
            name: v.label,
            value: v.value,
            ...v,
            itemStyle: {
              color: statusMap[versionStatus] || "#6b7280"
            }
          }
        }),
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }]
    }
    
    pieChart.setOption(option)
  }
}

// 监听 selectedNode 变化，重新初始化饼图
watch(() => selectedNode.value, () => {
  nextTick(() => {
    if (selectedNode.value) {
      initPieChart()
    }
  })
})

// 加载发布计划数据
const loadScheduledReleases = async () => {
  try {
    const response = await mockApi.getScheduledReleases()
    scheduledReleases.value = response
  } catch (err) {
    console.error('加载发布计划失败:', err)
  }
}
</script>

<style scoped>
.home-container {
  height: 100vh;
  width: 100vw;
  background: linear-gradient(to bottom, #f8fafc, #ffffff);
  padding: 24px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.title {
  font-size: 24px;
  font-weight: 600;
}

.subtitle {
  text-align: center;
  font-size: 18px;
  font-weight: 500;
  margin-bottom: 12px;
}

.topology-card {
  margin-bottom: 24px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.topology-container {
  position: relative;
  width: 100%;
  height: calc(100vh - 200px);
  min-height: 560px;
}

.loading-container, .error-container {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: #6b7280;
  font-size: 14px;
  z-index: 10;
}

.error-container {
  color: #f43f5e;
}

.error-container .el-button {
  margin-top: 8px;
}

.edges-svg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.service-node {
  position: absolute;
  transform: translate(-50%, -50%);
  cursor: pointer;
}

.node-circle {
  position: relative;
  width: 64px;
  height: 64px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border: 2px solid rgba(255, 255, 255, 0.2);
}

.node-name {
  color: white;
  font-size: 14px;
  font-weight: 500;
  user-select: none;
}

.rolling-indicator {
  position: absolute;
  top: -4px;
  right: -4px;
  width: 24px;
  height: 24px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 50%;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.rolling-indicator svg {
  width: 100%;
  height: 100%;
}

.legend {
  position: absolute;
  bottom: 12px;
  right: 12px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(8px);
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  padding: 12px;
  font-size: 12px;
  color: #475569;
}

.legend-title {
  font-weight: 500;
  color: #1e293b;
  margin-bottom: 8px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.legend-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.legend-divider {
  height: 1px;
  background: #e2e8f0;
  margin: 8px 0;
}

.rolling-example {
  width: 16px;
  height: 16px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 50%;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.rolling-example svg {
  width: 100%;
  height: 100%;
}

.service-detail {
  padding: 16px 0;
  width: 100%;
  height: 100%;
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.detail-desc {
  color: #6b7280;
  margin-bottom: 24px;
  font-size: 14px;
}

.detail-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
  width: 100%;
  height: 100%;
}

.pie-chart-container {
  height: 320px;
  width: 100%;
}

.pie-chart {
  width: 100%;
  height: 100%;
}

.version-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
}

.version-title {
  font-size: 14px;
  color: #6b7280;
}

.version-table {
  width: 100%;
}

.release-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  width: 100%;
}

.release-plans {
  max-height: 256px;
  overflow-y: auto;
}

.plans-header {
  font-weight: 500;
  font-size: 14px;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #e5e7eb;
}

.plans-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.plan-item {
  padding: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  transition: background-color 0.2s;
}

.plan-item:hover {
  background-color: #f9fafb;
}

.plan-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.plan-version {
  font-weight: 500;
  font-size: 14px;
}

.plan-actions {
  display: flex;
  gap: 4px;
}

.plan-details {
  font-size: 12px;
  color: #6b7280;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.no-plans {
  text-align: center;
  color: #6b7280;
  font-size: 14px;
  padding: 16px;
}

@media (max-width: 768px) {
  .detail-content {
    grid-template-columns: 1fr;
  }
  
  .release-controls {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>