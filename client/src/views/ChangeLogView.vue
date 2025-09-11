<template>
  <div class="changelog-container">
    <!-- 顶部导航 -->
    <div class="header">
      <div class="nav-section">
        <el-button @click="$router.push('/')" class="back-btn">
          <el-icon><ArrowLeft /></el-icon>
          返回首页
        </el-button>
        <div class="divider"></div>
        <div class="title">Zero Ops</div>
      </div>
    </div>

    <div class="page-title">系统状态变更记录</div>

    <!-- 标签页 -->
    <el-tabs v-model="activeTab" class="tabs">
      <el-tab-pane label="服务变更记录" name="service">
        <!-- 搜索框 -->
        <div class="search-section">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索服务/版本/状态…"
            clearable
            @input="handleSearch"
          />
        </div>

        <!-- 变更记录列表 -->
        <div class="change-list">
          <ChangeCard
            v-for="item in filteredChangeItems"
            :key="item.id"
            :item="item"
          />
          <div v-if="filteredChangeItems.length === 0" class="no-results">
            无匹配记录
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="告警变更记录" name="alarm">
        <div class="change-list">
          <AlarmChangeCard
            v-for="item in alarmChangeItems"
            :key="item.id"
            :item="item"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="xxx 变更记录" name="other1">
        <div class="placeholder">占位</div>
      </el-tab-pane>

      <el-tab-pane label="xxx 变更记录" name="other2">
        <div class="placeholder">占位</div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAppStore, type ChangeItem, type AlarmChangeItem } from '@/stores/app'
import ChangeCard from '@/components/ChangeCard.vue'
import AlarmChangeCard from '@/components/AlarmChangeCard.vue'

const appStore = useAppStore()

const activeTab = ref('service')
const searchKeyword = ref('')

// 模拟数据
const changeItems = ref<ChangeItem[]>([
  {
    id: 'chg-1',
    service: 'Stg',
    version: 'v1.0.3',
    state: '发布中',
    progress: 50,
    ok: true,
    batches: [
      { name: '第一批', status: '正常', start: '2025/9/4 12:00:00', end: '2025/9/4 12:10:00' },
      { name: '第二批', status: '正常', start: '2025/9/4 12:20:00', end: '2025/9/4 12:35:00' },
      {
        name: '第三批',
        status: '异常',
        start: '2025/9/4 12:50:00',
        end: '2025/9/4 13:10:00',
        anomaly: 'Stg服务指标异常。',
        moduleRecords: [
          {
            id: 'event-1',
            module: '发布系统',
            action: '发布到节点',
            timestamp: '2025/9/4 12:50:00',
            status: '成功',
            details: '发布到 bj1-node-002, qn1-node-002 节点'
          },
          {
            id: 'event-2',
            module: '发布系统',
            action: '部署新版本',
            timestamp: '2025/9/4 12:50:15',
            status: '成功',
            eventData: {
              deployment: {
                service: "Stg服务v1.0.3",
                environment: "生产环境",
                healthCheck: "通过",
                loadBalancer: "已更新路由规则",
                trafficStatus: "新版本开始接收流量"
              },
              deploymentResult: {
                status: "成功",
                process: "顺利",
                readiness: "所有健康检查通过"
              }
            }
          },
          {
            id: 'event-3',
            module: '指标分析',
            action: '指标检测',
            timestamp: '2025/9/4 12:50:20',
            status: '成功',
            details: '指标检测中...'
          },
          {
            id: 'event-4',
            module: '指标分析',
            action: '指标检测',
            timestamp: '2025/9/4 12:50:25',
            status: '成功',
            details: '指标检测无异常'
          },
          {
            id: 'event-5',
            module: '监控告警系统',
            action: '灰度发布监控',
            timestamp: '2025/9/4 12:55:20',
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
              },
              analysis: {
                rootCause: "新版本中的数据库查询优化存在问题",
                impact: "导致某些复杂查询性能下降"
              }
            }
          },
          {
            id: 'event-6',
            module: '指标分析',
            action: '指标检测异常',
            timestamp: '2025/9/4 12:55:35',
            status: '告警',
            details: '指标检测出现异常:响应时长高,错误率达 2.3%'
          },
          {
            id: 'event-7',
            module: '体检中心',
            action: '请求问题分级',
            timestamp: '2025/9/4 12:56:00',
            status: '告警',
            eventData: {
              problemClassification: {
                requestStatus: "正在请求问题分级系统",
                aiAnalysis: "AI分析异常指标后",
                classification: "WARNING级别"
              },
              assessment: {
                errorRate: "较高",
                serviceResponse: "仍可正常响应",
                coreFunctionImpact: "暂未影响核心功能"
              },
              recommendation: {
                action: "继续观察并准备回滚预案"
              }
            }
          },
          {
            id: 'event-8',
            module: '下钻分析',
            action: '问题下钻',
            timestamp: '2025/9/4 12:56:15',
            status: '告警',
            details: '正在请求问题下钻模块'
          },
          {
            id: 'event-9',
            module: '下钻分析',
            action: '问题下钻分析',
            timestamp: '2025/9/4 12:56:30',
            status: '告警',
            eventData: {
              analysisResult: {
                problemScope: "用户查询接口",
                rootCause: "新版本中的数据库索引优化导致某些复杂查询性能下降",
                impactDetails: [
                  "用户列表查询延迟增加300%",
                  "搜索功能错误率上升",
                  "数据库连接池使用率接近上限"
                ],
                recommendation: "建议立即回滚到稳定版本"
              },
              metrics: {
                queryLatencyIncrease: "300%",
                errorRateRise: "2.3%",
                connectionPoolUsage: "95%"
              },
              affectedServices: ["用户服务", "搜索服务", "数据库服务"]
            }
          },
          {
            id: 'event-10',
            module: '监控告警系统',
            action: '等待告警确认',
            timestamp: '2025/9/4 13:00:00',
            status: '告警',
            eventData: {
              alertConfirmation: {
                waitingTime: "5分钟",
                monitoringStatus: "持续监控中",
                alertSystem: "外部告警系统",
                confirmationProcess: "收集和分析指标数据"
              },
              currentMetrics: {
                responseTime: "450ms",
                errorRate: "2.3%",
                serviceStatus: "异常但可响应"
              },
              nextActions: [
                "告警系统确认异常",
                "触发自动回滚流程",
                "验证服务健康状态"
              ]
            }
          },
          {
            id: 'event-11',
            module: '发布系统',
            action: '自动回滚',
            timestamp: '2025/9/4 13:05:45',
            status: '回滚',
            eventData: {
              rollbackTrigger: {
                trigger: "AI检测到异常后自动触发回滚流程",
                aiJudgment: "基于延迟和错误率指标，AI判断当前版本存在严重问题"
              },
              rollbackStrategy: [
                "停止新版本流量",
                "恢复旧版本配置",
                "验证服务健康状态"
              ],
              rollbackResult: {
                status: "回滚完成",
                metrics: "所有指标恢复正常"
              }
            }
          }
        ]
      }
    ]
  },
  {
    id: 'chg-2',
    service: 'Stg',
    version: 'v1.0.2',
    state: '灰度中',
    progress: 30,
    ok: true,
    batches: [
      { name: '第一批', status: '正常', start: '2025/9/3 14:00:00', end: '2025/9/3 14:15:00' },
      { name: '第二批', status: '正常', start: '2025/9/3 14:15:00', end: '2025/9/3 14:30:00' },
      { name: '第三批', status: '进行中', start: '2025/9/3 14:30:00', end: '-' }
    ]
  },
  {
    id: 'chg-3',
    service: 'Stg',
    version: 'v1.0.1',
    state: '已完成',
    ok: true,
    batches: [
      { name: '第一批', status: '正常', start: '2025/9/2 10:00:00', end: '2025/9/2 10:20:00' },
      { name: '第二批', status: '正常', start: '2025/9/2 10:20:00', end: '2025/9/2 10:40:00' },
      { name: '第三批', status: '正常', start: '2025/9/2 10:40:00', end: '2025/9/2 11:00:00' }
    ]
  }
])

