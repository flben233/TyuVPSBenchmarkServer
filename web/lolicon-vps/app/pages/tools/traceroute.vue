<script setup>
const { traceroute } = useTool();

const target = ref("");
const mode = ref("icmp");
const port = ref(null);
const loading = ref(false);
const result = ref(null);
const errorMessage = ref("");

const handleQuery = async () => {
  if (!target.value) {
    errorMessage.value = "请输入目标IP或域名";
    return;
  }

  loading.value = true;
  errorMessage.value = "";
  result.value = null;

  const response = await traceroute(target.value, mode.value, port.value);

  if (response.code === 0) {
    result.value = response.data;
  } else {
    errorMessage.value = response.message || "路由追踪失败";
  }

  loading.value = false;
};

const handleReset = () => {
  target.value = "";
  mode.value = "icmp";
  port.value = null;
  result.value = null;
  errorMessage.value = "";
};
</script>

<template>
  <div id="tool-root">
    <el-row>
      <el-col :span="24" id="tool-title"> 路由追踪 </el-col>
      <el-col :span="17" :xs="24">
        <div class="tool-form-card">
          <el-form label-width="120px" label-position="left">
            <el-form-item label="追踪目标">
              <el-input
                v-model="target"
                placeholder="输入IP地址或域名"
                clearable
                @keyup.enter="handleQuery"
              />
            </el-form-item>
            <el-form-item label="模式">
              <el-select v-model="mode" placeholder="选择模式">
                <el-option label="ICMP" value="icmp" />
                <el-option label="TCP" value="tcp" />
              </el-select>
            </el-form-item>
            <el-form-item label="端口" v-if="mode === 'tcp'">
              <el-input-number
                v-model="port"
                :min="1"
                :max="65535"
                placeholder="TCP模式端口号"
              />
            </el-form-item>
            <el-form-item label="数据源">
              <div style="color: var(--el-text-color-regular); padding: 0 8px;">NextTrace</div>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleQuery" :loading="loading">
                追踪
              </el-button>
              <el-button @click="handleReset">重置</el-button>
            </el-form-item>
          </el-form>
        </div>

        <div v-if="errorMessage" class="error-message">
          <el-alert type="error" :closable="false">
            {{ errorMessage }}
          </el-alert>
        </div>

        <div v-if="result" class="result-section">
          <div class="result-header">追踪结果</div>

          <!-- Hops Table -->
            <el-table :data="result.Hops" border style="width: 100%">
              <el-table-column prop="TTL" label="跳数" width="80" align="center">
                <template #default="{ row }">
                  {{ row[0]?.TTL || '-' }}
                </template>
              </el-table-column>
              <el-table-column label="IP地址" min-width="150">
                <template #default="{ row }">
                  <span v-if="row[0]?.Success && row[0]?.Address">{{ row[0].Address.IP }}</span>
                  <span v-else style="color: #909399;">*</span>
                </template>
              </el-table-column>
              <el-table-column label="主机名" min-width="150">
                <template #default="{ row }">
                  <span v-if="row[0]?.Hostname">{{ row[0].Hostname }}</span>
                  <span v-else style="color: #909399;">-</span>
                </template>
              </el-table-column>
              <el-table-column label="RTT" min-width="120">
                <template #default="{ row }">
                  <span v-if="row[0]?.Success && row[0]?.RTT">{{ (row[0].RTT / 1000000).toFixed(2) }} ms</span>
                  <span v-else style="color: #909399;">*</span>
                </template>
              </el-table-column>
              <el-table-column label="地理位置" min-width="200">
                <template #default="{ row }">
                    <span v-if="row[0]?.Success && row[0]?.Geo">
                      <span v-if="row[0].Geo.country || row[0].Geo.prov || row[0].Geo.city">
                        {{ row[0].Geo.country }} {{ row[0].Geo.prov }} {{ row[0].Geo.city }}
                      </span>
                      <span v-else-if="row[0].Geo.whois">{{ row[0].Geo.whois }}</span>
                      <span v-else style="color: #909399;">-</span>
                    </span>
                    <span v-else style="color: #909399;">-</span>
                </template>
              </el-table-column>
              <el-table-column label="AS" min-width="150">
                <template #default="{ row }">
                  <span v-if="row[0]?.Success && row[0]?.Geo?.asnumber">{{ row[0].Geo.asnumber }}</span>
                  <span v-else style="color: #909399;">-</span>
                </template>
              </el-table-column>
            </el-table>

          <!-- Trace Map -->
          <div v-if="result.TraceMapUrl.includes('https://assets.nxtrace.org/tracemap/')" class="map-section">
            <div class="result-header">路由地图</div>
            <iframe
              :src="result.TraceMapUrl"
              class="trace-map-iframe"
              frameborder="0"
            ></iframe>
          </div>
        </div>
      </el-col>

      <el-col :span="6" :xs="0" :offset="1">
        <Profile>
          <div>
            <div style="font-weight: 600; color: #303133">关于路由追踪</div>
            <div class="hint-item">追踪到目标的网络路径</div>
            <div class="hint-item">支持ICMP和TCP模式</div>
            <div class="hint-item">显示路由中的每一跳</div>
            <div class="hint-item">帮助诊断网络问题</div>
          </div>
        </Profile>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
#tool-root {
  width: 100%;
  padding: 16px;
  box-sizing: border-box;
  overflow-x: hidden;
}

#tool-title {
  font-size: 28px;
  font-weight: 300;
  margin-bottom: 16px;
}

.tool-form-card {
  margin-bottom: 16px;
}

.error-message {
  margin-bottom: 16px;
}

.result-section {
  margin-top: 16px;
}

.result-header {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
}

.result-content {
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: "Courier New", monospace;
  font-size: 13px;
  color: #303133;
}

.hint-item {
  margin-top: 8px;
  font-size: 14px;
  color: #606266;
}

.map-section {
  margin-top: 24px;
}

.trace-map-iframe {
  width: 100%;
  height: 600px;
  border: none;
}
</style>
