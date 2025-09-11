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
import { ref, computed, onMounted } from 'vue'
import { useAppStore, type ChangeItem, type AlarmChangeItem } from '@/stores/app'
import { mockApi } from '@/mock/api'
import type { DeploymentChangelogResponse, DeploymentChangelogItem } from '@/mock/services'
import ChangeCard from '@/components/ChangeCard.vue'
import AlarmChangeCard from '@/components/AlarmChangeCard.vue'

const appStore = useAppStore()

const activeTab = ref('service')
const searchKeyword = ref('')
const loading = ref(false)
const error = ref<string | null>(null)

// 部署变更记录数据
const deploymentChangelog = ref<DeploymentChangelogResponse | null>(null)
const changeItems = ref<ChangeItem[]>([])

// 数据转换函数：将API返回的数据转换为前端需要的格式
const transformDeploymentChangelogToChangeItems = (changelogData: any[]): ChangeItem[] => {
  return changelogData.map((item, index) => {
    // 根据健康状态确定状态
    let state: '发布中' | '灰度中' | '已完成'
    let progress: number | undefined
    let ok: boolean
    
    if (item.endTime) {
      // 有结束时间，说明已完成
      state = '已完成'
      progress = 100
      ok = item.health === 'Normal'
    } else if (item.instances < item.totalInstances) {
      // 实例数小于总数，说明在灰度中
      state = '灰度中'
      progress = Math.round((item.instances / item.totalInstances) * 100)
      ok = item.health === 'Normal'
    } else {
      // 实例数等于总数，说明在发布中
      state = '发布中'
      progress = 50 // 默认进度
      ok = item.health === 'Normal'
    }
    
    // 如果有分批次数据，直接使用；否则创建默认批次
    let batches = item.batches || [
      {
        name: '部署批次',
        status: item.health === 'Normal' ? '正常' : item.health === 'Warning' ? '异常' : '异常',
        start: new Date(item.startTime).toLocaleString('zh-CN'),
        end: item.endTime ? new Date(item.endTime).toLocaleString('zh-CN') : '-',
        anomaly: item.health !== 'Normal' ? `${item.service}服务指标异常` : undefined
      }
    ]
    
    return {
      id: `chg-${index + 1}`,
      service: item.service,
      version: item.version,
      state,
      progress,
      ok,
      batches
    }
  })
}

// 加载部署变更记录
const loadDeploymentChangelog = async (start?: string, limit?: number) => {
  try {
    loading.value = true
    error.value = null
    
    const response = await mockApi.getDeploymentChangelog(start, limit)
    deploymentChangelog.value = response
    
    // 转换数据格式
    changeItems.value = transformDeploymentChangelogToChangeItems(response.items)
    
    console.log('部署变更记录加载成功:', response)
  } catch (err) {
    error.value = '加载部署变更记录失败'
    console.error('加载部署变更记录失败:', err)
  } finally {
    loading.value = false
  }
}

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

// 生命周期
onMounted(() => {
  loadDeploymentChangelog()
})
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
