<template>
  <el-card class="alarm-change-card" shadow="hover">
    <div class="card-content">
      <div class="card-header">
        <div class="alarm-info">
          <div class="service-name">{{ item.service }}</div>
          <div class="change-description">{{ item.change }}</div>
          <div class="timestamp">{{ item.timestamp }}</div>
        </div>
        <el-button
          type="text"
          size="small"
          @click="toggleExpanded"
          class="expand-btn"
        >
          <el-icon>
            <ArrowUp v-if="isExpanded" />
            <ArrowDown v-else />
          </el-icon>
        </el-button>
      </div>

      <!-- 详细信息 -->
      <div v-if="isExpanded && item.details" class="details-section">
        <div class="details-content">
          {{ item.details }}
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { AlarmChangeItem } from '@/stores/app'

interface Props {
  item: AlarmChangeItem
}

const props = defineProps<Props>()

const isExpanded = ref(false)

const toggleExpanded = () => {
  isExpanded.value = !isExpanded.value
}
</script>

<style scoped>
.alarm-change-card {
  margin-bottom: 16px;
}

.card-content {
  padding: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.alarm-info {
  flex: 1;
}

.service-name {
  font-weight: 500;
  font-size: 14px;
  margin-bottom: 4px;
}

.change-description {
  font-size: 14px;
  color: #374151;
  margin-bottom: 4px;
}

.timestamp {
  font-size: 12px;
  color: #6b7280;
}

.expand-btn {
  width: 32px;
  height: 32px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-left: 16px;
}

.details-section {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e5e7eb;
}

.details-content {
  padding: 12px;
  background: #f8fafc;
  border-radius: 6px;
  border: 1px solid #e2e8f0;
  font-size: 14px;
  color: #4b5563;
  line-height: 1.5;
}
</style>
