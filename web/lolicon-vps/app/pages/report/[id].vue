<script setup>
import MonitorWidget from "~/components/monitor-widget.vue";

const route = useRoute();
const { getReportDetails } = useReport();
const { getStatistics } = useMonitor();
const reportId = route.params.id;

const report = ref(null);
const resp = await getReportDetails(reportId);
let data = { data: null };
if (resp.data.value && resp.data.value.code === 0) {
  data = resp.data.value;
}

// Dynamic metadata based on report data
const reportTitle = computed(
  () => report.value?.title || `测试报告 #${reportId}`,
);
const reportDescription = computed(() => {
  if (!report.value) return "VPS性能测试报告详情";
  return `${report.value.title || `测试报告 #${reportId}`} - 查看详细的VPS性能测试数据，包括CPU、内存、磁盘、网络速度、回程路由、流媒体解锁等信息`;
});

useHead({
  title: () => `${reportTitle.value} - Lolicon VPS`,
  meta: [
    { name: "description", content: () => reportDescription.value },
    {
      name: "keywords",
      content:
        "VPS测试报告,性能测试详情,服务器评测,CPU测试,内存测试,磁盘测试,网络测试,回程路由,流媒体解锁",
    },
    {
      property: "og:title",
      content: () => `${reportTitle.value} - Lolicon VPS`,
    },
    { property: "og:description", content: () => reportDescription.value },
    { property: "og:type", content: "article" },
  ],
});

const speedTestLabels = ["大陆三网多线程", "大陆三网单线程", "国际方向多线程"];
const diskLabels = ["测试项目", "读速度 (MB/s)", "写速度 (MB/s)"];
const monitorData = ref(null);
report.value = data.data;

let intervalId;
onMounted(async () => {
  if (report.value?.monitor_id) {
    intervalId = setInterval(async () => {
      monitorData.value = (await getStatistics(report.value.monitor_id))[0];
    }, 60000);
    monitorData.value = (await getStatistics(report.value.monitor_id))[0];
  }
});

onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId);
  }
});

const formatDate = (dateStr) => {
  if (!dateStr) return "N/A";
  return new Date(dateStr).toLocaleString("zh-CN");
};

const getStatusColor = (status) => {
  if (!status) return "";
  const lowerStatus = status.toLowerCase();
  if (
    lowerStatus.includes("yes") ||
    lowerStatus.includes("解锁") ||
    lowerStatus.includes("success")
  ) {
    return "success";
  } else if (
    lowerStatus.includes("no") ||
    lowerStatus.includes("失败") ||
    lowerStatus.includes("blocked")
  ) {
    return "danger";
  }
  return "info";
};

const getMediaStatusColor = (status) => {
  if (
    status.includes("Banned") ||
    status.includes("Risky") ||
    status.includes("Only") ||
    status.includes("No") ||
    status.includes("Failed")
  ) {
    return "var(--el-color-danger)";
  }
  return "var(--el-color-success)";
};

const getIPScoreColor = (score) => {
  if (score.includes("%")) {
    score = parseFloat(score.replace("%", "")) / 100;
  }
  if (isNaN(score)) {
    return "var(--el-color-info)";
  }
  if (score < 1) {
    score *= 100;
  }
  if (score < 20) {
    return "var(--el-color-success)";
  } else if (score < 60) {
    return "var(--el-color-warning)";
  } else {
    return "var(--el-color-danger)";
  }
};

const goBack = () => {
  useRouter().back();
};
</script>

