<template>
  <div class="alerts-page">
    <div class="content-wrapper">
      <!-- 顶部导航 -->
      <div class="header">
        <div class="nav-section">
          <el-button 
            @click="$router.push('/')"
            class="back-button"
          >
            <el-icon><ArrowLeft /></el-icon>
            返回首页
          </el-button>
          <div class="divider"></div>
          <div class="title">Zero Ops</div>
        </div>
      </div>

      <!-- 页面标题 -->
      <div class="page-title">告警记录</div>

      <!-- 状态过滤 -->
      <div class="filter-section">
        <div class="filter-buttons">
          <button 
            v-for="filter in filters" 
            :key="filter.key"
            :class="['filter-btn', { active: filterState === filter.key }]"
            @click="setFilterState(filter.key as 'all' | 'open' | 'closed')"
          >
            <span>{{ filter.label }}</span>
            <span class="count-badge">{{ filter.count }}</span>
          </button>
        </div>
      </div>

      <!-- 告警列表 -->
      <div class="alerts-list">
        <div 
          v-for="alert in filteredAlerts" 
          :key="alert.id"
          class="alert-card"
        >
          <div class="alert-content">
            <!-- 状态图标 -->
            <div class="status-icon">
              <el-icon v-if="alert.state === 'Open'" class="open-icon">
                <Warning />
              </el-icon>
              <el-icon v-else class="closed-icon">
                <Check />
              </el-icon>
            </div>

            <!-- 告警信息 -->
            <div class="alert-info">
              <div class="alert-header">
                <div class="alert-title">{{ alert.title }}</div>
                <div class="alert-badges">
                  <el-tag :type="getLevelType(alert.level)" size="small">
                    {{ alert.level }}
                  </el-tag>
                  <el-tag :type="getStateType(alert.alertState)" size="small">
                    {{ getStateText(alert.alertState) }}
                  </el-tag>
                </div>
              </div>

              <!-- 标签 -->
              <div class="alert-labels">
                <el-tag 
                  v-for="label in alert.labels" 
                  :key="label.key"
                  size="small" 
                  type="info" 
                  class="label-tag"
                >
                  {{ label.key }}={{ label.value }}
                </el-tag>
              </div>

              <!-- 时间信息 -->
              <div class="alert-time">
                {{ formatRelativeTime(alert.alertSince) }}
              </div>
            </div>

            <!-- 操作按钮 -->
            <div class="alert-actions">
              <el-button 
                v-if="canShowAnalysis(alert.alertState)"
                type="primary" 
                size="small" 
                @click="showAIAnalysis(alert)"
              >
                分析
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- AI分析对话框 -->
    <el-dialog
      v-model="showAnalysisDialog"
      title="AI分析处理记录"
      width="80%"
      :close-on-click-modal="false"
      class="analysis-dialog"
    >
      <div v-if="selectedAlert" class="analysis-content">
        <!-- 告警基本信息 -->
        <div class="alert-basic-info">
          <div class="info-badges">
            <el-tag :type="getLevelType(selectedAlert.level)">
              {{ selectedAlert.level }}
            </el-tag>
            <el-tag :type="getStateType(selectedAlert.alertState)">
              {{ getStateText(selectedAlert.alertState) }}
            </el-tag>
          </div>
          <div class="alert-time-info">
            告警时间: {{ formatRelativeTime(selectedAlert.alertSince) }}
          </div>
        </div>

        <!-- AI分析结果 -->
        <div v-if="selectedAlert.comments" class="ai-analysis-results">
          <div 
            v-for="(comment, index) in selectedAlert.comments" 
            :key="index"
            class="analysis-result"
          >
            <div class="result-header">
              <el-icon class="brain-icon"><InfoFilled /></el-icon>
              <span class="result-title">AI分析结果</span>
              <span class="result-time">{{ formatRelativeTime(comment.createAt) }}</span>
            </div>
            <div class="result-content">
              <pre class="markdown-content">{{ comment.content }}</pre>
            </div>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="analysis-actions">
          <el-button 
            v-if="selectedAlert.alertState === 'InProcessing'"
            type="danger" 
            size="small"
            @click="executeRollback"
          >
            执行回滚
          </el-button>
          <el-button 
            v-if="selectedAlert.alertState === 'InProcessing'"
            type="default" 
            size="small"
            @click="markAsRestored"
          >
            标记恢复正常
          </el-button>
          <div 
            v-if="selectedAlert.alertState === 'Restored'"
            class="restored-status"
          >
            <el-icon class="check-icon"><Check /></el-icon>
            已恢复正常
          </div>
          <div 
            v-if="selectedAlert.alertState === 'AutoRestored'"
            class="restored-status"
          >
            <el-icon class="check-icon"><Check /></el-icon>
            系统自动恢复
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { 
  ArrowLeft, 
  Warning, 
  Check, 
  InfoFilled 
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const router = useRouter()

// 状态管理
const filterState = ref<'all' | 'open' | 'closed'>('all')
const showAnalysisDialog = ref(false)
const selectedAlert = ref<any>(null)
const alerts = ref<any[]>([])

// 计算属性
const filters = computed(() => [
  { 
    key: 'all', 
    label: 'All', 
    count: alerts.value.length 
  },
  { 
    key: 'open', 
    label: 'Open', 
    count: alerts.value.filter(alert => alert.state === 'Open').length 
  },
  { 
    key: 'closed', 
    label: 'Closed', 
    count: alerts.value.filter(alert => alert.state === 'Closed').length 
  }
])

const filteredAlerts = computed(() => {
  if (filterState.value === 'all') return alerts.value
  if (filterState.value === 'open') return alerts.value.filter(alert => alert.state === 'Open')
  if (filterState.value === 'closed') return alerts.value.filter(alert => alert.state === 'Closed')
  return alerts.value
})

// 方法
const setFilterState = (state: 'all' | 'open' | 'closed') => {
  filterState.value = state
}

const formatRelativeTime = (timestamp: string) => {
  const now = new Date()
  const alertTime = new Date(timestamp)
  const diffMs = now.getTime() - alertTime.getTime()
  const diffMinutes = Math.floor(diffMs / (1000 * 60))
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))
  const diffMonths = Math.floor(diffDays / 30)

  if (diffMonths >= 1) {
    return alertTime.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  } else if (diffDays >= 1) {
    return `${diffDays}天前`
  } else if (diffHours >= 1) {
    return `${diffHours}小时前`
  } else if (diffMinutes >= 1) {
    return `${diffMinutes}分钟前`
  } else {
    return '刚刚'
  }
}

