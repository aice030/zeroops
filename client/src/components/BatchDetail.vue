<template>
  <div class="batch-detail" :class="{ 'expandable': batch.moduleRecords }" @click="handleBatchClick">
    <div class="batch-header">
      <div class="batch-info">
        <div class="batch-name">{{ batch.name }}</div>
        <el-tag
          :type="getBatchStatusType(batch.status)"
          size="small"
          class="batch-status"
        >
          {{ batch.status }}
        </el-tag>
      </div>
      <div v-if="batch.moduleRecords" class="expand-indicator">
        <el-icon>
          <ArrowUp v-if="isModuleExpanded" />
          <ArrowDown v-else />
        </el-icon>
      </div>
    </div>

    <div class="batch-details">
      <div class="time-info">
        <div>开始时间: {{ batch.start }}</div>
        <div>结束时间: {{ batch.end }}</div>
        <div v-if="batch.anomaly" class="anomaly-info">
          <el-icon class="warning-icon"><Warning /></el-icon>
          <span>{{ batch.anomaly }}</span>
        </div>
      </div>
    </div>

    <!-- 模块记录列表 -->
    <div v-if="isModuleExpanded" class="module-records">
      <div v-if="batch.moduleRecords && batch.moduleRecords.length > 0" class="records-list">
        <ModuleRecord
          v-for="record in batch.moduleRecords"
          :key="record.id"
          :record="record"
        />
      </div>
      <div v-else class="no-records">
        暂无模块发布记录
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { Batch } from '@/stores/app'
import ModuleRecord from './ModuleRecord.vue'

interface Props {
  batch: Batch
}

const props = defineProps<Props>()

const isModuleExpanded = ref(false)

const toggleModuleExpanded = () => {
  isModuleExpanded.value = !isModuleExpanded.value
}

const handleBatchClick = (event: Event) => {
  if (props.batch.moduleRecords) {
    event.stopPropagation() // 阻止事件冒泡到父级
    toggleModuleExpanded()
  }
}

const getBatchStatusType = (status: string) => {
  switch (status) {
    case '异常':
      return 'danger'
    case '进行中':
      return 'warning'
    case '正常':
      return 'success'
    default:
      return 'info'
  }
}
</script>

<style scoped>
.batch-detail {
  padding: 12px;
  background: #f8fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.batch-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.batch-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.batch-name {
  font-weight: 500;
  font-size: 14px;
}

.batch-status {
  font-size: 12px;
}

.expand-indicator {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
}

.expandable {
  cursor: pointer;
  transition: all 0.2s ease;
}

.expandable:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.batch-details {
  margin-bottom: 12px;
}

.time-info {
  font-size: 12px;
  color: #6b7280;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.anomaly-info {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #dc2626;
  font-weight: 500;
}

.warning-icon {
  width: 12px;
  height: 12px;
}

.module-records {
  margin-top: 12px;
}

.records-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.no-records {
  text-align: center;
  color: #6b7280;
  font-size: 14px;
  padding: 16px;
  background: #f8fafc;
  border-radius: 6px;
  border: 1px solid #e2e8f0;
}
</style>