<template>
  <div id="report-root">
    <el-row>
      <el-col :span="24">
        <el-button @click="goBack" style="margin-bottom: 16px">
          ← 返回
        </el-button>
      </el-col>
    </el-row>

    <div v-if="report">
      <!-- Header -->
      <div id="report-header">
        <h1>{{ report.title || "测试报告" }}</h1>
        <div class="report-meta">
          <el-tag effect="dark">ID: {{ report.id }}</el-tag>
          <el-tag effect="dark" type="info" style="margin-left: 8px"
            >创建时间: {{ formatDate(report.time) }}</el-tag
          >
          <el-tag effect="dark" type="info" style="margin-left: 8px"
            >更新时间: {{ formatDate(report.updated_at) }}</el-tag
          >
        </div>
        <div v-if="report.link" class="report-link">
          <el-link :href="report.link" target="_blank" type="primary">
            商家链接
          </el-link>
        </div>
      </div>

      <monitor-widget
        v-if="monitorData"
        :host-data="monitorData"
        style="margin-bottom: 16px"
      />

      <!-- ECS Benchmark Results -->
      <div shadow="never" class="section-card" v-if="report.ecs">
        <div class="card-header">
          <span class="section-title">融合怪测试</span>
        </div>

        <!-- System Info -->
        <div v-if="report.ecs.info" class="subsection">
          <h3>系统信息</h3>
          <el-descriptions :column="2" border>
            <el-descriptions-item
              v-for="(value, key) in report.ecs.info"
              :key="key"
              :label="key"
            >
              {{ value }}
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- CPU Performance -->
        <div v-if="report.ecs.cpu" class="subsection">
          <h3>CPU 性能</h3>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-statistic
                title="单核得分"
                :value="report.ecs.cpu.single || 0"
              />
            </el-col>
            <el-col :span="12">
              <el-statistic
                title="多核得分"
                :value="report.ecs.cpu.multi || 0"
              />
            </el-col>
          </el-row>
        </div>

        <!-- Memory Performance -->
        <div v-if="report.ecs.mem" class="subsection">
          <h3>内存性能</h3>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-statistic
                title="读取速度 (MB/s)"
                :value="report.ecs.mem.read || 0"
                :precision="2"
              />
            </el-col>
            <el-col :span="12">
              <el-statistic
                title="写入速度 (MB/s)"
                :value="report.ecs.mem.write || 0"
                :precision="2"
              />
            </el-col>
          </el-row>
        </div>

        <!-- Disk Performance -->
        <div v-if="report.ecs.disk" class="subsection">
          <h3>磁盘性能</h3>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="顺序读取">{{
              report.ecs.disk.seq_read
            }}</el-descriptions-item>
            <el-descriptions-item label="顺序写入">{{
              report.ecs.disk.seq_write
            }}</el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- Mail Ports -->
        <div v-if="report.ecs.mail" class="subsection">
          <h3>邮件端口测试</h3>
          <div
            v-for="(values, port) in report.ecs.mail"
            :key="port"
            style="margin-bottom: 12px"
          >
            <h4>{{ port }}</h4>
            <div class="tag-group">
              <el-tag
                effect="dark"
                :type="values[0] ? 'success' : 'danger'"
                style="margin: 4px"
              >
                SMTP: {{ values[0] ? "✓" : "✗" }}
              </el-tag>
              <el-tag
                effect="dark"
                :type="values[1] ? 'success' : 'danger'"
                style="margin: 4px"
              >
                POP3: {{ values[1] ? "✓" : "✗" }}
              </el-tag>
              <el-tag
                effect="dark"
                :type="values[2] ? 'success' : 'danger'"
                style="margin: 4px"
              >
                IMAP: {{ values[2] ? "✓" : "✗" }}
              </el-tag>
              <el-tag
                effect="dark"
                :type="values[3] ? 'success' : 'danger'"
                style="margin: 4px"
              >
                SMTP-SSL: {{ values[3] ? "✓" : "✗" }}
              </el-tag>
              <el-tag
                effect="dark"
                :type="values[4] ? 'success' : 'danger'"
                style="margin: 4px"
              >
                POP3-SSL: {{ values[4] ? "✓" : "✗" }}
              </el-tag>
              <el-tag
                effect="dark"
                :type="values[5] ? 'success' : 'danger'"
                style="margin: 4px"
              >
                IMAP-SSL: {{ values[5] ? "✓" : "✗" }}
              </el-tag>
            </div>
          </div>
        </div>

        <!-- Network Trace -->
        <div v-if="report.ecs.trace" class="subsection">
          <h3>网络路由</h3>
          <el-descriptions :column="1" border>
            <el-descriptions-item
              v-if="report.ecs.trace.back_route"
              label="回程路由"
            >
              <pre class="code-block">{{ report.ecs.trace.back_route }}</pre>
            </el-descriptions-item>
            <el-descriptions-item
              v-if="report.ecs.trace.types"
              label="路由类型"
            >
              <div class="card-grid">
                <el-card
                  v-for="(value, key) in report.ecs.trace.types"
                  :key="key"
                  shadow="hover"
                  :class="{
                    'vip-route':
                      value.includes('精品线路') || value.includes('优质线路'),
                  }"
                >
                  {{ key }}: {{ value }}
                </el-card>
              </div>
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- TikTok & IP Quality -->
        <div class="subsection">
          <el-row :gutter="16">
            <el-col :span="12" v-if="report.ecs.tiktok">
              <h3>TikTok 解锁状态</h3>
              <el-tag
                effect="dark"
                :type="getStatusColor(report.ecs.tiktok)"
                size="large"
              >
                {{ report.ecs.tiktok }}
              </el-tag>
            </el-col>
            <el-col :span="12" v-if="report.ecs.ip_quality">
              <h3>IP 质量</h3>
              <pre class="code-block">{{ report.ecs.ip_quality }}</pre>
            </el-col>
          </el-row>
        </div>
      </div>

      <!-- Disk Test Results -->
      <div class="section-card" v-if="report.disk">
        <div class="card-header">
          <span class="section-title">磁盘测试</span>
        </div>

        <el-table :data="report.disk.data" border stripe>
          <el-table-column
            v-for="(col, index) in report.disk.data[0] || []"
            :key="index"
            :label="diskLabels[index]"
            min-width="100"
          >
            <template #default="{ row }">
              {{ row[index] }}
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- Speed Test Results -->
      <div class="section-card" v-if="report.spdtest?.length">
        <span class="section-title">速度测试</span>
        <el-tabs>
          <el-tab-pane
            v-for="(test, index) in report.spdtest"
            :key="index"
            :label="speedTestLabels[index]"
            class="subsection"
          >
            <h3 v-if="test.time">测试时间: {{ test.time }}</h3>
            <el-table :data="test.results" border stripe>
              <el-table-column prop="spot" label="测试点" min-width="150" />
              <el-table-column
                prop="download"
                label="下载速度 (Mbps)"
                min-width="150"
              >
                <template #default="{ row }">
                  {{ row.download?.toFixed(2) || "N/A" }}
                </template>
              </el-table-column>
              <el-table-column
                prop="upload"
                label="上传速度 (Mbps)"
                min-width="150"
              >
                <template #default="{ row }">
                  {{ row.upload?.toFixed(2) || "N/A" }}
                </template>
              </el-table-column>
              <el-table-column prop="latency" label="延迟 (ms)" min-width="120">
                <template #default="{ row }">
                  {{ row.latency?.toFixed(2) || "N/A" }}
                </template>
              </el-table-column>
              <el-table-column prop="jitter" label="抖动 (ms)" min-width="120">
                <template #default="{ row }">
                  {{ row.jitter?.toFixed(2) || "N/A" }}
                </template>
              </el-table-column>
            </el-table>
          </el-tab-pane>
        </el-tabs>
      </div>

      <!-- Media Unlock Results -->
      <div class="section-card" v-if="report.media">
        <span class="section-title">流媒体解锁测试</span>

        <div v-if="report.media.ipv4?.length" class="subsection">
          <h3>IPv4 解锁情况</h3>
          <div class="card-grid media-unlock-grid">
            <el-card
              shadow="never"
              v-for="(block, index) in report.media.ipv4"
              :key="'ipv4-' + index"
            >
              <h4>{{ block.region }}</h4>
              <div
                v-for="(item, idx) in block.results"
                :key="idx"
                class="media-unlock-body"
              >
                <div>{{ item.media }}</div>
                <div :style="{ color: getMediaStatusColor(item.unlock) }">
                  {{ item.unlock }}
                </div>
              </div>
            </el-card>
          </div>
        </div>

        <div v-if="report.media.ipv6?.length" class="subsection">
          <h3>IPv6 解锁情况</h3>
          <div class="card-grid media-unlock-grid">
            <el-card
              shadow="never"
              v-for="(block, index) in report.media.ipv6"
              :key="'ipv6-' + index"
            >
              <h4>{{ block.region }}</h4>
              <div
                v-for="(item, idx) in block.results"
                :key="idx"
                class="media-unlock-body"
              >
                <div>{{ item.media }}</div>
                <div :style="{ color: getMediaStatusColor(item.unlock) }">
                  {{ item.unlock }}
                </div>
              </div>
            </el-card>
          </div>
        </div>
      </div>

      <!-- IP Quality Results -->
      <div class="section-card" v-if="report.ipquality">
        <span class="section-title">IP 质量检测</span>

        <!-- Head Info -->
        <div v-if="report.ipquality.Head" class="subsection">
          <h3>检测信息</h3>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="IP">{{
              report.ipquality.Head.IP
            }}</el-descriptions-item>
            <el-descriptions-item label="检测时间">{{
              report.ipquality.Head.Time
            }}</el-descriptions-item>
            <el-descriptions-item label="版本">{{
              report.ipquality.Head.Version
            }}</el-descriptions-item>
            <el-descriptions-item label="GitHub">
              <el-link :href="report.ipquality.Head.GitHub" target="_blank">
                {{ report.ipquality.Head.GitHub }}
              </el-link>
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- Geographic Info -->
        <div v-if="report.ipquality.Info" class="subsection">
          <h3>地理位置信息</h3>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="ASN">{{
              report.ipquality.Info.ASN
            }}</el-descriptions-item>
            <el-descriptions-item label="组织">{{
              report.ipquality.Info.Organization
            }}</el-descriptions-item>
            <el-descriptions-item label="大洲">
              {{ report.ipquality.Info.Continent?.Name }} ({{
                report.ipquality.Info.Continent?.Code
              }})
            </el-descriptions-item>
            <el-descriptions-item label="国家/地区">
              {{ report.ipquality.Info.Region?.Name }} ({{
                report.ipquality.Info.Region?.Code
              }})
            </el-descriptions-item>
            <el-descriptions-item label="城市">
              {{ report.ipquality.Info.City?.Name }}
            </el-descriptions-item>
            <el-descriptions-item label="时区">{{
              report.ipquality.Info.TimeZone
            }}</el-descriptions-item>
            <el-descriptions-item label="经纬度" :span="2">
              {{ report.ipquality.Info.Latitude }},
              {{ report.ipquality.Info.Longitude }}
            </el-descriptions-item>
            <el-descriptions-item
              label="地图链接"
              :span="2"
              v-if="report.ipquality.Info.Map"
            >
              <el-link :href="report.ipquality.Info.Map" target="_blank"
                >查看地图</el-link
              >
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- Risk Scores -->
        <div v-if="report.ipquality.Score" class="subsection">
          <h3>风险评分</h3>
          <el-descriptions :column="3" border>
            <el-descriptions-item
              v-for="(value, key) in report.ipquality.Score"
              :key="key"
              :label="key"
            >
              <div :style="{ color: getIPScoreColor(value) }">{{ value }}</div>
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- Risk Factors -->
        <div v-if="report.ipquality.Factor" class="subsection">
          <h3>风险因素</h3>
          <div
            v-for="(factorData, factorName) in report.ipquality.Factor"
            :key="factorName"
          >
            <h4>{{ factorName }}</h4>
            <el-descriptions
              :column="3"
              border
              size="small"
              class="factor-descriptions"
            >
              <el-descriptions-item
                v-for="(value, provider) in factorData"
                :key="provider"
                :label="provider"
                label-width="150px"
                width="150px"
              >
                <span v-if="value === null" class="value-na">N/A</span>
                <span
                  v-else-if="typeof value === 'boolean'"
                  :class="value ? 'value-danger' : 'value-success'"
                >
                  {{ value ? "是" : "否" }}
                </span>
                <span v-else>{{ value }}</span>
              </el-descriptions-item>
            </el-descriptions>
          </div>
        </div>

        <!-- Mail Test Results -->
        <div v-if="report.ipquality.Mail" class="subsection">
          <h3>邮件服务测试</h3>
          <div class="tag-group">
            <el-tag
              effect="dark"
              v-for="(value, key) in report.ipquality.Mail"
              :key="key"
              :type="
                key === 'DNSBlacklist' ? 'info' : value ? 'success' : 'danger'
              "
              style="margin: 4px"
            >
              <template v-if="key === 'DNSBlacklist'">
                DNS黑名单: {{ value.Blacklisted }}/{{ value.Total }} 黑名单
              </template>
              <template v-else> {{ key }}: {{ value ? "✓" : "✗" }} </template>
            </el-tag>
          </div>
        </div>

        <!-- Media Unlock -->
        <div v-if="report.ipquality.Media" class="subsection">
          <h3>流媒体服务</h3>
          <el-row :gutter="16">
            <el-col
              :span="8"
              v-for="(value, key) in report.ipquality.Media"
              :key="key"
              class="media-item"
            >
              <el-card shadow="hover" class="media-card">
                <h4>{{ key }}</h4>
                <el-descriptions :column="1" size="small">
                  <el-descriptions-item label="状态">
                    <el-tag
                      effect="dark"
                      :type="getStatusColor(value.Status)"
                      >{{ value.Status }}</el-tag
                    >
                  </el-descriptions-item>
                  <el-descriptions-item label="地区">{{
                    value.Region
                  }}</el-descriptions-item>
                  <el-descriptions-item label="类型">{{
                    value.Type
                  }}</el-descriptions-item>
                </el-descriptions>
              </el-card>
            </el-col>
          </el-row>
        </div>

        <!-- Type Info -->
        <div v-if="report.ipquality.Type" class="subsection">
          <h3>IP 类型</h3>
          <el-row :gutter="16">
            <el-col :span="12" v-if="report.ipquality.Type.Company">
              <h4>公司类型</h4>
              <el-descriptions :column="1" border size="small">
                <el-descriptions-item
                  v-for="(value, key) in report.ipquality.Type.Company"
                  :key="key"
                  :label="key"
                >
                  {{ value }}
                </el-descriptions-item>
              </el-descriptions>
            </el-col>
            <el-col :span="12" v-if="report.ipquality.Type.Usage">
              <h4>使用类型</h4>
              <el-descriptions :column="1" border size="small">
                <el-descriptions-item
                  v-for="(value, key) in report.ipquality.Type.Usage"
                  :key="key"
                  :label="key"
                >
                  {{ value }}
                </el-descriptions-item>
              </el-descriptions>
            </el-col>
          </el-row>
        </div>
      </div>

      <!-- Best Trace Results -->
      <div class="section-card" v-if="report.besttrace?.length">
        <span class="section-title">回程路由追踪</span>

        <div
          v-for="(trace, index) in report.besttrace"
          :key="index"
          class="subsection"
        >
          <h3>{{ trace.region }}</h3>
          <pre class="code-block">{{ trace.route }}</pre>
        </div>
      </div>

      <!-- ITDog Results -->
      <div class="section-card" v-if="report.itdog">
        <span class="section-title">ITDog</span>

        <div v-if="report.itdog.ping" class="subsection">
          <h3>全国 Ping</h3>
          <img
            :src="report.itdog.ping"
            alt="Ping Result"
            style="max-width: 100%"
          />
        </div>

        <div v-if="report.itdog.route?.length" class="subsection">
          <h3>去程路由追踪</h3>
          <div
            v-for="(routeImg, index) in report.itdog.route"
            :key="index"
            style="margin-bottom: 16px"
          >
            <h4>路由 {{ index + 1 }}</h4>
            <img :src="routeImg" alt="Route Result" style="max-width: 100%" />
          </div>
        </div>
      </div>
    </div>

    <el-empty v-else description="未找到报告数据" />
  </div>
