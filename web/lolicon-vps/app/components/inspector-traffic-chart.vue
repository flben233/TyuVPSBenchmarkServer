<script setup>
import VChart from "vue-echarts";
import { use } from "echarts/core";
import { CanvasRenderer } from "echarts/renderers";
import { PieChart } from "echarts/charts";
import { LegendComponent, TooltipComponent } from "echarts/components";
import { formatTrafficAmount } from "~/utils/inspector";
import { TitleComponent } from 'echarts/components';

use([CanvasRenderer, PieChart, LegendComponent, TooltipComponent, TitleComponent]);

const props = defineProps({
  sent: {
    type: Number,
    default: 0,
  },
  recv: {
    type: Number,
    default: 0,
  },
});

const hasData = computed(() => Number(props.sent) > 0 || Number(props.recv) > 0);
const totalTraffic = computed(() => formatTrafficAmount(Number(props.sent) + Number(props.recv)));

const chartOption = computed(() => ({
  animation: true,
  tooltip: {
    trigger: "item",
    formatter: ({ name, value, percent }) => `${name}<br/>${formatTrafficAmount(value)} (${percent}%)`,
  },
  legend: {
    bottom: 0,
    itemWidth: 12,
    textStyle: { color: "#606266" },
  },
  title: {
    text: totalTraffic.value,
    left: 'center',
    top: 'center',
    textStyle: {
      fontSize: 13,
      color: '#333',
      fontWeight: 'bold'
    }
  },
  series: [
    {
      type: "pie",
      radius: ["40%", "65%"],
      avoidLabelOverlap: true,
      label: {
        formatter: ({ percent }) => `${percent}%`,
        color: "#303133",
      },
      padAngle: 3,
      itemStyle: {
        borderRadius: 6
      },
      data: [
        { value: Number(props.recv || 0), name: "入站", itemStyle: { color: "#67c23a" } },
        { value: Number(props.sent || 0), name: "出站", itemStyle: { color: "#409eff" } },
      ],
    },
  ],
}));
</script>

<template>
  <div class="inspector-chart-panel">
    <div class="panel-title">流量统计</div>
    <div v-if="hasData" class="chart-wrapper">
      <ClientOnly>
        <VChart :option="chartOption" autoresize class="chart" />
      </ClientOnly>
    </div>
    <div v-else class="empty-chart">暂无流量数据</div>
  </div>
</template>

<style scoped>
.inspector-chart-panel {
  background: rgba(255, 255, 255, 0.78);
  border: var(--border);
  border-radius: 6px;
  padding: 12px;
  box-sizing: border-box;
}

.panel-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
}

.chart-wrapper,
.chart {
  width: 100%;
  height: 180px;
}

.empty-chart {
  min-height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
  background: rgba(245, 247, 250, 0.85);
  border-radius: 12px;
}
</style>
