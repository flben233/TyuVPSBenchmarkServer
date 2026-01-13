<script setup>
const route = useRoute();
const { getReportDetails } = useReport();
const reportId = route.params.id;

const report = ref(null);
const loading = ref(true);
const data = await getReportDetails(reportId);
report.value = data.data;
loading.value = false;

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

    <el-skeleton :loading="loading" animated>
      <template #template>
        <el-skeleton-item
          variant="h1"
          style="width: 50%; margin-bottom: 16px"
        />
        <el-skeleton-item
          variant="text"
          style="width: 30%; margin-bottom: 32px"
        />
        <el-skeleton-item
          variant="rect"
          style="height: 200px; margin-bottom: 16px"
        />
        <el-skeleton-item
          variant="rect"
          style="height: 200px; margin-bottom: 16px"
        />
        <el-skeleton-item variant="rect" style="height: 200px" />
      </template>

      <template #default>
        <div v-if="report">
          <!-- Header -->
          <div id="report-header">
            <h1>{{ report.title || "测试报告" }}</h1>
            <div class="report-meta">
              <el-tag>ID: {{ report.id }}</el-tag>
              <el-tag type="info" style="margin-left: 8px"
                >创建时间: {{ formatDate(report.created_at) }}</el-tag
              >
              <el-tag type="info" style="margin-left: 8px"
                >更新时间: {{ formatDate(report.updated_at) }}</el-tag
              >
            </div>
            <div v-if="report.link" class="report-link">
              <el-link :href="report.link" target="_blank" type="primary">
                商家链接
              </el-link>
            </div>
          </div>

          <!-- ECS Benchmark Results -->
          <el-card shadow="never" class="section-card" v-if="report.ecs?.data">
            <template #header>
              <div class="card-header">
                <span class="section-title">ECS 性能测试</span>
                <el-tag v-if="report.ecs.data.time" size="small">{{
                  report.ecs.data.time
                }}</el-tag>
              </div>
            </template>

            <!-- System Info -->
            <div v-if="report.ecs.data.info" class="subsection">
              <h3>系统信息</h3>
              <el-descriptions :column="2" border>
                <el-descriptions-item
                  v-for="(value, key) in report.ecs.data.info"
                  :key="key"
                  :label="key"
                >
                  {{ value }}
                </el-descriptions-item>
              </el-descriptions>
            </div>

            <!-- CPU Performance -->
            <div v-if="report.ecs.data.cpu" class="subsection">
              <h3>CPU 性能</h3>
              <el-row :gutter="16">
                <el-col :span="12">
                  <el-statistic
                    title="单核得分"
                    :value="report.ecs.data.cpu.single || 0"
                  />
                </el-col>
                <el-col :span="12">
                  <el-statistic
                    title="多核得分"
                    :value="report.ecs.data.cpu.multi || 0"
                  />
                </el-col>
              </el-row>
            </div>

            <!-- Memory Performance -->
            <div v-if="report.ecs.data.mem" class="subsection">
              <h3>内存性能</h3>
              <el-row :gutter="16">
                <el-col :span="12">
                  <el-statistic
                    title="读取速度 (MB/s)"
                    :value="report.ecs.data.mem.read || 0"
                    :precision="2"
                  />
                </el-col>
                <el-col :span="12">
                  <el-statistic
                    title="写入速度 (MB/s)"
                    :value="report.ecs.data.mem.write || 0"
                    :precision="2"
                  />
                </el-col>
              </el-row>
            </div>

            <!-- Disk Performance -->
            <div v-if="report.ecs.data.disk" class="subsection">
              <h3>磁盘性能</h3>
              <el-descriptions :column="2" border>
                <el-descriptions-item label="顺序读取">{{
                  report.ecs.data.disk.seq_read
                }}</el-descriptions-item>
                <el-descriptions-item label="顺序写入">{{
                  report.ecs.data.disk.seq_write
                }}</el-descriptions-item>
              </el-descriptions>
            </div>

            <!-- Mail Ports -->
            <div v-if="report.ecs.data.mail" class="subsection">
              <h3>邮件端口测试</h3>
              <div class="tag-group">
                <el-tag
                  v-for="(values, port) in report.ecs.data.mail"
                  :key="port"
                  :type="values.some((v) => v) ? 'success' : 'danger'"
                  style="margin: 4px"
                >
                  {{ port }}: {{ values.filter((v) => v).length }}/{{
                    values.length
                  }}
                  可用
                </el-tag>
              </div>
            </div>

            <!-- Network Trace -->
            <div v-if="report.ecs.data.trace" class="subsection">
              <h3>网络路由</h3>
              <el-descriptions :column="1" border>
                <el-descriptions-item
                  v-if="report.ecs.data.trace.back_route"
                  label="回程路由"
                >
                  <pre class="code-block">{{
                    report.ecs.data.trace.back_route
                  }}</pre>
                </el-descriptions-item>
                <el-descriptions-item
                  v-if="report.ecs.data.trace.types"
                  label="路由类型"
                >
                  <div class="tag-group">
                    <el-tag
                      v-for="(value, key) in report.ecs.data.trace.types"
                      :key="key"
                      style="margin: 4px"
                    >
                      {{ key }}: {{ value }}
                    </el-tag>
                  </div>
                </el-descriptions-item>
              </el-descriptions>
            </div>

            <!-- TikTok & IP Quality -->
            <div class="subsection">
              <el-row :gutter="16">
                <el-col :span="12" v-if="report.ecs.data.tiktok">
                  <h3>TikTok 解锁状态</h3>
                  <el-tag
                    :type="getStatusColor(report.ecs.data.tiktok)"
                    size="large"
                  >
                    {{ report.ecs.data.tiktok }}
                  </el-tag>
                </el-col>
                <el-col :span="12" v-if="report.ecs.data.ip_quality">
                  <h3>IP 质量</h3>
                  <pre class="code-block">{{ report.ecs.data.ip_quality }}</pre>
                </el-col>
              </el-row>
            </div>
          </el-card>

          <!-- Speed Test Results -->
          <el-card
            shadow="never"
            class="section-card"
            v-if="report.spdtest?.data?.length"
          >
            <template #header>
              <span class="section-title">速度测试</span>
            </template>

            <div
              v-for="(test, index) in report.spdtest.data"
              :key="index"
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
                <el-table-column
                  prop="latency"
                  label="延迟 (ms)"
                  min-width="120"
                >
                  <template #default="{ row }">
                    {{ row.latency?.toFixed(2) || "N/A" }}
                  </template>
                </el-table-column>
                <el-table-column
                  prop="jitter"
                  label="抖动 (ms)"
                  min-width="120"
                >
                  <template #default="{ row }">
                    {{ row.jitter?.toFixed(2) || "N/A" }}
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-card>

          <!-- Media Unlock Results -->
          <el-card
            shadow="never"
            class="section-card"
            v-if="report.media?.data"
          >
            <template #header>
              <span class="section-title">流媒体解锁测试</span>
            </template>

            <div v-if="report.media.data.ipv4?.length" class="subsection">
              <h3>IPv4 解锁情况</h3>
              <div
                v-for="(block, index) in report.media.data.ipv4"
                :key="'ipv4-' + index"
              >
                <h4>{{ block.region }}</h4>
                <div class="tag-group">
                  <el-tag
                    v-for="(item, idx) in block.results"
                    :key="idx"
                    :type="getStatusColor(item.unlock)"
                    style="margin: 4px"
                  >
                    {{ item.media }}: {{ item.unlock }}
                  </el-tag>
                </div>
              </div>
            </div>

            <div v-if="report.media.data.ipv6?.length" class="subsection">
              <h3>IPv6 解锁情况</h3>
              <div
                v-for="(block, index) in report.media.data.ipv6"
                :key="'ipv6-' + index"
              >
                <h4>{{ block.region }}</h4>
                <div class="tag-group">
                  <el-tag
                    v-for="(item, idx) in block.results"
                    :key="idx"
                    :type="getStatusColor(item.unlock)"
                    style="margin: 4px"
                  >
                    {{ item.media }}: {{ item.unlock }}
                  </el-tag>
                </div>
              </div>
            </div>
          </el-card>

          <!-- IP Quality Results -->
          <el-card
            shadow="never"
            class="section-card"
            v-if="report.ipquality?.data"
          >
            <template #header>
              <span class="section-title">IP 质量检测</span>
            </template>

            <!-- Head Info -->
            <div v-if="report.ipquality.data.Head?.[0]" class="subsection">
              <h3>检测信息</h3>
              <el-descriptions :column="2" border>
                <el-descriptions-item label="IP">{{
                  report.ipquality.data.Head[0].IP
                }}</el-descriptions-item>
                <el-descriptions-item label="检测时间">{{
                  report.ipquality.data.Head[0].Time
                }}</el-descriptions-item>
                <el-descriptions-item label="版本">{{
                  report.ipquality.data.Head[0].Version
                }}</el-descriptions-item>
                <el-descriptions-item label="GitHub">
                  <el-link
                    :href="report.ipquality.data.Head[0].GitHub"
                    target="_blank"
                  >
                    {{ report.ipquality.data.Head[0].GitHub }}
                  </el-link>
                </el-descriptions-item>
              </el-descriptions>
            </div>

            <!-- Geographic Info -->
            <div v-if="report.ipquality.data.Info?.[0]" class="subsection">
              <h3>地理位置信息</h3>
              <el-descriptions :column="2" border>
                <el-descriptions-item label="ASN">{{
                  report.ipquality.data.Info[0].ASN
                }}</el-descriptions-item>
                <el-descriptions-item label="组织">{{
                  report.ipquality.data.Info[0].Organization
                }}</el-descriptions-item>
                <el-descriptions-item label="大洲">
                  {{ report.ipquality.data.Info[0].Continent?.Name }} ({{
                    report.ipquality.data.Info[0].Continent?.Code
                  }})
                </el-descriptions-item>
                <el-descriptions-item label="国家/地区">
                  {{ report.ipquality.data.Info[0].Region?.Name }} ({{
                    report.ipquality.data.Info[0].Region?.Code
                  }})
                </el-descriptions-item>
                <el-descriptions-item label="城市">
                  {{ report.ipquality.data.Info[0].City?.Name }}
                </el-descriptions-item>
                <el-descriptions-item label="时区">{{
                  report.ipquality.data.Info[0].TimeZone
                }}</el-descriptions-item>
                <el-descriptions-item label="经纬度" :span="2">
                  {{ report.ipquality.data.Info[0].Latitude }},
                  {{ report.ipquality.data.Info[0].Longitude }}
                </el-descriptions-item>
                <el-descriptions-item
                  label="地图链接"
                  :span="2"
                  v-if="report.ipquality.data.Info[0].Map"
                >
                  <el-link
                    :href="report.ipquality.data.Info[0].Map"
                    target="_blank"
                    >查看地图</el-link
                  >
                </el-descriptions-item>
              </el-descriptions>
            </div>

            <!-- Risk Scores -->
            <div v-if="report.ipquality.data.Score?.[0]" class="subsection">
              <h3>风险评分</h3>
              <el-descriptions :column="3" border>
                <el-descriptions-item
                  v-for="(value, key) in report.ipquality.data.Score[0]"
                  :key="key"
                  :label="key"
                >
                  {{ value }}
                </el-descriptions-item>
              </el-descriptions>
            </div>

            <!-- Risk Factors -->
            <div v-if="report.ipquality.data.Factor?.[0]" class="subsection">
              <h3>风险因素</h3>
              <div class="tag-group">
                <el-tag
                  style="margin: 4px"
                  v-if="report.ipquality.data.Factor[0].VPN !== undefined"
                >
                  VPN: {{ report.ipquality.data.Factor[0].VPN ? "是" : "否" }}
                </el-tag>
                <el-tag
                  style="margin: 4px"
                  v-if="report.ipquality.data.Factor[0].Proxy !== undefined"
                >
                  代理:
                  {{ report.ipquality.data.Factor[0].Proxy ? "是" : "否" }}
                </el-tag>
                <el-tag
                  style="margin: 4px"
                  v-if="report.ipquality.data.Factor[0].Tor !== undefined"
                >
                  Tor: {{ report.ipquality.data.Factor[0].Tor ? "是" : "否" }}
                </el-tag>
                <el-tag
                  style="margin: 4px"
                  v-if="report.ipquality.data.Factor[0].Server !== undefined"
                >
                  服务器:
                  {{ report.ipquality.data.Factor[0].Server ? "是" : "否" }}
                </el-tag>
                <el-tag
                  style="margin: 4px"
                  v-if="report.ipquality.data.Factor[0].Abuser !== undefined"
                >
                  滥用者:
                  {{ report.ipquality.data.Factor[0].Abuser ? "是" : "否" }}
                </el-tag>
                <el-tag
                  style="margin: 4px"
                  v-if="report.ipquality.data.Factor[0].Robot !== undefined"
                >
                  机器人:
                  {{ report.ipquality.data.Factor[0].Robot ? "是" : "否" }}
                </el-tag>
              </div>
            </div>

            <!-- Mail Test Results -->
            <div v-if="report.ipquality.data.Mail?.[0]" class="subsection">
              <h3>邮件服务测试</h3>
              <div class="tag-group">
                <el-tag
                  v-for="(value, key) in report.ipquality.data.Mail[0]"
                  :key="key"
                  :type="
                    key === 'DNSBlacklist'
                      ? 'info'
                      : value
                      ? 'success'
                      : 'danger'
                  "
                  style="margin: 4px"
                >
                  <template v-if="key === 'DNSBlacklist'">
                    DNS黑名单: {{ value.Blacklisted }}/{{ value.Total }} 黑名单
                  </template>
                  <template v-else>
                    {{ key }}: {{ value ? "✓" : "✗" }}
                  </template>
                </el-tag>
              </div>
            </div>

            <!-- Media Unlock -->
            <div v-if="report.ipquality.data.Media?.[0]" class="subsection">
              <h3>流媒体服务</h3>
              <el-row :gutter="16">
                <el-col
                  :span="8"
                  v-for="(value, key) in report.ipquality.data.Media[0]"
                  :key="key"
                >
                  <el-card shadow="hover" class="media-card">
                    <h4>{{ key }}</h4>
                    <el-descriptions :column="1" size="small">
                      <el-descriptions-item label="状态">
                        <el-tag :type="getStatusColor(value.status)">{{
                          value.status
                        }}</el-tag>
                      </el-descriptions-item>
                      <el-descriptions-item label="地区">{{
                        value.region
                      }}</el-descriptions-item>
                      <el-descriptions-item label="类型">{{
                        value.type
                      }}</el-descriptions-item>
                    </el-descriptions>
                  </el-card>
                </el-col>
              </el-row>
            </div>

            <!-- Type Info -->
            <div v-if="report.ipquality.data.Type?.[0]" class="subsection">
              <h3>IP 类型</h3>
              <el-row :gutter="16">
                <el-col :span="12" v-if="report.ipquality.data.Type[0].Company">
                  <h4>公司类型</h4>
                  <el-descriptions :column="1" border size="small">
                    <el-descriptions-item
                      v-for="(value, key) in report.ipquality.data.Type[0]
                        .Company"
                      :key="key"
                      :label="key"
                    >
                      {{ value }}
                    </el-descriptions-item>
                  </el-descriptions>
                </el-col>
                <el-col :span="12" v-if="report.ipquality.data.Type[0].Usage">
                  <h4>使用类型</h4>
                  <el-descriptions :column="1" border size="small">
                    <el-descriptions-item
                      v-for="(value, key) in report.ipquality.data.Type[0]
                        .Usage"
                      :key="key"
                      :label="key"
                    >
                      {{ value }}
                    </el-descriptions-item>
                  </el-descriptions>
                </el-col>
              </el-row>
            </div>
          </el-card>

          <!-- Best Trace Results -->
          <el-card
            shadow="never"
            class="section-card"
            v-if="report.besttrace?.data?.length"
          >
            <template #header>
              <span class="section-title">BestTrace 路由追踪</span>
            </template>

            <div
              v-for="(trace, index) in report.besttrace.data"
              :key="index"
              class="subsection"
            >
              <h3>{{ trace.region }}</h3>
              <pre class="code-block">{{ trace.route }}</pre>
            </div>
          </el-card>

          <!-- ITDog Results -->
          <el-card
            shadow="never"
            class="section-card"
            v-if="report.itdog?.data"
          >
            <template #header>
              <span class="section-title">ITDog 测试结果</span>
            </template>

            <div v-if="report.itdog.data.ping" class="subsection">
              <h3>Ping 测试结果</h3>
              <img
                :src="'data:image/png;base64,' + report.itdog.data.ping"
                alt="Ping Result"
                style="max-width: 100%"
              />
            </div>

            <div v-if="report.itdog.data.route?.length" class="subsection">
              <h3>路由追踪结果</h3>
              <div
                v-for="(routeImg, index) in report.itdog.data.route"
                :key="index"
                style="margin-bottom: 16px"
              >
                <h4>路由 {{ index + 1 }}</h4>
                <img
                  :src="'data:image/png;base64,' + routeImg"
                  alt="Route Result"
                  style="max-width: 100%"
                />
              </div>
            </div>
          </el-card>

          <!-- Disk Test Results -->
          <el-card shadow="never" class="section-card" v-if="report.disk?.data">
            <template #header>
              <div class="card-header">
                <span class="section-title">磁盘测试</span>
                <el-tag v-if="report.disk.data.time" size="small">{{
                  report.disk.data.time
                }}</el-tag>
              </div>
            </template>

            <el-table :data="report.disk.data.data" border stripe>
              <el-table-column
                v-for="(col, index) in report.disk.data.data[0] || []"
                :key="index"
                :label="'列 ' + (index + 1)"
                min-width="100"
              >
                <template #default="{ row }">
                  {{ row[index] }}
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </div>

        <el-empty v-else description="未找到报告数据" />
      </template>
    </el-skeleton>
  </div>
</template>

<style scoped>
#report-root {
  width: 100%;
  max-width: 1400px;
  margin: 0 auto;
  padding: 24px;
  box-sizing: border-box;
}

#report-header {
  margin-bottom: 32px;
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
  font-size: 20px;
  font-weight: 600;
  color: #303133;
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
  margin-bottom: 16px;
}

.media-card h4 {
  margin: 0 0 12px 0;
  font-size: 16px;
  color: #303133;
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
</style>
