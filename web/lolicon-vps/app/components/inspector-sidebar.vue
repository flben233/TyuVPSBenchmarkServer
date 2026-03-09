<script setup>
import {
  Clock,
  Connection,
  DataBoard,
  Filter,
  PieChart,
  TrendCharts,
} from "@element-plus/icons-vue";

const props = defineProps({
  stats: {
    type: Object,
    default: () => ({
      hostCountText: "0 / 0",
      onlineCountText: "0 台",
      totalTrafficText: "0 MB",
      storageUsageText: "0%",
      storageUsageDetail: "0 B / 0 B",
      averageCpuText: "0.0%",
      averageMemoryText: "0.0%",
    }),
  },
  range: {
    type: Array,
    default: () => [],
  },
  interval: {
    type: String,
    default: "1h",
  },
  intervalOptions: {
    type: Array,
    default: () => [],
  },
  tags: {
    type: Array,
    default: () => [],
  },
  selectedTags: {
    type: Array,
    default: () => [],
  },
  loading: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits([
  "update:range",
  "update:interval",
  "update:selectedTags",
  "apply",
  "reset",
  "preset",
]);

const rangeModel = computed({
  get: () => props.range,
  set: (value) => emit("update:range", value),
});

const intervalModel = computed({
  get: () => props.interval,
  set: (value) => emit("update:interval", value),
});

const selectedTagsModel = computed({
  get: () => props.selectedTags,
  set: (value) => emit("update:selectedTags", value),
});

const overviewItems = computed(() => [
  {
    icon: Connection,
    label: "在线主机",
    value: props.stats.onlineCountText,
    detail: `当前展示 ${props.stats.hostCountText}`,
  },
  {
    icon: PieChart,
    label: "总流量",
    value: props.stats.totalTrafficText,
    detail: "累计入站 + 出站",
  },
  {
    icon: DataBoard,
    label: "总内存占比",
    value: props.stats.storageUsageText,
    detail: props.stats.storageUsageDetail,
  },
  {
    icon: TrendCharts,
    label: "平均 CPU",
    value: props.stats.averageCpuText,
    detail: `平均内存 ${props.stats.averageMemoryText}`,
  },
]);
</script>

<template>
  <div class="inspector-sidebar">
    <el-card shadow="never" class="sidebar-card">


      <div class="overview-list">
        <div v-for="item in overviewItems" :key="item.label" class="overview-item">
          <div class="overview-icon">
            <el-icon><component :is="item.icon" /></el-icon>
          </div>
          <div class="overview-content">
            <div class="overview-label">{{ item.label }}</div>
            <div class="overview-value">{{ item.value }}</div>
            <div class="overview-detail">{{ item.detail }}</div>
          </div>
        </div>
      </div>
    </el-card>

    <el-card shadow="never" class="sidebar-card">
      <div class="section-title">
        <el-icon><Clock /></el-icon>
        <span>时间范围</span>
      </div>

      <div class="preset-row">
        <el-button link @click="emit('preset', '6h')">近 6 小时</el-button>
        <el-button link @click="emit('preset', '24h')">近 24 小时</el-button>
        <el-button link @click="emit('preset', '7d')">近 7 天</el-button>
      </div>

      <el-form label-position="top">
        <el-form-item label="查询时间段">
          <el-date-picker
            v-model="rangeModel"
            type="datetimerange"
            unlink-panels
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            format="YYYY-MM-DD HH:mm"
            class="full-width"
          />
        </el-form-item>

        <el-form-item label="时间粒度">
          <el-select v-model="intervalModel" class="full-width">
            <el-option
              v-for="option in intervalOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
        </el-form-item>
      </el-form>

      <div class="action-row">
        <el-button link @click="emit('reset')">重置</el-button>
        <el-button link type="primary" :loading="loading" @click="emit('apply')">应用</el-button>
      </div>
    </el-card>

    <el-card shadow="never" class="sidebar-card">
      <div class="section-title">
        <el-icon><Filter /></el-icon>
        <span>标签筛选</span>
      </div>

      <el-select
        v-model="selectedTagsModel"
        multiple
        collapse-tags
        collapse-tags-tooltip
        placeholder="选择需要显示的标签"
        class="full-width"
      >
        <el-option v-for="tag in tags" :key="tag" :label="tag" :value="tag" />
      </el-select>

      <div v-if="tags.length === 0" class="empty-tip">当前没有可用标签</div>
    </el-card>
  </div>
</template>

<style scoped>
.inspector-sidebar {
  display: grid;
  gap: 16px;
  position: sticky;
  top: 16px;
}

.sidebar-card {
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.72);
  border: var(--border);
  backdrop-filter: blur(6px);
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}

.overview-list {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(2, 1fr);
}

.overview-item {
  display: flex;
  padding: 12px 0;
}

.overview-icon {
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  background: rgba(57, 197, 187, 0.12);
  color: #39c5bb;
  font-size: 18px;
}

.overview-content {
  min-width: 0;
  margin-left: 8px;
}

.overview-label {
  color: #909399;
  font-size: 12px;
  margin-bottom: 4px;
}

.overview-value {
  color: #303133;
  font-size: 18px;
  font-weight: 600;
  line-height: 1.2;
}

.overview-detail {
  margin-top: 4px;
  color: #606266;
  font-size: 12px;
}

.preset-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 8px;
}

.preset-row :deep(.el-button) {
  margin-left: 0;
  padding: 2px 0;
}

.full-width {
  width: 100%;
}

.action-row {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.empty-tip {
  margin-top: 10px;
  color: #909399;
  font-size: 12px;
}

@media screen and (max-width: 1024px) {
  .inspector-sidebar {
    position: static;
  }
}
</style>