const getLevelType = (level: string) => {
  switch (level) {
    case 'P0':
    case 'P1':
      return 'danger'
    case 'P2':
      return 'warning'
    case 'Warning':
      return 'warning'
    default:
      return 'info'
  }
}

const getStateType = (alertState: string) => {
  switch (alertState) {
    case 'Restored':
    case 'AutoRestored':
      return 'success'
    case 'InProcessing':
      return 'danger'
    default:
      return 'info'
  }
}

const getStateText = (alertState: string) => {
  switch (alertState) {
    case 'Restored':
      return '已恢复'
    case 'AutoRestored':
      return '自然恢复'
    case 'InProcessing':
      return '处理中'
    default:
      return alertState
  }
}

const canShowAnalysis = (alertState: string) => {
  return ['InProcessing', 'Restored', 'AutoRestored'].includes(alertState)
}

const showAIAnalysis = (alert: any) => {
  selectedAlert.value = alert
  showAnalysisDialog.value = true
}

const executeRollback = () => {
  ElMessage.success('执行回滚操作，跳转到系统变更记录页面')
  showAnalysisDialog.value = false
  router.push('/changelog')
}

const markAsRestored = () => {
  ElMessage.success('标记为恢复正常')
  showAnalysisDialog.value = false
}

// 模拟告警数据 - 完整版本，与prototyping2一致
const mockAlerts = [
  {
    id: 'alert-1',
    state: 'Closed',
    level: 'P1',
    alertState: 'Restored',
    title: 'yzh S3APIV2 s3apiv2.putobject 0_64K上传响应时间95值: 500.12ms > 450ms',
    labels: [
      { key: 'api', value: 's3apiv2.putobject' },
      { key: 'idc', value: 'yzh' },
      { key: 'org', value: 'kodo' },
      { key: 'prophet_service', value: 's3apiv2' },
      { key: 'prophet_type', value: 'app' }
    ],
    alertSince: '2025-09-01T19:14:12.382331146Z',
    comments: [
      {
        createAt: '2025-09-01T19:15:00Z',
        content: `## AI分析结果

**问题类型**: 发版本导致的问题
**根因分析**: 新版本中的数据库查询优化存在问题，导致某些复杂查询性能下降

**处理建议**: 
- 立即回滚到稳定版本
- 优化数据库索引配置
- 增加监控告警阈值

**执行状态**: 已执行回滚操作，等待指标恢复正常`
      }
    ],
    prometheusConfig: {
      name: 'apitime_s3apiv2_upload_2xx_p1',
      expr: 'histogram_quantile(0.95, sum by (le, idc, api, org) (rate(service_response_duration_ms_bucket{api=~"s3apiv2.putobject|s3apiv2.uploadpart|s3apiv2.completemultipartupload|s3apiv2.initiatemultipartupload|s3apiv2.postobject",code=~"2..",reqlength="0_65536",service="S3APIV2"}[5m]))) > 150 and sum by (idc, api, org) (rate(service_response_code{service="S3APIV2"}[5m])) > 450',
      for: '5m',
      labels: {
        env: 'hd',
        prophet_service: 's3apiv2',
        prophet_type: 'app',
        severity: 'P1'
      },
      annotations: {
        description: '{{ $labels.idc }} S3APIV2 {{ $labels.api }} 0_64K上传响应时间95值: {{ $value | printf `%.2f`}}ms > 450ms',
        val: '{{ $value | printf `%.2f` }}'
      }
    },
    aiAnalysis: {
      mainTask: "yzh S3APIV2 s3apiv2.putobject 0_64K上传响应时间95值: 500.12ms > 450ms",
      mainPlan: {
        thought: "检测到S3APIV2服务P95延迟超过450ms阈值，这是一个P1级别告警，需要立即进行多维度分析来定位问题根源。",
        subTasks: [
          "1. 异常指标下钻分析（sub task 1）",
          "2. 资源使用情况分析（sub task 2）", 
          "3. 问题根因诊断（sub task 3）",
          "4. 告警规则评估（sub task 4）"
        ]
      },
      subTasks: [
        {
          id: 'subtask-1',
          name: '异常指标下钻分析',
          timestamp: '2025-09-01T19:14:15.000Z',
          status: 'completed',
          type: 'subagent',
          plan: {
            thought: "需要从多个维度进行下钻分析，包括IDC、API类型、请求大小、时间段等维度，以精确定位延迟异常的具体原因。",
            subTasks: [
              "1. IDC维度延迟分析（sub task 1.1）",
              "2. API类型性能分析（sub task 1.2）",
              "3. 请求大小影响分析（sub task 1.3）"
            ]
          },
          subSubTasks: [
            {
              id: 'subtask-1-1',
              name: 'IDC维度延迟分析',
              timestamp: '2025-09-01T19:14:18.000Z',
              thought: "分析各IDC的延迟分布情况，识别是否存在特定IDC的延迟异常。",
              result: "分析结果：北京IDC延迟异常，P95延迟从120ms激增至500ms，其他IDC正常。异常时间窗口：19:10-19:15。"
            },
            {
              id: 'subtask-1-2', 
              name: 'API类型性能分析',
              timestamp: '2025-09-01T19:14:20.000Z',
              thought: "分析不同API接口的性能表现，确定哪些接口受到影响。",
              result: "分析结果：主要影响s3apiv2.putobject接口，其他接口性能正常。影响范围：0-64KB文件上传。"
            },
            {
              id: 'subtask-1-3',
              name: '请求大小影响分析', 
              timestamp: '2025-09-01T19:14:22.000Z',
              thought: "分析不同请求大小对延迟的影响，确定是否存在特定请求大小的性能问题。",
              result: "分析结果：0-64KB请求延迟异常，大文件请求正常。说明问题可能与小文件处理逻辑相关。"
            }
          ],
          result: "下钻分析完成：北京IDC的s3apiv2.putobject接口在0-64KB文件上传时出现延迟异常，需要进一步分析资源使用情况。"
        },
        {
          id: 'subtask-2',
          name: '资源使用情况分析',
          timestamp: '2025-09-01T19:14:25.000Z',
          status: 'completed',
          type: 'tool',
          thought: "需要分析服务器资源使用情况，包括CPU、内存、网络等指标，确定是否存在资源瓶颈。",
          result: "资源分析结果：CPU使用率上升至85%，内存使用率正常(65%)，网络带宽使用率正常(45%)。CPU资源紧张是主要问题。"
        },
        {
          id: 'subtask-3',
          name: '问题根因诊断',
          timestamp: '2025-09-01T19:14:35.000Z',
          status: 'completed',
          type: 'subagent',
          plan: {
            thought: "基于前面的分析结果，需要综合判断问题的根本原因，并制定解决方案。",
            subTasks: [
              "1. 根因分析（sub task 3.1）",
              "2. 影响评估（sub task 3.2）",
              "3. 解决方案制定（sub task 3.3）"
            ]
          },
          subSubTasks: [
            {
              id: 'subtask-3-1',
              name: '根因分析',
              timestamp: '2025-09-01T19:14:36.000Z',
              thought: "综合分析所有指标，确定问题的根本原因。",
              result: "根因分析：北京IDC服务器CPU资源不足，导致s3apiv2.putobject接口在处理0-64KB文件时响应变慢。"
            },
            {
              id: 'subtask-3-2',
              name: '影响评估',
              timestamp: '2025-09-01T19:14:38.000Z',
              thought: "评估问题的影响范围和严重程度。",
              result: "影响评估：P1级别，主要影响0-64KB文件上传功能，影响用户体验，需要立即处理。"
            },
            {
              id: 'subtask-3-3',
              name: '解决方案制定',
              timestamp: '2025-09-01T19:14:40.000Z',
              thought: "制定立即和长期的解决方案。",
              result: "解决方案：1) 立即扩容北京IDC服务器资源 2) 优化s3apiv2.putobject接口性能 3) 考虑流量调度到其他IDC。"
            }
          ],
          result: "问题诊断完成：根本原因是北京IDC服务器CPU资源不足，建议立即扩容并优化接口性能。"
        },
        {
          id: 'subtask-4',
          name: '告警规则评估',
          timestamp: '2025-09-01T19:14:45.000Z',
          status: 'completed',
          type: 'tool',
          thought: "评估当前告警规则的合理性，确定是否需要调整阈值或监控策略。",
          result: "告警规则评估：当前阈值450ms合理，能够及时发现性能问题。历史基线120ms，当前500ms明显异常。建议保持当前规则不变。"
        }
      ]
    }
  },
  {
    id: 'alert-2',
    state: 'Open',
    level: 'P2',
    alertState: 'InProcessing',
    title: 'bj1 S3APIV2 s3apiv2.getobject 错误率过高: 5.2% > 3%',
    labels: [
      { key: 'api', value: 's3apiv2.getobject' },
      { key: 'idc', value: 'bj1' },
      { key: 'org', value: 'kodo' },
      { key: 'prophet_service', value: 's3apiv2' },
      { key: 'prophet_type', value: 'app' }
    ],
    alertSince: '2025-09-01T18:45:30.123456789Z',
    comments: [
      {
        createAt: '2025-09-01T18:46:00Z',
        content: `## AI分析结果

**问题类型**: 非发版本导致的问题
**根因分析**: 数据库连接池配置不足，导致大量请求无法获取数据库连接

**处理建议**: 
- 增加数据库连接池大小
- 优化数据库连接管理
- 考虑读写分离缓解压力

**执行状态**: 正在处理中，等待指标恢复正常`
      }
    ],
    prometheusConfig: {
      name: 'high_error_rate_s3apiv2',
      expr: 'rate(service_response_code{service="S3APIV2",code=~"5.."}[5m]) / rate(service_response_code{service="S3APIV2"}[5m]) * 100 > 3',
      for: '3m',
      labels: {
        env: 'hd',
        prophet_service: 's3apiv2',
        prophet_type: 'app',
        severity: 'P2'
      },
      annotations: {
        description: 'S3APIV2服务错误率过高: {{ $value | printf `%.2f`}}% > 3%',
        val: '{{ $value | printf `%.2f` }}'
      }
    },
    aiAnalysis: {
      mainTask: "bj1 S3APIV2 s3apiv2.getobject 错误率过高: 5.2% > 3%",
      mainPlan: {
        thought: "检测到S3APIV2服务错误率异常，5xx错误率达到5.2%，超过3%阈值。这是一个P2级别告警，需要及时分析错误原因并采取治愈措施。",
        subTasks: [
          "1. 错误码下钻分析（sub task 1）",
          "2. 资源使用情况分析（sub task 2）",
          "3. 问题根因诊断（sub task 3）",
          "4. 告警规则评估（sub task 4）"
        ]
      },
      subTasks: [
        {
          id: 'subtask-1',
          name: '错误码下钻分析',
          timestamp: '2025-09-01T18:45:35.000Z',
          status: 'completed',
          type: 'subagent',
          plan: {
            thought: "需要从错误码类型、API接口、IDC维度、时间段等维度进行下钻分析，快速定位具体的错误类型和影响范围。",
            subTasks: [
              "1. 错误码类型分析（sub task 1.1）",
              "2. API接口错误分布分析（sub task 1.2）",
              "3. IDC维度错误率分析（sub task 1.3）"
            ]
          },
          subSubTasks: [
            {
              id: 'subtask-1-1',
              name: '错误码类型分析',
              timestamp: '2025-09-01T18:45:40.000Z',
              thought: "分析不同错误码的分布情况，确定主要的错误类型。",
              result: "分析结果：主要错误码为503 Service Unavailable，占比60%。其他错误码：500(20%)、502(15%)、504(5%)。"
            },
            {
              id: 'subtask-1-2',
              name: 'API接口错误分布分析',
              timestamp: '2025-09-01T18:45:42.000Z',
              thought: "分析不同API接口的错误分布，确定哪些接口受影响最严重。",
              result: "分析结果：s3apiv2.getobject接口错误率最高，占比80%。其他接口错误率正常。"
            },
            {
              id: 'subtask-1-3',
              name: 'IDC维度错误率分析',
              timestamp: '2025-09-01T18:45:44.000Z',
              thought: "分析各IDC的错误率分布，确定是否存在特定IDC的问题。",
              result: "分析结果：北京IDC错误率异常，其他IDC正常。错误时间窗口：18:40-18:50。"
            }
          ],
          result: "下钻分析完成：主要是503错误，集中在s3apiv2.getobject接口，北京IDC异常，需要进一步分析资源使用情况。"
        },
        {
          id: 'subtask-2',
          name: '资源使用情况分析',
          timestamp: '2025-09-01T18:45:50.000Z',
          status: 'completed',
          type: 'tool',
          thought: "需要分析服务器资源使用情况，特别是数据库连接池、CPU、内存等指标，确定是否存在资源瓶颈。",
          result: "资源分析结果：数据库连接池使用率100%，连接池耗尽。CPU使用率正常(45%)，内存使用率正常(60%)。数据库连接池是主要瓶颈。"
        },
        {
          id: 'subtask-3',
          name: '问题根因诊断',
          timestamp: '2025-09-01T18:46:00.000Z',
          status: 'completed',
          type: 'subagent',
          plan: {
            thought: "基于前面的分析结果，需要综合判断问题的根本原因，并制定解决方案。",
            subTasks: [
              "1. 根因分析（sub task 3.1）",
              "2. 影响评估（sub task 3.2）",
              "3. 解决方案制定（sub task 3.3）"
            ]
          },
          subSubTasks: [
            {
              id: 'subtask-3-1',
              name: '根因分析',
              timestamp: '2025-09-01T18:46:02.000Z',
              thought: "综合分析所有指标，确定问题的根本原因。",
              result: "根因分析：数据库连接池配置不足，导致大量请求无法获取数据库连接，返回503错误。"
            },
            {
              id: 'subtask-3-2',
              name: '影响评估',
              timestamp: '2025-09-01T18:46:04.000Z',
              thought: "评估问题的影响范围和严重程度。",
              result: "影响评估：P2级别，主要影响文件下载功能，影响部分用户，需要及时处理。"
            },
            {
              id: 'subtask-3-3',
              name: '解决方案制定',
              timestamp: '2025-09-01T18:46:06.000Z',
              thought: "制定立即和长期的解决方案。",
              result: "解决方案：1) 立即增加数据库连接池大小 2) 优化数据库连接管理 3) 考虑读写分离缓解压力。"
            }
          ],
          result: "问题诊断完成：根本原因是数据库连接池配置不足，建议立即增加连接池大小并优化连接管理。"
        },
        {
          id: 'subtask-4',
          name: '告警规则评估',
          timestamp: '2025-09-01T18:46:10.000Z',
          status: 'completed',
          type: 'tool',
          thought: "评估当前告警规则的合理性，确定是否需要调整阈值或监控策略。",
          result: "告警规则评估：当前阈值3%合理，能够及时发现服务异常。历史基线0.5%，当前5.2%明显异常。建议保持当前规则不变。"
        }
      ]
    }
  },
  {
    id: 'alert-3',
    state: 'Open',
    level: 'P2',
    alertState: 'InProcessing',
    title: 'sh1 S3APIV2 s3apiv2.listobjects CPU使用率过高: 85.6% > 80%',
    labels: [
      { key: 'api', value: 's3apiv2.listobjects' },
      { key: 'idc', value: 'sh1' },
      { key: 'org', value: 'kodo' },
      { key: 'prophet_service', value: 's3apiv2' },
      { key: 'prophet_type', value: 'app' }
    ],
    alertSince: '2025-09-01T20:15:45.789123456Z',
    comments: [
      {
        createAt: '2025-09-01T20:16:00Z',
        content: `## AI分析结果

**问题类型**: 非发版本导致的问题
**根因分析**: CPU使用率过高，可能是由于系统负载增加或资源不足

**处理建议**: 
- 检查系统资源使用情况
- 优化CPU密集型操作
- 考虑扩容或负载均衡

**执行状态**: 正在处理中，等待指标恢复正常`
      }
    ],
    prometheusConfig: {
      name: 'high_cpu_usage_s3apiv2',
      expr: 'rate(container_cpu_usage_seconds_total{service="S3APIV2"}[5m]) * 100 > 80',
      for: '2m',
      labels: {
        env: 'hd',
        prophet_service: 's3apiv2',
        prophet_type: 'app',
        severity: 'P2'
      },
      annotations: {
        description: 'S3APIV2服务CPU使用率过高: {{ $value | printf `%.2f`}}% > 80%',
        val: '{{ $value | printf `%.2f` }}'
      }
    }
  },
  {
    id: 'alert-4',
    state: 'Open',
    level: 'P1',
    alertState: 'InProcessing',
    title: 'gz1 S3APIV2 s3apiv2.getobject 数据库连接池耗尽: 100% > 95%',
    labels: [
      { key: 'api', value: 's3apiv2.getobject' },
      { key: 'idc', value: 'gz1' },
      { key: 'org', value: 'kodo' },
      { key: 'prophet_service', value: 's3apiv2' },
      { key: 'prophet_type', value: 'app' }
    ],
    alertSince: '2025-09-01T21:30:12.456789123Z',
    comments: [
      {
        createAt: '2025-09-01T21:31:00Z',
        content: `## AI分析结果

**问题类型**: 非发版本导致的问题
**根因分析**: 数据库连接池配置不足，导致连接池耗尽

**处理建议**: 
- 立即增加数据库连接池大小
- 优化数据库连接管理
- 检查是否有连接泄漏

**执行状态**: 正在处理中，等待指标恢复正常`
      }
    ],
    prometheusConfig: {
      name: 'database_connection_pool_exhausted',
      expr: 'database_connections_active / database_connections_max * 100 > 95',
      for: '1m',
      labels: {
        env: 'hd',
        prophet_service: 's3apiv2',
        prophet_type: 'app',
        severity: 'P1'
      },
      annotations: {
        description: '数据库连接池使用率过高: {{ $value | printf `%.2f`}}% > 95%',
        val: '{{ $value | printf `%.2f` }}'
      }
    }
  },
  {
    id: 'alert-5',
    state: 'Closed',
    level: 'P2',
    alertState: 'AutoRestored',
    title: 'cd1 S3APIV2 s3apiv2.uploadpart 内存使用率过高: 92.3% > 90%',
    labels: [
      { key: 'api', value: 's3apiv2.uploadpart' },
      { key: 'idc', value: 'cd1' },
      { key: 'org', value: 'kodo' },
      { key: 'prophet_service', value: 's3apiv2' },
      { key: 'prophet_type', value: 'app' }
    ],
    alertSince: '2025-09-01T22:45:30.123456789Z',
    comments: [
      {
        createAt: '2025-09-01T22:46:00Z',
        content: `## AI分析结果

**问题类型**: 非发版本导致的问题
**根因分析**: 内存使用率过高，可能是由于内存泄漏或资源不足

**处理建议**: 
- 检查内存使用情况
- 优化内存管理
- 考虑增加内存资源

**执行状态**: 系统自动恢复，指标已恢复正常`
      }
    ],
    prometheusConfig: {
      name: 'memory_usage_high_s3apiv2',
      expr: 'container_memory_usage_bytes / container_memory_limit_bytes * 100 > 90',
      for: '3m',
      labels: {
        env: 'hd',
        prophet_service: 's3apiv2',
        prophet_type: 'app',
        severity: 'P2'
      },
      annotations: {
        description: 'S3APIV2服务内存使用率过高: {{ $value | printf `%.2f`}}% > 90%',
        val: '{{ $value | printf `%.2f` }}'
      }
    },
    aiAnalysis: {
      mainTask: "cd1 S3APIV2 s3apiv2.uploadpart 内存使用率过高: 92.3% > 90%",
      mainPlan: {
        thought: "检测到S3APIV2服务内存使用率异常，达到92.3%，超过90%阈值。这是一个P2级别告警，需要分析内存使用情况并找出内存泄漏或资源不足的原因。",
        subTasks: [
          "1. 内存使用情况分析（sub task 1）",
          "2. 内存泄漏检测（sub task 2）",
          "3. 问题根因诊断（sub task 3）",
          "4. 告警规则评估（sub task 4）"
        ]
      },
      subTasks: [
        {
          id: 'subtask-1',
          name: '内存使用情况分析',
          timestamp: '2025-09-01T22:45:35.000Z',
          status: 'completed',
          type: 'tool',
          thought: "需要分析内存使用情况，包括堆内存、非堆内存、缓存等各个组件的使用情况。",
          result: "内存分析结果：堆内存使用率85%，非堆内存使用率7%，缓存使用率正常。主要问题集中在堆内存。"
        },
        {
          id: 'subtask-2',
          name: '内存泄漏检测',
          timestamp: '2025-09-01T22:45:45.000Z',
          status: 'failed',
          type: 'tool',
          thought: "需要调用内存泄漏检测工具来分析是否存在内存泄漏问题。",
          result: "工具调用失败：内存泄漏检测工具连接超时，无法获取内存快照数据。错误信息：Connection timeout after 30 seconds."
        },
        {
          id: 'subtask-3',
          name: '问题根因诊断',
          timestamp: '2025-09-01T22:46:00.000Z',
          status: 'failed',
          type: 'subagent',
          plan: {
            thought: "由于内存泄漏检测工具失败，无法获取完整的分析数据，需要基于现有信息进行诊断。",
            subTasks: [
              "1. 基于现有数据分析（sub task 3.1）",
              "2. 历史数据对比（sub task 3.2）"
            ]
          },
          subSubTasks: [
            {
              id: 'subtask-3-1',
              name: '基于现有数据分析',
              timestamp: '2025-09-01T22:46:05.000Z',
              thought: "基于堆内存使用率85%的数据进行分析。",
              result: "分析结果：堆内存使用率较高，可能存在内存泄漏或对象未及时释放的问题。"
            },
            {
              id: 'subtask-3-2',
              name: '历史数据对比',
              timestamp: '2025-09-01T22:46:10.000Z',
              thought: "对比历史内存使用数据，分析内存增长趋势。",
              result: "历史数据对比失败：历史数据查询服务不可用，无法获取对比数据。"
            }
          ],
          result: "问题诊断失败：由于工具调用失败和数据服务不可用，无法完成完整的根因分析。"
        },
        {
          id: 'subtask-4',
          name: '告警规则评估',
          timestamp: '2025-09-01T22:46:15.000Z',
          status: 'completed',
          type: 'tool',
          thought: "评估当前告警规则的合理性。",
          result: "告警规则评估：当前阈值90%合理，能够及时发现内存使用异常。建议保持当前规则不变。"
        }
      ]
    }
  }
]

