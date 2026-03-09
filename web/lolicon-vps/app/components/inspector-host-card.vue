<script setup>
import {
  formatBytes,
  formatNetworkSpeed,
  formatPercent,
  formatTrafficAmount,
  formatUptime,
  getLatestPingValue,
} from "~/utils/inspector";

const props = defineProps({
  host: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(["edit", "delete"]);

const isAgentActive = computed(() => Number(props.host.uptimeSeconds) > 0);
let summaryItems;

if (isAgentActive.value) {
  summaryItems = computed(() => [
    { label: "系统信息", value: props.host.system || "未知系统" },
    { label: "上传速率", value: formatNetworkSpeed(props.host.uploadMbps) },
    { label: "下载速率", value: formatNetworkSpeed(props.host.downloadMbps) },
    { label: "运行时间", value: formatUptime(props.host.uptimeSeconds) },
    { label: "最新延迟", value: getLatestPingValue(props.host.ping) },
  ]);
} else {
  summaryItems = computed(() => [
    { label: "最新延迟", value: getLatestPingValue(props.host.ping) },
  ]);
}
</script>

<template>
  <el-card shadow="never" class="inspector-host-card">
    <div class="card-header">
      <div class="header-title-row">
        <div class="host-name">{{ host.name }}</div>
        <div class="header-actions">
          <el-button link @click="emit('edit', host)">编辑</el-button>
          <el-button link type="danger" @click="emit('delete', host)">删除</el-button>
        </div>
      </div>
      <div class="tag-row">
        <el-tag size="small" effect="dark">ID {{ host.id }}</el-tag>
        <el-tag size="small" :type="host.notify ? 'warning' : 'info'" effect="dark">
          {{ host.notify ? '通知已开启' : '通知未开启' }}
        </el-tag>
        <el-tag size="small" :type="isAgentActive ? 'success' : 'info'" effect="dark">
          {{ isAgentActive ? 'Agent 已连接' : '未接入 Agent' }}
        </el-tag>
      </div>
      <div class="tag-row" v-if="host.tags.length > 0">
        <el-tag v-for="tag in host.tags" :key="tag" size="small" effect="dark">{{ tag }}</el-tag>
      </div>
    </div>


    <div class="summary-grid">
      <div v-for="item in summaryItems" :key="item.label" class="summary-item">
        <div class="summary-label">{{ item.label }}</div>
        <div class="summary-value">{{ item.value }}</div>
      </div>
    </div>

    <div class="progress-grid" v-if="isAgentActive">
      <div class="progress-panel">
        <div class="progress-title">CPU 使用率</div>
        <el-progress :stroke-width="20" :text-inside="true" :percentage="Number(host.cpuUsagePercent.toFixed(1))" />
        <div style="height: 8px"></div>
        <div class="progress-title">内存使用率 {{`(${formatBytes(props.host.memoryUsedBytes)} / ${formatBytes(props.host.memoryTotalBytes)})`}}</div>
        <el-progress :stroke-width="20" :text-inside="true" :percentage="Number(host.memoryUsagePercent.toFixed(1))" />
        <div style="height: 8px"></div>
        <div class="progress-title">累计流量</div>
        <div class="traffic-values">
          <span>入站 {{ formatTrafficAmount(host.recv) }}</span>
          <span>出站 {{ formatTrafficAmount(host.sent) }}</span>
        </div>
      </div>
    </div>

    <div class="progress-panel">
      <InspectorPingChart :points="host.ping" />
    </div>
  </el-card>
</template>

<style scoped>
.inspector-host-card {
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid rgba(255, 255, 255, 0.65);
  backdrop-filter: blur(12px);
}

.card-header {
  margin-bottom: 16px;
}

.header-title-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  justify-content: space-between;
}

.host-name {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}


.summary-item {
  display: flex;
  width: 100%;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 8px;
}

.summary-item:last-child {
  margin-bottom: 0;
}

.summary-grid,
.progress-panel {
  background: rgba(255, 255, 255, 0.75);
  border: 1px solid rgba(255, 255, 255, 0.55);
  border-radius: 6px;
  padding: 8px;
  box-sizing: border-box;
}

.summary-label,
.progress-title {
  color: #909399;
  font-size: 12px;
}

.progress-title {
  margin-bottom: 8px;
}

.summary-value {
  color: var(--el-text-color-regular);
  font-size: 14px;
}

.progress-grid,
.chart-grid {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: 16px;
}

.chart-grid {
  padding: 8px;
  box-sizing: border-box;
}

.summary-grid,
.progress-grid {
  margin-bottom: 8px;
}

.traffic-values {
  display: flex;
  gap: 8px;
  color: var(--el-text-color-regular);
  font-size: 14px;
}

.chart-grid > :deep(*) {
  min-width: 0;
}

.chart-grid > :first-child {
  grid-column: span 2;
}

@media screen and (max-width: 1280px) {
  .chart-grid > :first-child {
    grid-column: span 1;
  }
}

@media screen and (max-width: 768px) {
  .card-header {
    flex-direction: column;
  }

  .header-actions,
  .summary-grid,
  .progress-grid,
  .chart-grid {
    width: 100%;
    grid-template-columns: 1fr;
  }
}
</style>
