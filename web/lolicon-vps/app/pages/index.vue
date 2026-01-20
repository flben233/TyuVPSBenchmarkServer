<script setup>
// const { userInfo, token, login } = useAuth();
const { listReports } = useReport();
const reports = ref([]);
const page = ref(1);
const disabled = ref(false);
const total = ref(0);
const loading = ref(false);
const pageSize = 7;
const resp = await listReports(page.value, pageSize);
let reportsData;
if (reportsData.data.value && reportsData.data.value.code === 0) {
  reportsData = reportsData.data.value;
}
reports.value = reportsData.data || [];
total.value = reportsData.total || 0;
const { getServerStatus } = useMonitor();
const fmtStatus = (status) => {
  return {
    uptime: fmtSecond(status.uptime_seconds),
    cpuUsage: `${status.cpu_usage_percent.toFixed(2)}%`,
    memoryUsage: `${status.memory_usage_percent.toFixed(2)}%`,
    uploadSpeed: `${status.upload_mbps.toFixed(2)} Mbps`,
    downloadSpeed: `${status.download_mbps.toFixed(2)} Mbps`,
  };
};
const status = ref({
  uptime: "0天 1时 23分 45秒",
  cpuUsage: "56%",
  memoryUsage: "78%",
  uploadSpeed: "9 Mbps",
  downloadSpeed: "1 Mbps",
});
const statusData = await getServerStatus();
if (statusData) {
  status.value = fmtStatus(statusData); 
}

const gotoDetail = (reportId) => {
  useRouter().push(`/report/${reportId}`);
};

watch(page, async (newPage) => {
  loading.value = true;
  disabled.value = true;
  const resp = await listReports(newPage, pageSize);
  let reportsData;
  if (resp.data.value && resp.data.value.code === 0) {
    reportsData = resp.data.value;
  }
  reports.value = reportsData.data || [];
  disabled.value = false;
  loading.value = false;
});

let intervalId;
onMounted(async () => {
  intervalId = setInterval(async () => {
    const statusData = await getServerStatus();
    status.value = fmtStatus(statusData);
  }, 10000);
});

onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId);
  }
});
</script>

<template>
  <div id="index-root">
    <el-row>
      <el-col :span="24" id="report-title"> 测试记录 </el-col>
      <el-col :span="17" :xs="24">
        <el-skeleton :rows="5" animated v-if="loading" />
        <el-card
          v-if="!loading"
          shadow="never"
          class="report-item"
          v-for="report in reports"
          :key="report.id"
          @click="gotoDetail(report.id)"
        >
          <div class="report-item-header">{{ report.name }}</div>
          <div>创建时间: {{ report.date }}</div>
        </el-card>
        <el-pagination
          v-model:current-page="page"
          :disabled="disabled"
          :background="false"
          layout="total, prev, pager, next, jumper"
          :total="total"
          :page-size="pageSize"
        />
      </el-col>
      <el-col :span="6" :xs="0" :offset="1">
        <Profile>
          <ClientOnly>
            <div>
              <div style="font-weight: 600; color: #303133">服务器状态</div>
              <div class="s-item">
                <div>开机时间</div>
                <div>{{ status.uptime }}</div>
              </div>
              <div class="s-item">
                <div>CPU占用</div>
                <div>{{ status.cpuUsage }}</div>
              </div>
              <div class="s-item">
                <div>内存占用</div>
                <div>{{ status.memoryUsage }}</div>
              </div>
              <div class="s-item">
                <div>上传速度</div>
                <div>{{ status.uploadSpeed }}</div>
              </div>
              <div class="s-item">
                <div>下载速度</div>
                <div>{{ status.downloadSpeed }}</div>
              </div>
            </div>
          </ClientOnly>
        </Profile>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
#index-root {
  width: 100%;
  padding: 16px;
  box-sizing: border-box;
  overflow-x: hidden;
}
#report-title {
  font-size: 28px;
  font-weight: 300;
  margin-bottom: 16px;
}
.report-item {
  margin-bottom: 16px;
  cursor: pointer;
}
.report-item:hover {
  border-color: var(--el-color-primary);
}
.report-item-header {
  font-weight: 600;
  font-size: 18px;
  margin-bottom: 8px;
}
#site-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 16px;
  color: #303133;
}
#site-owner-container {
  margin-bottom: 24px;
  display: flex;
  align-items: center;
}
#site-owner {
  margin-left: 8px;
  font-size: 16px;
  font-weight: 500;
  justify-content: space-between;
  display: flex;
  flex-direction: column;
  height: 56px;
  box-sizing: border-box;
  padding: 4px;
}
.a-icon {
  width: 13px;
  height: 13px;
  margin-right: 2px;
}
.s-item {
  margin-top: 8px;
  font-size: 14px;
  display: grid;
  grid-template-columns: auto auto;
  justify-content: space-between;
  color: #303133;
}
</style>