</template>

<style scoped>
#report-root {
  width: 100%;
  padding: 16px;
  box-sizing: border-box;
  overflow-x: hidden;
  overflow-y: auto;
}

#report-header {
  margin-bottom: 16px;
}

#report-header h1 {
  font-size: 32px;
  font-weight: 600;
  margin: 0 0 16px 0;
  color: #303133;
}

.report-meta {
  margin-bottom: 12px;
}

.report-link {
  margin-top: 12px;
}

.section-card {
  margin-bottom: 24px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
  display: inline-block;
}

.subsection {
  margin-bottom: 24px;
}

.subsection:last-child {
  margin-bottom: 0;
}

.subsection h3 {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 12px 0;
  color: #606266;
}

.subsection h4 {
  font-size: 14px;
  font-weight: 600;
  margin: 12px 0 8px 0;
  color: #909399;
}

.tag-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.code-block {
  background-color: #f5f7fa;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 12px;
  font-family: "Courier New", Courier, monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-wrap: break-word;
  color: #303133;
  margin: 0;
  max-height: 400px;
  overflow-y: auto;
}

.media-card {
  height: 100%;
}

.media-item {
  margin: 8px 0;
}

.media-card h4 {
  margin: 0 0 12px 0;
  font-size: 16px;
  color: #303133;
}

.vip-route {
  color: var(--el-color-success);
  background-color: var(--el-color-success-light-8);
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 16px;
}

.media-unlock-body {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  margin: 8px 0;
  padding: 8px;
  color: #303133;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.media-unlock-grid :deep(.el-card__body) {
  padding: 12px;
}

.media-unlock-grid h4 {
  margin: 2px 0 8px 0;
}

:deep(.el-descriptions__label) {
  font-weight: 600;
}

:deep(.el-statistic__head) {
  font-size: 14px;
  color: #909399;
}

:deep(.el-statistic__content) {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.value-na {
  color: #909399;
}

.value-success {
  color: #67c23a;
  font-weight: 600;
}

.value-danger {
  color: #f56c6c;
  font-weight: 600;
}

.factor-descriptions :deep(.factor-label) {
  width: 150px;
  min-width: 150px;
}

:deep(.el-tag) {
  border: none;
}
</style>
