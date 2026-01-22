<script setup>
const { whois } = useTool();

const target = ref("");
const loading = ref(false);
const result = ref(null);
const errorMessage = ref("");

const handleQuery = async () => {
  if (!target.value) {
    errorMessage.value = "请输入域名或IP地址";
    return;
  }

  loading.value = true;
  errorMessage.value = "";
  result.value = null;

  const response = await whois(target.value);

  if (response.code === 0) {
    result.value = response.data;
  } else {
    errorMessage.value = response.message || "查询WHOIS信息失败";
  }

  loading.value = false;
};

const handleReset = () => {
  target.value = "";
  result.value = null;
  errorMessage.value = "";
};
</script>

<template>
  <div id="tool-root">
    <el-row>
      <el-col :span="24" id="tool-title"> WHOIS查询 </el-col>
      <el-col :span="17" :xs="24">
        <div class="tool-form-card">
          <el-form label-width="120px" label-position="left">
            <el-form-item label="查询目标">
              <el-input
                v-model="target"
                placeholder="输入域名或IP地址"
                clearable
                @keyup.enter="handleQuery"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleQuery" :loading="loading">
                查询
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
          <div class="result-header">WHOIS结果</div>
          <el-card shadow="never" class="result-card">
            <pre class="result-content">{{ result.raw }}</pre>
          </el-card>
        </div>
      </el-col>

      <el-col :span="6" :xs="0" :offset="1">
        <Profile>
          <div>
            <div style="font-weight: 600; color: #303133">关于WHOIS</div>
            <div class="hint-item">查询域名注册信息</div>
            <div class="hint-item">查看注册商详情</div>
            <div class="hint-item">检查域名过期日期</div>
            <div class="hint-item">获取域名服务器信息</div>
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

.result-card {
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
</style>
