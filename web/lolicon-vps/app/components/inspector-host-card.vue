<script setup>
import { ArrowDown, ArrowUp, Delete, EditPen } from "@element-plus/icons-vue";
import {
  formatBytes,
  formatLatency,
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

const expanded = ref(false);
const isAgentActive = computed(() => Number(props.host.uptimeSeconds) > 0);
const isOnline = computed(() => props.host.latestPing > 0);
const latestPingText = computed(() => getLatestPingValue(props.host.ping, props.host.latestPing));

const compactMetrics = computed(() => {
  if (!isAgentActive.value) {
    return [{ label: "延迟", value: latestPingText.value }];
  }

  return [
    { label: "延迟", value: latestPingText.value },
    { label: "CPU", value: formatPercent(props.host.cpuUsagePercent) },
    { label: "内存", value: formatPercent(props.host.memoryUsagePercent) },
    {
      label: "网速",
      value: `↑ ${formatNetworkSpeed(props.host.uploadMbps)} / ↓ ${formatNetworkSpeed(props.host.downloadMbps)}`,
    },
  ];
});

const detailItems = computed(() => {
  if (!isAgentActive.value) {
    return [{ label: "最新延迟", value: formatLatency(props.host.latestPing) }];
  }

  return [
    { label: "目标地址", value: props.host.target || "-" },
    { label: "系统信息", value: props.host.system || "未知系统" },
    { label: "运行时间", value: formatUptime(props.host.uptimeSeconds) },
    { label: "最新延迟", value: formatLatency(props.host.latestPing) },
  ];
});
</script>

<template>
  <el-card shadow="never" class="inspector-host-card">
    <div class="compact-row" @click="expanded = !expanded">
      <div class="host-identity">
        <span class="status-dot" :class="{ online: isOnline }"></span>
        <div class="host-text">
          <div class="host-name">{{ host.name }}</div>
          <div class="host-subtitle">
            {{ isAgentActive ? "Agent 已连接" : "未接入 Agent" }}
            <span class="host-subtitle-separator">/</span>
            {{ host.target }}
          </div>
        </div>
      </div>

      <div class="compact-metrics">
        <div v-for="item in compactMetrics" :key="item.label" class="metric-chip">
          <span class="metric-label">{{ item.label }}</span>
          <span class="metric-value">{{ item.value }}</span>
        </div>
      </div>

      <el-icon class="expand-indicator">
        <component :is="expanded ? ArrowUp : ArrowDown" />
      </el-icon>
    </div>

    <transition name="expand-fade">
      <div v-if="expanded" class="expanded-content">
        <div class="expanded-header">
          <div class="tag-row">
            <el-tag size="small" effect="dark">ID {{ host.id }}</el-tag>
            <el-tag size="small" :type="host.notify ? 'warning' : 'info'" effect="dark">
              {{ host.notify ? "通知已开启" : "通知未开启" }}
            </el-tag>
            <el-tag size="small" type="info" effect="dark">
              {{ host.lastUpdate ? `上次上报 ${formatUptime((Date.now() - new Date(host.lastUpdate).getTime()) / 1000)} 前` : "未上报" }}
            </el-tag>
            <el-tag v-for="tag in host.tags" :key="tag" size="small" effect="dark">{{ tag }}</el-tag>
          </div>

          <div class="header-actions">
            <el-button link :icon="EditPen" @click.stop="emit('edit', host)">编辑</el-button>
            <el-button link type="danger" :icon="Delete" @click.stop="emit('delete', host)">删除</el-button>
          </div>
        </div>

        <div class="detail-grid" :class="{ single: !isAgentActive }">
          <div class="detail-panel">
            <div class="panel-title">基础信息</div>
            <div v-for="item in detailItems" :key="item.label" class="detail-item">
              <span class="detail-label">{{ item.label }}</span>
              <span class="detail-value">{{ item.value }}</span>
            </div>
          </div>

          <div class="detail-panel" v-if="isAgentActive">
            <div class="panel-title">资源状态</div>
            <div class="progress-title">CPU 使用率</div>
            <el-progress :stroke-width="18" :text-inside="true" :percentage="Number(host.cpuUsagePercent.toFixed(1))" />
            <div class="spacer"></div>
            <div class="progress-title">
              内存使用率 {{ `(${formatBytes(host.memoryUsedBytes)} / ${formatBytes(host.memoryTotalBytes)})` }}
            </div>
            <el-progress :stroke-width="18" :text-inside="true" :percentage="Number(host.memoryUsagePercent.toFixed(1))" />
            <div class="spacer"></div>
            <div class="traffic-values">
              <span>入站 {{ formatTrafficAmount(host.recv) }}</span>
              <span>出站 {{ formatTrafficAmount(host.sent) }}</span>
            </div>
          </div>
        </div>

        <div class="chart-grid" :class="{ single: !isAgentActive }">
          <InspectorTrafficChart v-if="isAgentActive" :sent="host.sent" :recv="host.recv" />
          <div class="detail-panel ping-panel">
            <div class="panel-title">Ping 趋势</div>
            <InspectorPingChart :points="host.ping" />
          </div>
        </div>
      </div>
    </transition>
  </el-card>
</template>

<style scoped>
.inspector-host-card {
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.72);
  border: var(--border);
  backdrop-filter: blur(6px);
}

