<script setup>
import { ArrowLeft, Plus, RefreshRight, Setting } from "@element-plus/icons-vue";
import { ElMessageBox } from "element-plus";
import {
  exceedsInspectorPointLimit,
  getDefaultInspectorQuery,
  getEmptyInspectorSettings,
  INSPECTOR_INTERVAL_OPTIONS,
  formatBytes,
  formatPercent,
  formatTrafficAmount,
} from "~/utils/inspector.js";

useHead({
  title: "Lolicon Monitor",
  meta: [
    { name: "description", content: "Inspector 页面用于管理服务器、查看 Ping 延迟与流量统计，并配置通知和背景图。" },
    { name: "keywords", content: "Inspector, VPS 监控, Ping 监控, 流量统计, 服务器管理" },
  ],
});

const { warn, err, success } = useMessage()
const { clientId } = useAppConfig();
const GITHUB_OAUTH_URL = "https://github.com/login/oauth/authorize?client_id=" + clientId;
const REFRESH_INTERVAL_MS = 60 * 1000;
const DEFAULT_RANGE_MS = 24 * 60 * 60 * 1000;
const MAX_QUERY_POINTS = 288;
const RELATIVE_RANGE_OPTIONS = [
  { label: "过去 1 小时", value: 1 * 60 * 60 * 1000 },
  { label: "过去 6 小时", value: 6 * 60 * 60 * 1000 },
  { label: "过去 24 小时", value: 24 * 60 * 60 * 1000 },
  { label: "过去 7 天", value: 7 * 24 * 60 * 60 * 1000 },
];

const { token } = useAuth();
const router = useRouter();
const {
  loadDashboard,
  createHost,
  updateHost,
  deleteHost,
  getSettings,
  updateSettings,
  getVisitorPage
} = useInspector();

const ownerName = ref("");
const ownerId = useRoute().params.id;
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
const defaultQuery = getDefaultInspectorQuery();
const queryRange = ref([
  new Date(defaultQuery.start / 1_000_000),
  new Date(defaultQuery.end / 1_000_000),
]);
const queryMode = ref("relative");
const queryRelativeRange = ref(DEFAULT_RANGE_MS);
const queryInterval = ref(defaultQuery.interval);
const activeQuery = ref(defaultQuery);
const activeQueryMeta = ref({
  mode: "relative",
  relativeRange: DEFAULT_RANGE_MS,
});
const selectedTags = ref([]);

const { userInfo } = useAuth();

const intervalOptions = INSPECTOR_INTERVAL_OPTIONS;
const relativeRangeOptions = RELATIVE_RANGE_OPTIONS;
const isLoggedIn = computed(() => Boolean(token.value));

const readOnly = computed(() => {
  let oid = useRoute().params.id;
  return oid != null && oid !== "";
});

const pageBackgroundStyle = computed(() => ({
  backgroundImage: `url('${settings.value.bgUrl.replace(/'/g, "\\'")}')`,
  backgroundSize: settings.value.bgUrl ? "cover" : "auto",
  backgroundPosition: "center",
  backgroundAttachment: "fixed",
}));

const availableTags = computed(() => {
  const tags = new Set();
  hosts.value.forEach((host) => {
    (host.tags || []).forEach((tag) => tags.add(tag));
  });
  return Array.from(tags).sort((left, right) => left.localeCompare(right, "zh-CN"));
});

const filteredHosts = computed(() => {
  if (selectedTags.value.length === 0) {
    return hosts.value;
  }

  return hosts.value.filter((host) =>
    selectedTags.value.some((tag) => (host.tags || []).includes(tag)),
  );
});

const dashboardStats = computed(() => {
  const visibleHosts = filteredHosts.value;
  const activeHosts = visibleHosts.filter((host) => Number(host.latestPing) > 0);
  const totalTraffic = visibleHosts.reduce(
    (sum, host) => sum + Number(host.recv || 0) + Number(host.sent || 0),
    0,
  );
  const totalMemoryUsed = activeHosts.reduce((sum, host) => sum + Number(host.memoryUsedBytes || 0), 0);
  const totalMemory = activeHosts.reduce((sum, host) => sum + Number(host.memoryTotalBytes || 0), 0);
  const storagePercent = totalMemory > 0 ? (totalMemoryUsed / totalMemory) * 100 : 0;
  const averageCpu = activeHosts.length > 0
    ? activeHosts.reduce((sum, host) => sum + Number(host.cpuUsagePercent || 0), 0) / activeHosts.length
    : 0;
  const averageMemory = activeHosts.length > 0
    ? activeHosts.reduce((sum, host) => sum + Number(host.memoryUsagePercent || 0), 0) / activeHosts.length
    : 0;

  return {
    hostCountText: `${visibleHosts.length} / ${hosts.value.length}`,
    onlineCountText: `${activeHosts.length} 台`,
    totalTrafficText: formatTrafficAmount(totalTraffic),
    storageUsageText: formatPercent(storagePercent),
    storageUsageDetail: `${formatBytes(totalMemoryUsed)} / ${formatBytes(totalMemory)}`,
    averageCpuText: formatPercent(averageCpu),
    averageMemoryText: formatPercent(averageMemory),
  };
});

