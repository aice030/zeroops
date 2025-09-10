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
              <el-select v-model="selectedVersion" placeholder="选择目标版本" style="width: 200px">
                <el-option
                  v-for="option in versionOptions"
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
                  <div v-if="scheduledReleases.length > 0" class="plans-list">
                    <div
                      v-for="release in scheduledReleases"
                      :key="release.id"
                      class="plan-item"
                    >
                      <div class="plan-header">
                        <span class="plan-version">{{ release.version }}</span>
                        <div class="plan-actions">
                          <el-button size="small" @click="editRelease(release)">编辑</el-button>
                          <el-button size="small" type="danger" @click="confirmCancel(release)">取消</el-button>
                        </div>
                      </div>
                      <div class="plan-details">
                        <div>开始时间: {{ release.startTime }}</div>
                        <div>创建人: {{ release.creator }}</div>
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
import * as echarts from 'echarts'

const router = useRouter()

// 响应式数据
const dialogVisible = ref(false)
const selectedNode = ref<any>(null)
const selectedVersion = ref('v1.0.7')
const scheduledStart = ref('')
const scheduledReleases = ref<any[]>([])
const pieChartRef = ref<HTMLElement>()

// 模拟数据 - 服务拓扑
const nodes = ref([
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
const versionOptions = computed(() => [
  { label: 'v1.0.4', value: 'v1.0.4' },
  { label: 'v1.0.5', value: 'v1.0.5' },
  { label: 'v1.0.6', value: 'v1.0.6' },
  { label: 'v1.0.7', value: 'v1.0.7' },
])

// 方法
const goToChangelog = () => {
  router.push('/changelog')
}

const getNodeStatus = (node: any) => {
  return node.versions.some((v: any) => v.anomalous) ? 'abnormal' : 
         (node.versions.some((v: any) => v.observing) ? 'canary' : 'healthy')
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
  return node.versions.some((v: any) => v.rolling)
}

const getNodePosition = (nodeId: string) => {
  const node = nodes.value.find(n => n.id === nodeId)
  return node ? { x: node.x, y: node.y } : { x: 0, y: 0 }
}

const handleNodeClick = (node: any) => {
  selectedNode.value = { ...node, status: getNodeStatus(node) }
  dialogVisible.value = true
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

const getVersionTableData = (versions: any[]) => {
  return versions.map(v => ({
    version: v.label,
    latency: Math.floor(Math.random() * 40 + 8),
    traffic: Math.floor(Math.random() * 800 + 400),
    errors: (Math.random() * 4.7 + 0.3).toFixed(1),
    saturation: Math.floor(Math.random() * 57 + 35),
    status: v.anomalous ? '异常' : (v.observing ? '观察中' : '正常')
  }))
}

const handleCloseDialog = () => {
  selectedNode.value = null
  dialogVisible.value = false
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
          html += `<div style="margin-top: 4px; color: ${statusColor}; font-size: 12px;">${statusText}</div>`
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

onMounted(() => {
  // 初始化发布计划数据
  scheduledReleases.value = [
    { id: '1', version: 'v1.0.4', startTime: '2025/09/01 19:00', creator: '张三' },
    { id: '2', version: 'v1.0.5', startTime: '2025/09/02 19:00', creator: '李四' },
    { id: '3', version: 'v1.0.6', startTime: '2025/09/03 19:00', creator: '王五' },
    { id: '4', version: 'v1.0.7', startTime: '2025/09/04 19:00', creator: '赵六' },
  ]
})
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