const alarmChangeItems = ref<AlarmChangeItem[]>([
  {
    id: 'alarm-1',
    service: 'Stg服务',
    change: '延时告警阈值调整: 10ms -> 15ms',
    timestamp: '2025/9/4 12:00:00',
    details: '由于业务增长，系统负载增加，原有10ms的延时阈值过于严格，导致频繁告警。经过AI分析历史数据，建议将阈值调整为15ms，既能及时发现性能问题，又避免误报。'
  },
  {
    id: 'alarm-2',
    service: 'Stg服务',
    change: '饱和度告警阈值调整: 50% -> 45%',
    timestamp: '2025/9/3 15:00:00',
    details: '监控发现系统在50%饱和度时已出现性能下降，提前预警有助于避免系统过载。调整后可以更早发现资源瓶颈，确保服务稳定性。'
  },
  {
    id: 'alarm-3',
    service: 'Mongo服务',
    change: '延时告警阈值调整: 10ms -> 5ms',
    timestamp: '2025/9/3 10:00:00',
    details: 'MongoDB服务经过优化后性能显著提升，原有10ms阈值已不适用。调整为5ms可以更精确地监控数据库性能，及时发现潜在问题。'
  },
  {
    id: 'alarm-4',
    service: 'Meta服务',
    change: '错误告警阈值调整: 10 -> 5',
    timestamp: '2025/9/1 15:00:00',
    details: 'Meta服务作为核心服务，对错误率要求更加严格。将错误告警阈值从10降低到5，可以更敏感地发现服务异常，确保数据一致性。'
  }
])

// 计算属性
const filteredChangeItems = computed(() => {
  if (!searchKeyword.value) {
    return changeItems.value
  }
  
  const keyword = searchKeyword.value.toLowerCase()
  return changeItems.value.filter(item => 
    (item.service + item.version + item.state).toLowerCase().includes(keyword)
  )
})

// 方法
const handleSearch = () => {
  // 搜索逻辑已在计算属性中处理
}
</script>

<style scoped>
.changelog-container {
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

.nav-section {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 4px;
}

.divider {
  width: 1px;
  height: 20px;
  background: #e2e8f0;
}

.title {
  font-size: 24px;
  font-weight: 600;
}

.page-title {
  font-size: 32px;
  font-weight: 600;
  margin-bottom: 16px;
}

.tabs {
  width: 100%;
}

.search-section {
  margin: 16px 0;
}

.change-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.no-results {
  text-align: center;
  color: #6b7280;
  font-size: 14px;
  padding: 32px;
}

.placeholder {
  text-align: center;
  color: #6b7280;
  font-size: 14px;
  padding: 32px;
}
</style>