const childTheme = computed(() => {

  if (settings.value.bgUrl || settings.value.bgUrl !== "") {
    console.log(settings.value)
    return "theme-bg";
  }
  return "theme-default";
});

function buildQueryFromState({
  mode = queryMode.value,
  range = queryRange.value,
  interval = queryInterval.value,
  relativeRange = queryRelativeRange.value,
} = {}) {
  if (mode === "relative") {
    const durationMs = Number(relativeRange);
    if (!Number.isFinite(durationMs) || durationMs <= 0) {
      return null;
    }

    const end = Date.now();
    const start = end - durationMs;
    return {
      query: {
        start: start * 1_000_000,
        end: end * 1_000_000,
        interval,
      },
      range: [new Date(start), new Date(end)],
      mode,
      relativeRange: durationMs,
    };
  }

  const [startTime, endTime] = Array.isArray(range) ? range : [];
  const start = new Date(startTime).getTime();
  const end = new Date(endTime).getTime();

  if (!Number.isFinite(start) || !Number.isFinite(end) || start >= end) {
    return null;
  }

  return {
    query: {
      start: start * 1_000_000,
      end: end * 1_000_000,
      interval,
    },
    range: [new Date(start), new Date(end)],
    mode,
    relativeRange: Number(relativeRange) || DEFAULT_RANGE_MS,
  };
}

function refreshActiveQueryForRelativeRange() {
  if (activeQueryMeta.value.mode !== "relative") {
    return;
  }

  const refreshed = buildQueryFromState({
    mode: "relative",
    interval: activeQuery.value.interval,
    relativeRange: activeQueryMeta.value.relativeRange,
  });

  if (!refreshed) {
    return;
  }

  activeQuery.value = refreshed.query;
  queryRange.value = refreshed.range;
}

function startGithubLogin() {
  window.location.href = GITHUB_OAUTH_URL;
}

async function loadInspectorData({ silent = false } = {}) {
  if (!silent) {
    loading.value = true;
    errorMessage.value = "";
  } else {
    refreshing.value = true;
  }

  if (!token.value && !readOnly.value) {
    return;
  }

  refreshActiveQueryForRelativeRange();

  let dashboardResult, settingsResult;

  if (readOnly.value) {
    let [result] = await Promise.allSettled([
      getVisitorPage(ownerId, activeQuery.value),
    ]);
    settingsResult = { status: "fulfilled" };
    dashboardResult = { status: "fulfilled" };
    if (result.status === "fulfilled" && result.value.success) {
      settingsResult.value = { success: true, data: { bgUrl: result.value.data.bgUrl } }
      dashboardResult.value = { success: true, data: result.value.data.hosts || [] };
      ownerName.value = result.value.data.ownerName || "";
    }
  } else {
    [dashboardResult, settingsResult] = await Promise.allSettled([
      loadDashboard(activeQuery.value),
      getSettings(),
    ]);
  }

  const errors = [];

  if (dashboardResult.status === "fulfilled") {
    if (dashboardResult.value.success) {
      hosts.value = dashboardResult.value.data || [];
    } else {
      errors.push(dashboardResult.value.message);
      console.log("加载服务器数据失败:", dashboardResult.value.message);
    }
  } else {
    errors.push(dashboardResult.reason?.message || "加载服务器数据失败");
      console.log("加载服务器数据失败:", dashboardResult.reason);
  }

  if (settingsResult.status === "fulfilled") {
    if (settingsResult.value.success) {
      settings.value = settingsResult.value.data || getEmptyInspectorSettings();
    } else {
      errors.push(settingsResult.value.message);
        console.log("加载设置数据失败:", settingsResult.value.message);
    }
  } else {
    errors.push(settingsResult.reason?.message || "加载设置失败");
      console.log("加载设置数据失败:", settingsResult.reason);
  }

  errorMessage.value = errors.join("；");
  if (errorMessage.value && !silent) {
    err(errorMessage.value);
  }

  loading.value = false;
  refreshing.value = false;
}

