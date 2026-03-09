<script setup>
import { ArrowLeft, Plus, RefreshRight, Setting } from "@element-plus/icons-vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { getEmptyInspectorSettings } from "~/utils/inspector";

useHead({
  title: "Lolicon Monitor",
  meta: [
    { name: "description", content: "Inspector 页面用于管理服务器、查看 Ping 延迟与流量统计，并配置通知和背景图。" },
    { name: "keywords", content: "Inspector, VPS 监控, Ping 监控, 流量统计, 服务器管理" },
  ],
});

const GITHUB_OAUTH_URL = "https://github.com/login/oauth/authorize?client_id=Ov23limxDDoGO9of9P4m";
const REFRESH_INTERVAL_MS = 60 * 1000;

const { token } = useAuth();
const {
  loadDashboard,
  createHost,
  updateHost,
  deleteHost,
  getSettings,
  updateSettings,
} = useInspector();

const loading = ref(false);
const refreshing = ref(false);
const submittingHost = ref(false);
const submittingSettings = ref(false);
const errorMessage = ref("");
const hosts = ref([]);
const settings = ref(getEmptyInspectorSettings());
const addDialogVisible = ref(false);
const editDialogVisible = ref(false);
const settingsDialogVisible = ref(false);
const selectedHost = ref(null);
const refreshTimer = ref(null);

const isLoggedIn = computed(() => Boolean(token.value));

const pageBackgroundStyle = computed(() => {
  return {
    backgroundImage: `url('${settings.value.bgUrl.replace(/'/g, "\\'")}')`,
    backgroundSize: settings.value.bgUrl ? "cover" : "auto",
    backgroundPosition: "center",
    backgroundAttachment: "fixed",
  };
});

function startGithubLogin() {
  window.location.href = GITHUB_OAUTH_URL;
}

async function loadInspectorData({ silent = false } = {}) {
  if (!token.value) {
    return;
  }

  if (!silent) {
    loading.value = true;
    errorMessage.value = "";
  } else {
    refreshing.value = true;
  }

  const [dashboardResult, settingsResult] = await Promise.allSettled([
    loadDashboard(token.value),
    getSettings(token.value),
  ]);

  const errors = [];

  if (dashboardResult.status === "fulfilled") {
    if (dashboardResult.value.success) {
      hosts.value = dashboardResult.value.data || [];
    } else {
      errors.push(dashboardResult.value.message);
    }
  } else {
    errors.push(dashboardResult.reason?.message || "加载服务器数据失败");
  }

  if (settingsResult.status === "fulfilled") {
    if (settingsResult.value.success) {
      settings.value = settingsResult.value.data || getEmptyInspectorSettings();
    } else {
      errors.push(settingsResult.value.message);
    }
  } else {
    errors.push(settingsResult.reason?.message || "加载设置失败");
  }

  errorMessage.value = errors.join("；");
  if (errorMessage.value && !silent) {
    ElMessage.error(errorMessage.value);
  }

  loading.value = false;
  refreshing.value = false;
}

function startAutoRefresh() {
  if (refreshTimer.value || !process.client) {
    return;
  }

  refreshTimer.value = setInterval(() => {
    loadInspectorData({ silent: true });
  }, REFRESH_INTERVAL_MS);
}

function stopAutoRefresh() {
  if (!refreshTimer.value) {
    return;
  }

  clearInterval(refreshTimer.value);
  refreshTimer.value = null;
}

function openEditDialog(host) {
  selectedHost.value = host;
  editDialogVisible.value = true;
}

async function handleCreateHost(payload) {
  submittingHost.value = true;
  const result = await createHost(token.value, payload);
  submittingHost.value = false;

  if (!result.success) {
    ElMessage.error(result.message || "创建服务器失败");
    return;
  }

  addDialogVisible.value = false;
  ElMessage.success("服务器已创建");
  await loadInspectorData({ silent: true });
}

async function handleUpdateHost(payload) {
  if (!selectedHost.value) {
    return;
  }

  submittingHost.value = true;
  const result = await updateHost(token.value, selectedHost.value.id, {
    name: payload.name,
    target: payload.target,
    tags: payload.tags,
  });
  submittingHost.value = false;

  if (!result.success) {
    ElMessage.error(result.message || "更新服务器失败");
    return;
  }

  editDialogVisible.value = false;
  selectedHost.value = null;
  ElMessage.success("服务器信息已更新");
  await loadInspectorData({ silent: true });
}

