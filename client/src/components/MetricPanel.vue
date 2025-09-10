<template>
  <el-card class="metric-panel" shadow="hover">
    <template #header>
      <div class="metric-header">
        <div class="metric-title">{{ title }}</div>
        <div v-if="note" class="metric-note">{{ note }}</div>
      </div>
    </template>
    
    <div class="metric-content">
      <div class="metric-value">
        {{ lastValue }} <span class="metric-unit">{{ unit }}</span>
      </div>
      
      <div ref="chartRef" class="metric-chart"></div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'

interface Props {
  title: string
  note?: string
  unit: string
  dataKey: 'latency' | 'traffic' | 'errors' | 'saturation'
}

const props = defineProps<Props>()

const chartRef = ref<HTMLElement>()
let chart: echarts.ECharts | null = null

// 生成模拟数据
const generateSampleData = () => {
  const data = []
  let baseValue = 0
  
  // 根据不同的指标设置不同的基础值
  switch (props.dataKey) {
    case 'latency':
      baseValue = 120
      break
    case 'traffic':
      baseValue = 380
      break
    case 'errors':
      baseValue = 0.9
      break
    case 'saturation':
      baseValue = 63
      break
  }
  
  for (let i = 0; i < 40; i++) {
    let value = baseValue
    
    // 添加随机波动
    switch (props.dataKey) {
      case 'latency':
        value = jitter(baseValue, 1.5, 80, 320)
        break
      case 'traffic':
        value = jitter(baseValue, 6, 120, 1200)
        break
      case 'errors':
        value = clamp(jitter(baseValue, 0.2, 0.1, 5), 0.1, 5)
        break
      case 'saturation':
        value = clamp(jitter(baseValue, 1.2, 30, 95), 30, 95)
        break
    }
    
    data.push({
      time: i,
      value: round(value, props.dataKey === 'errors' ? 2 : 0)
    })
    
    baseValue = value
  }
  
  return data
}

// 工具函数
const jitter = (value: number, delta: number, min: number, max: number) => {
  const newValue = value + (Math.random() - 0.5) * delta * 2
  return Math.max(min, Math.min(max, newValue))
}

const clamp = (value: number, min: number, max: number) => {
  return Math.max(min, Math.min(max, value))
}

const round = (value: number, precision = 0) => {
  const multiplier = Math.pow(10, precision)
  return Math.round(value * multiplier) / multiplier
}

const data = ref(generateSampleData())
const lastValue = ref(data.value[data.value.length - 1]?.value || 0)

// 初始化图表
const initChart = () => {
  if (chartRef.value) {
    chart = echarts.init(chartRef.value)
    
    const option = {
      grid: {
        left: 0,
        right: 0,
        top: 8,
        bottom: 0,
        containLabel: false
      },
      xAxis: {
        type: 'category',
        data: data.value.map(d => d.time),
        show: false
      },
      yAxis: {
        type: 'value',
        show: false
      },
      series: [{
        data: data.value.map(d => d.value),
        type: 'line',
        smooth: true,
        lineStyle: {
          width: 2
        },
        areaStyle: {
          opacity: 0.1
        },
        symbol: 'none'
      }],
      tooltip: {
        trigger: 'axis',
        formatter: (params: any) => {
          const data = params[0]
          return `${props.title}: ${data.value} ${props.unit}`
        }
      }
    }
    
    chart.setOption(option)
  }
}

// 监听窗口大小变化
const handleResize = () => {
  if (chart) {
    chart.resize()
  }
}

onMounted(() => {
  initChart()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  if (chart) {
    chart.dispose()
  }
  window.removeEventListener('resize', handleResize)
})

// 监听数据变化
watch(() => props.dataKey, () => {
  data.value = generateSampleData()
  lastValue.value = data.value[data.value.length - 1]?.value || 0
  if (chart) {
    const option = {
      xAxis: {
        data: data.value.map(d => d.time)
      },
      series: [{
        data: data.value.map(d => d.value)
      }]
    }
    chart.setOption(option)
  }
})
</script>

<style scoped>
.metric-panel {
  height: 100%;
}

.metric-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.metric-title {
  font-size: 14px;
  font-weight: 500;
}

.metric-note {
  font-size: 12px;
  color: #6b7280;
  font-weight: normal;
}

.metric-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.metric-value {
  font-size: 24px;
  font-weight: 600;
  color: #1f2937;
}

.metric-unit {
  font-size: 16px;
  color: #6b7280;
  font-weight: normal;
}

.metric-chart {
  height: 160px;
  width: 100%;
}
</style>