async function applyQuery({ silent = false } = {}) {
  const nextState = buildQueryFromState();
  if (!nextState) {
    warn("请选择有效的时间范围");
    return;
  }

  const startMs = Number(nextState.query.start) / 1_000_000;
  const endMs = Number(nextState.query.end) / 1_000_000;
  if (endMs - startMs > 90 * 24 * 60 * 60 * 1000) {
    warn("请选择不超过 90 天的时间范围");
    return;
  }
  if (exceedsInspectorPointLimit(startMs, endMs, nextState.query.interval, MAX_QUERY_POINTS)) {
    warn(`当前时间范围和粒度会超过 ${MAX_QUERY_POINTS} 个数据点，请增大时间粒度或缩短时间范围`);
    return;
  }

  activeQuery.value = nextState.query;
  activeQueryMeta.value = {
    mode: nextState.mode,
    relativeRange: nextState.relativeRange,
  };
  queryRange.value = nextState.range;
  await loadInspectorData({ silent });
}

async function resetQuery() {
  const end = new Date();
  queryMode.value = "relative";
  queryRelativeRange.value = DEFAULT_RANGE_MS;
  queryRange.value = [new Date(end.getTime() - DEFAULT_RANGE_MS), end];
  queryInterval.value = defaultQuery.interval;
  selectedTags.value = [];
  await applyQuery();
}

function applyPresetRange(preset) {
  queryMode.value = "relative";
  if (preset === "6h") {
    queryRelativeRange.value = 6 * 60 * 60 * 1000;
    queryInterval.value = "15m";
  } else if (preset === "7d") {
    queryRelativeRange.value = 7 * 24 * 60 * 60 * 1000;
    queryInterval.value = "6h";
  } else {
    queryRelativeRange.value = DEFAULT_RANGE_MS;
    queryInterval.value = defaultQuery.interval;
  }

  const nextState = buildQueryFromState({
    mode: queryMode.value,
    relativeRange: queryRelativeRange.value,
    interval: queryInterval.value,
  });
  if (nextState) {
    queryRange.value = nextState.range;
  }
}

function updateQueryRange(value) {
  queryRange.value = value;
  queryMode.value = "absolute";
}

function updateQueryMode(value) {
  queryMode.value = value;

  if (value === "relative") {
    const nextState = buildQueryFromState({
      mode: value,
      relativeRange: queryRelativeRange.value,
      interval: queryInterval.value,
    });
    if (nextState) {
      queryRange.value = nextState.range;
    }
  }
}

function updateQueryRelativeRange(value) {
  queryRelativeRange.value = value;

  if (queryMode.value !== "relative") {
    return;
  }

  const nextState = buildQueryFromState({
    mode: queryMode.value,
    relativeRange: queryRelativeRange.value,
    interval: queryInterval.value,
  });
  if (nextState) {
    queryRange.value = nextState.range;
  }
}

function updateQueryInterval(value) {
  queryInterval.value = value;
}

function updateSelectedTags(value) {
  selectedTags.value = value;
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
  const result = await createHost(payload);
  submittingHost.value = false;

  if (!result.success) {
    err(result.message || "创建服务器失败");
    return;
  }

  addDialogVisible.value = false;
  success("服务器已创建");
  await loadInspectorData({ silent: true });
}

async function handleUpdateHost(payload) {
  if (!selectedHost.value) {
    return;
  }

  submittingHost.value = true;
  const result = await updateHost(selectedHost.value.id, {
    name: payload.name,
    target: payload.target,
    tags: payload.tags,
    notify: payload.notify,
    notify_tolerance: payload.notify_tolerance,
  });
  submittingHost.value = false;

  if (!result.success) {
    err(result.message || "更新服务器失败");
    return;
  }

  editDialogVisible.value = false;
  selectedHost.value = null;
  success("服务器信息已更新");
  await loadInspectorData({ silent: true });
}

async function handleDeleteHost(host) {
  try {
    ElMessageBox.confirm(`确定删除服务器「${host.name}」吗？`, "删除确认", {
      type: "warning",
      confirmButtonText: "删除",
      cancelButtonText: "取消",
    });
  } catch {
    return;
  }

  const result = await deleteHost(host.id);
  if (!result.success) {
    err(result.message || "删除服务器失败");
    return;
  }

  success("服务器已删除");
  await loadInspectorData({ silent: true });
}