// 生命周期
onMounted(() => {
  alerts.value = mockAlerts
})
</script>

<style scoped>
.alerts-page {
  width: 100vw;
  min-height: 100vh;
  background: linear-gradient(to bottom, #f8fafc, #ffffff);
  padding: 0;
  margin: 0;
  box-sizing: border-box;
  overflow-x: hidden;
}

.alerts-page .content-wrapper {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
  width: 100%;
  box-sizing: border-box;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.nav-section {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-button {
  display: flex;
  align-items: center;
  gap: 4px;
}

.divider {
  width: 1px;
  height: 20px;
  background-color: #e2e8f0;
}

.title {
  font-size: 20px;
  font-weight: 600;
  color: #1e293b;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 16px;
}

.filter-section {
  margin-bottom: 24px;
}

.filter-buttons {
  display: flex;
  gap: 16px;
}

.filter-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: #64748b;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-btn:hover {
  color: #1e293b;
}

.filter-btn.active {
  background-color: #f1f5f9;
  color: #1e293b;
  font-weight: 500;
}

.count-badge {
  background-color: #f1f5f9;
  color: #475569;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 12px;
}

.filter-btn.active .count-badge {
  background-color: #e2e8f0;
  color: #334155;
}

.alerts-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.alert-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border: 1px solid #e2e8f0;
}

.alert-content {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px;
}

.status-icon {
  flex-shrink: 0;
  margin-top: 4px;
}

.open-icon {
  color: #ef4444;
  font-size: 16px;
}

.closed-icon {
  color: #22c55e;
  font-size: 16px;
}

.alert-info {
  flex: 1;
}

.alert-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.alert-title {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
  line-height: 1.4;
}

.alert-badges {
  display: flex;
  gap: 8px;
}

.alert-labels {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-bottom: 8px;
}

.label-tag {
  font-size: 12px;
}

.alert-time {
  font-size: 12px;
  color: #64748b;
}

.alert-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.analysis-dialog {
  max-height: 80vh;
}

.analysis-content {
  max-height: 60vh;
  overflow-y: auto;
}

.alert-basic-info {
  background-color: #f8fafc;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.info-badges {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}

.alert-time-info {
  font-size: 14px;
  color: #64748b;
}

.ai-analysis-results {
  margin-bottom: 16px;
}

.analysis-result {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.brain-icon {
  color: #3b82f6;
  font-size: 16px;
}

.result-title {
  font-weight: 500;
  color: #1e293b;
}

.result-time {
  font-size: 12px;
  color: #64748b;
  margin-left: auto;
}

.result-content {
  background-color: #f8fafc;
  border-radius: 6px;
  padding: 12px;
}

.markdown-content {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  font-size: 14px;
  line-height: 1.6;
  color: #374151;
  white-space: pre-wrap;
  margin: 0;
}

.analysis-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
}

.restored-status {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #22c55e;
}

.check-icon {
  font-size: 16px;
}
</style>