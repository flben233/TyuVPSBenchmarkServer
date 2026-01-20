<script setup>
const { getStatistics } = useMonitor();
const hostsData = ref([]);
const loading = ref(true);

// Fetch data initially
const data = await getStatistics();
hostsData.value = data;
loading.value = false;

// Auto-refresh every 10 seconds
const refreshInterval = ref(null);
onMounted(() => {
  refreshInterval.value = setInterval(async () => {
    const data = await getStatistics();
    hostsData.value = data;
  }, 60000);
});

onUnmounted(() => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value);
  }
});
</script>

<template>
  <div id="monitor-root">
    <el-row>
      <el-col :span="24" id="monitor-title">主机监控</el-col>
      <el-col :span="24">
        <el-skeleton :rows="5" animated v-if="loading" />
        <monitor-widget
         class="monitor-card"
          v-if="!loading"
          v-for="host in hostsData"
          :key="host.name"
          :host-data="host"
        />
        <el-empty
          v-if="!loading && hostsData.length === 0"
          description="暂无监控数据"
        />
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
#monitor-root {
  width: 100%;
  padding: 16px;
  box-sizing: border-box;
}

#monitor-title {
  font-size: 28px;
  font-weight: 300;
  margin-bottom: 16px;
}

.monitor-card {
  margin-bottom: 16px;
}
</style>
