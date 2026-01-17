<script setup>
const props = defineProps(["hostData"]);
const host = computed(() => convertHost(props.hostData));

function convertHost(d) {
  return {
    ...d,
    stats: calculateStats(d.history || []),
  };
}

function calculateStats(history) {
  if (!history || history.length === 0) {
    return {
      sent: 0,
      loss: 0,
      lossRate: "0.00%",
      newest: "-",
      fastest: "-",
      slowest: "-",
      average: "-",
    };
  }

  const sent = history.length;
  const losses = history.filter((v) => v === 0).length;
  const lossRate = ((losses / sent) * 100).toFixed(2) + "%";

  const validLatencies = history.filter((v) => v > 0);

  if (validLatencies.length === 0) {
    return {
      sent,
      loss: losses,
      lossRate,
      newest: "-",
      fastest: "-",
      slowest: "-",
      average: "-",
    };
  }

  const newest =
    history[history.length - 1] > 0
      ? history[history.length - 1] + " ms"
      : "Loss";
  const fastest = Math.min(...validLatencies) + " ms";
  const slowest = Math.max(...validLatencies) + " ms";
  const average =
    (validLatencies.reduce((a, b) => a + b, 0) / validLatencies.length).toFixed(
      2,
    ) + " ms";

  return {
    sent,
    loss: losses,
    lossRate,
    newest,
    fastest,
    slowest,
    average,
  };
}

function getGraphPoints(history) {
  if (!history || history.length === 0) return [];

  const maxLatency = Math.max(...history.filter((v) => v > 0), 1);
  const width = 200;
  const height = 40;
  const pointWidth = width / Math.max(history.length - 1, 1);

  return history.map((latency, index) => {
    const x = index * pointWidth;
    const y =
      latency > 0 ? height - (latency / maxLatency) * (height - 5) : height;
    return { x, y, latency };
  });
}

function createPath(points) {
  if (points.length === 0) return "";

  let path = `M ${points[0].x} ${points[0].y}`;
  for (let i = 1; i < points.length; i++) {
    path += ` L ${points[i].x} ${points[i].y}`;
  }
  return path;
}
</script>

<template>
  <el-card shadow="never">
    <div class="host-container">
      <div class="host-info">
        <div class="host-name">{{ host.name }}</div>
        <div class="host-uploader">上传者: {{ host.uploader }}</div>
      </div>
      <div class="host-stats">
        <div class="stat-item">
          <div class="stat-label">丢包率</div>
          <div
            class="stat-value"
            :class="{ 'stat-warning': parseFloat(host.stats.lossRate) > 0 }"
          >
            {{ host.stats.lossRate }}
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-label">发送</div>
          <div class="stat-value">{{ host.stats.sent }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">最新</div>
          <div class="stat-value">{{ host.stats.newest }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">最快</div>
          <div class="stat-value stat-success">{{ host.stats.fastest }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">最慢</div>
          <div class="stat-value stat-danger">{{ host.stats.slowest }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">平均</div>
          <div class="stat-value">{{ host.stats.average }}</div>
        </div>
      </div>
      <div class="host-graph">
        <svg width="200" height="40" class="latency-graph">
          <path
            :d="createPath(getGraphPoints(host.history))"
            fill="none"
            stroke="#39C5BB"
            stroke-width="2"
          />
          <circle
            v-for="(point, index) in getGraphPoints(host.history)"
            :key="index"
            :cx="point.x"
            :cy="point.y"
            :r="point.latency === 0 ? 3 : 2"
            :fill="point.latency === 0 ? '#F56C6C' : '#39C5BB'"
          />
        </svg>
      </div>
    </div>
  </el-card>
</template>

<style>
.monitor-card {
  margin-bottom: 16px;
}

.host-container {
  display: flex;
  align-items: center;
  gap: 24px;
  flex-wrap: wrap;
}

.host-info {
  min-width: 150px;
}

.host-name {
  font-weight: 600;
  font-size: 18px;
  margin-bottom: 4px;
  color: #303133;
}

.host-uploader {
  font-size: 14px;
  color: #909399;
}

.host-stats {
  display: flex;
  gap: 20px;
  flex: 1;
  flex-wrap: wrap;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 80px;
}

.stat-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.stat-warning {
  color: #e6a23c;
}

.stat-success {
  color: #67c23a;
}

.stat-danger {
  color: #f56c6c;
}

.host-graph {
  margin-left: auto;
}

.latency-graph {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background-color: #f5f7fa;
}

@media screen and (max-width: 768px) {
  .host-stats {
    justify-content: center;
  }
  .host-container {
    flex-direction: column;
    gap: 8px;
  }
  .host-graph {
    margin-left: 0;
    margin-top: 12px;
  }
  .host-info {
    text-align: center;
    margin-bottom: 8px;
  }
}
</style>