async function handleDeleteHost(host) {
  try {
    await ElMessageBox.confirm(`确定删除服务器「${host.name}」吗？`, "删除确认", {
      type: "warning",
      confirmButtonText: "删除",
      cancelButtonText: "取消",
    });
  } catch {
    return;
  }

  const result = await deleteHost(token.value, host.id);
  if (!result.success) {
    ElMessage.error(result.message || "删除服务器失败");
    return;
  }

  ElMessage.success("服务器已删除");
  await loadInspectorData({ silent: true });
}

async function handleSaveSettings(payload) {
  submittingSettings.value = true;
  const result = await updateSettings(token.value, payload);
  submittingSettings.value = false;

  if (!result.success) {
    ElMessage.error(result.message || "保存设置失败");
    return;
  }

  settings.value = result.data || getEmptyInspectorSettings();
  settingsDialogVisible.value = false;
  ElMessage.success("设置已保存");
}

watch(
  () => token.value,
  async (currentToken) => {
    stopAutoRefresh();

    if (!currentToken) {
      hosts.value = [];
      settings.value = getEmptyInspectorSettings();
      loading.value = false;
      return;
    }

    await loadInspectorData();
    startAutoRefresh();
  },
  { immediate: true },
);

onUnmounted(() => {
  stopAutoRefresh();
});
</script>

<template>
  <div class="inspector-page" :style="pageBackgroundStyle">
    <div class="inspector-overlay">
      <div class="inspector-header">
        <div>
          <h1 class="page-title">Lolicon Monitor</h1>
        </div>
        <div class="toolbar">
          <el-button class="tool-btn" link :icon="ArrowLeft" @click="useRouter().back()"/>
          <el-button class="tool-btn" link :icon="RefreshRight" :loading="refreshing" @click="loadInspectorData({ silent: true })"/>
          <el-button class="tool-btn" link :icon="Setting" @click="settingsDialogVisible = true" :disabled="!isLoggedIn"/>
          <el-button class="tool-btn" link type="primary" :icon="Plus" @click="addDialogVisible = true" :disabled="!isLoggedIn"/>
        </div>
      </div>

      <el-alert
        v-if="errorMessage"
        class="page-alert"
        type="error"
        show-icon
        :closable="false"
        :title="errorMessage"
      />

      <div v-if="!isLoggedIn" class="login-state">
        <el-empty description="请先登录后访问">
          <el-button type="primary" @click="startGithubLogin">使用 GitHub 登录</el-button>
        </el-empty>
      </div>

      <template v-else>
        <div v-if="loading" class="skeleton-grid">
          <el-card v-for="index in 3" :key="index" shadow="never" class="skeleton-card">
            <el-skeleton :rows="8" animated />
          </el-card>
        </div>

        <el-empty
          v-else-if="hosts.length === 0"
          description=" "
        />

        <div v-else class="host-grid">
          <InspectorHostCard
            v-for="host in hosts"
            :key="host.id"
            :host="host"
            @edit="openEditDialog"
            @delete="handleDeleteHost"
          />
        </div>
      </template>
    </div>

    <InspectorHostDialog
      v-model="addDialogVisible"
      mode="create"
      :submitting="submittingHost"
      @submit="handleCreateHost"
    />

    <InspectorHostDialog
      v-model="editDialogVisible"
      mode="edit"
      :host="selectedHost"
      :submitting="submittingHost"
      @submit="handleUpdateHost"
    />

    <InspectorSettingsDialog
      v-model="settingsDialogVisible"
      :settings="settings"
      :submitting="submittingSettings"
      @save="handleSaveSettings"
    />
  </div>
</template>

<style scoped>
.inspector-page {
  flex: 1;
  min-width: 0;
  min-height: 100vh;
  max-height: 100vh;
  overflow-y: auto;
}

.inspector-overlay {
  min-height: 100vh;
  padding: 16px;
  box-sizing: border-box;
}

.inspector-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 20px;
}

.page-title {
  font-size: 28px;
  font-weight: 300;
  margin-bottom: 16px;
  margin-top: -3px;
  font-family: Noto Sans SC, sans-serif;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.tool-btn {
  font-size: 20px !important;
}

.page-alert {
  margin-bottom: 20px;
}

.login-state {
  min-height: 50vh;
  display: flex;
  align-items: center;
  justify-content: center;
}

.host-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.skeleton-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}

.skeleton-card {
  border-radius: 20px;
}

@media screen and (max-width: 768px) {
  .inspector-overlay {
    padding: 16px;
  }

  .inspector-header {
    flex-direction: column;
  }

  .toolbar {
    width: 100%;
    justify-content: stretch;
  }

  .toolbar :deep(.el-button) {
    flex: 1;
  }
}
</style>