async function handleSaveSettings(payload) {
  submittingSettings.value = true;
  const result = await updateSettings(payload);
  submittingSettings.value = false;

  if (!result.success) {
    err(result.message || "保存设置失败");
    return;
  }

  settings.value = result.data || getEmptyInspectorSettings();
  settingsDialogVisible.value = false;
  success("设置已保存");
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
    ownerId.value = userInfo.value.id;
  },
  { immediate: true },
);

watch(availableTags, (tags) => {
  selectedTags.value = selectedTags.value.filter((tag) => tags.includes(tag));
});

onUnmounted(() => {
  stopAutoRefresh();
});

onMounted(async () => {
  if (readOnly.value) {
    await loadInspectorData();
    startAutoRefresh();
  }
  console.log("Inspector 页面已加载，readOnly =", readOnly.value);
});
</script>

<template>
  <div class="inspector-page" :style="pageBackgroundStyle">
    <div class="inspector-overlay">
      <div class="inspector-header">
        <div>
          <h1 v-if="readOnly" class="page-title">{{ ownerName + " 的探针" }}</h1>
          <h1 v-else class="page-title">Lolicon Monitor</h1>
        </div>
        <div class="toolbar">
          <el-button class="tool-btn" link :icon="ArrowLeft" @click="router.back()"/>
          <el-button class="tool-btn" link :icon="RefreshRight" :loading="refreshing" @click="loadInspectorData({ silent: true })"/>
          <el-button v-if="!readOnly" class="tool-btn" link :icon="Setting" @click="settingsDialogVisible = true" :disabled="!isLoggedIn"/>
          <el-button v-if="!readOnly" class="tool-btn" link type="primary" :icon="Plus" @click="addDialogVisible = true" :disabled="!isLoggedIn"/>
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

      <div v-if="!isLoggedIn && !readOnly" class="login-state">
        <el-empty description="请先登录后访问">
          <el-button type="primary" @click="startGithubLogin">使用 GitHub 登录</el-button>
        </el-empty>
      </div>

      <template v-else>
        <div class="inspector-layout">
          <InspectorSidebar
            :class="childTheme"
            :stats="dashboardStats"
            :range="queryRange"
            :range-mode="queryMode"
            :relative-range="queryRelativeRange"
            :relative-range-options="relativeRangeOptions"
            :interval="queryInterval"
            :interval-options="intervalOptions"
            :tags="availableTags"
            :selected-tags="selectedTags"
            :loading="loading || refreshing"
            @update:range="updateQueryRange"
            @update:range-mode="updateQueryMode"
            @update:relative-range="updateQueryRelativeRange"
            @update:interval="updateQueryInterval"
            @update:selected-tags="updateSelectedTags"
            @apply="applyQuery"
            @reset="resetQuery"
            @preset="applyPresetRange"
          />

          <div class="content-column">
            <el-empty
              v-if="filteredHosts.length === 0"
              description="当前筛选条件下暂无主机"
            />

            <div v-else class="host-grid">
              <InspectorHostCard
                :class="childTheme"
                v-for="host in filteredHosts"
                :key="host.id"
                :host="host"
                :read-only="readOnly"
                @edit="openEditDialog"
                @delete="handleDeleteHost"
              />
            </div>
          </div>
        </div>
      </template>
    </div>

    <InspectorHostDialog
      :class="childTheme"
      v-model="addDialogVisible"
      mode="create"
      :submitting="submittingHost"
      @submit="handleCreateHost"
    />

    <InspectorHostDialog
      :class="childTheme"
      v-model="editDialogVisible"
      mode="edit"
      :host="selectedHost"
      :submitting="submittingHost"
      @submit="handleUpdateHost"
    />

    <InspectorSettingsDialog
      :class="childTheme"
      v-model="settingsDialogVisible"
      :settings="settings"
      :hosts="hosts"
      :owner-id="ownerId"
      :submitting="submittingSettings"
      @save="handleSaveSettings"
    />
  </div>
</template>

<style scoped>
.theme-default {
  --border: 1px solid #e4e7ed;
}

.theme-bg {
  --border: 1px solid rgba(255, 255, 255, 0.62);
}

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

.inspector-layout {
  display: grid;
  grid-template-columns: minmax(280px, 340px) minmax(0, 1fr);
  gap: 20px;
  align-items: start;
}

.content-column {
  min-width: 0;
}

.host-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

@media screen and (max-width: 768px) {
  .inspector-overlay {
    padding: 16px;
  }

  .inspector-header {
    flex-direction: column;
  }

  .inspector-layout {
    grid-template-columns: 1fr;
  }

  .toolbar {
    justify-content: stretch;
  }

  .toolbar :deep(.el-button) {
    flex: 1;
  }
}
</style>
