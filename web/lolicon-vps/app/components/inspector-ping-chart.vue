<script setup>
import VChart from "vue-echarts";
import { use } from "echarts/core";
import { CanvasRenderer } from "echarts/renderers";
import { LineChart } from "echarts/charts";
import { GridComponent, TooltipComponent } from "echarts/components";
import { formatTimestamp } from "~/utils/inspector";

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent]);

const props = defineProps({
  points: {
    type: Array,
    default: () => [],
  },
});

const hasData = computed(() => Array.isArray(props.points) && props.points.length > 0);

const chartOption = computed(() => {
  const xAxisData = props.points.map((point) => formatTimestamp(point.time));
  const seriesData = props.points.map((point) => Number(point.latency || 0));

  return {
    animation: true,
    grid: {
      top: 8,
      right: 6,
      bottom: 12,
      left: 20,
    },
    tooltip: {
      trigger: "axis",
      valueFormatter: (value) => `${Number(value || 0).toFixed(1)} ms`,
    },
    xAxis: {
      type: "category",
      boundaryGap: false,
      data: xAxisData,
      axisLabel: {
        color: "#909399",
        fontSize: 11,
      },
      axisLine: {
        lineStyle: { color: "#dcdfe6" },
      },
    },
    yAxis: {
      type: "value",
      min: 0,
      splitNumber: 3,
      axisLabel: {
        color: "#909399",
        fontSize: 11,
        formatter: "{value} ms",
      },
      splitLine: {
        lineStyle: { color: "#ebeef5" },
      },
    },
    series: [
      {
        type: "line",
        smooth: true,
        showSymbol: false,
        symbolSize: 6,
        lineStyle: {
          color: "#39c5bb",
          width: 2,
        },
        areaStyle: {
          color: "rgba(57, 197, 187, 0.15)",
        },
        data: seriesData,
      },
    ],
  };
});
</script>

<template>
  <div class="inspector-chart-panel">
    <div v-if="hasData" class="chart-wrapper">
      <ClientOnly>
        <VChart :option="chartOption" autoresize class="chart" />
      </ClientOnly>
    </div>
    <div v-else class="empty-chart">暂无 Ping 数据</div>
  </div>
</template>

<style scoped>
.inspector-chart-panel {
  box-sizing: border-box;
}

.chart-wrapper,
.chart {
  width: 100%;
  height: 120px;
}

.empty-chart {
  min-height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
}
</style>