.compact-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  cursor: pointer;
}

.host-identity {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1 1 260px;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: rgba(144, 147, 153, 0.55);
  box-shadow: 0 0 0 4px rgba(144, 147, 153, 0.14);
  flex: none;
}

.status-dot.online {
  background: #67c23a;
  box-shadow: 0 0 0 4px rgba(103, 194, 58, 0.18);
}

.host-text {
  min-width: 0;
}

.host-name {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.host-subtitle {
  margin-top: 4px;
  color: #909399;
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.host-subtitle-separator {
  padding: 0 6px;
}

.compact-metrics {
  display: flex;
  flex: 999 1 0;
  flex-wrap: wrap;
  align-items: stretch;
  justify-content: flex-end;
  gap: 10px;
  min-width: 0;
}

.metric-chip {
  min-width: 96px;
  max-width: 100%;
  padding: 10px 12px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.72);
  border: var(--border);
  display: flex;
  flex: 1 1 96px;
  flex-direction: column;
  gap: 4px;
}

.metric-label,
.detail-label,
.progress-title {
  color: #909399;
  font-size: 12px;
}

.metric-value {
  color: var(--el-text-color-regular);
  font-size: 13px;
  font-weight: 600;
  line-height: 1.3;
  word-break: break-word;
}

.expand-indicator {
  color: #909399;
  font-size: 18px;
  flex: none;
  margin-left: auto;
}

.expanded-content {
  padding-top: 16px;
  margin-top: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.48);
}

.expanded-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.detail-grid,
.chart-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.detail-grid.single,
.chart-grid.single {
  grid-template-columns: 1fr;
}

.chart-grid {
  margin-top: 12px;
}

.detail-panel {
  background: rgba(255, 255, 255, 0.75);
  border: var(--border);
  border-radius: 6px;
  padding: 12px;
  box-sizing: border-box;
}

.panel-title {
  color: #303133;
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
}

.detail-item {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 8px;
}

.detail-item:last-child {
  margin-bottom: 0;
}

.detail-value {
  color: #303133;
  font-size: 14px;
  text-align: right;
  word-break: break-word;
}

.spacer {
  height: 8px;
}

.traffic-values {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  color: var(--el-text-color-regular);
  font-size: 14px;
}

.ping-panel :deep(.chart-wrapper),
.ping-panel :deep(.chart) {
  height: 180px;
}

@media screen and (max-width: 900px) {
  .compact-row,
  .expanded-header {
    flex-direction: column;
    align-items: stretch;
  }

  .detail-grid,
  .chart-grid {
    grid-template-columns: 1fr;
  }

  .compact-metrics,
  .header-actions {
    width: 100%;
    justify-content: stretch;
  }

  .metric-chip {
    min-width: 0;
  }

  .expand-indicator {
    margin-left: 0;
    align-self: flex-end;
  }

  .host-identity {
    flex: none;
  }
}

.expand-fade-enter-active,
.expand-fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.expand-fade-enter-from,
.expand-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
