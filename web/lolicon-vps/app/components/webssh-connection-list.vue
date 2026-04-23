<script setup>
import { Plus, Edit, Delete, Link, Upload, Download, Key } from "@element-plus/icons-vue";
import {
  getConnections,
  saveConnection,
  saveConnections,
  deleteConnection,
  getPassword,
  clearPassword,
} from "~/utils/webssh-storage";
import { getLLMSettings, saveLLMSettings } from "~/utils/webssh-llm-settings";
import { encryptData, decryptData } from "~/utils/webssh-crypto";
import { error, success, warn } from "~/utils/message.js";

const { uploadEncryptedData, downloadEncryptedData, resetCloudData } = useWebsshCloud();

const props = defineProps({
  activeId: {
    type: String,
    default: null,
  },
  status: String,
});

const emit = defineEmits(["select", "connect"]);

const connections = ref([]);
const showDialog = ref(false);
const editingConnection = ref(null);
const showKeyDialog = ref(false);
const keyDialogMode = ref("set");
const uploading = ref(false);
const downloading = ref(false);
const showDiffDialog = ref(false);
const diffItems = ref([]);
const localOnlyItems = ref([]);
const pendingCloudConnections = ref([]);
const pendingNewConns = ref([]);

function isConnectionEqual(a, b) {
  const keys = ["name", "host", "port", "username", "authType", "password", "privateKey"];
  return keys.every((k) => a[k] === b[k]);
}

const FIELD_LABELS = {
  name: "名称",
  host: "主机",
  port: "端口",
  username: "用户名",
  authType: "认证方式",
  password: "密码",
  privateKey: "私钥",
};

function maskValue(key, val) {
  if ((key === "password" || key === "privateKey") && val) return "******";
  if ((key === "password" || key === "privateKey") && !val) return "无";
  return val ?? "";
}

function getConnectionDiffs(local, cloud) {
  const keys = ["name", "host", "port", "username", "authType", "password", "privateKey"];
  const diffs = [];
  for (const k of keys) {
    if (local[k] !== cloud[k]) {
      diffs.push({
        label: FIELD_LABELS[k],
        localValue: maskValue(k, local[k]),
        cloudValue: maskValue(k, cloud[k]),
      });
    }
  }
  return diffs;
}

function loadConnections() {
  connections.value = getConnections();
}

function handleSelect(conn) {
  if (props.status !== "disconnected" && props.status !== "error") {
    error("当前有连接正在进行中，请先断开后再选择其他连接");
    return
  }
  emit("select", conn);
}

function handleDoubleClick(conn) {
  if (props.status !== "disconnected" && props.status !== "error") {
    error("当前有连接正在进行中，请先断开后再选择其他连接");
    return
  }
  emit("connect", conn);
}

function handleAdd() {
  editingConnection.value = null;
  showDialog.value = true;
}

function handleEdit(conn) {
  editingConnection.value = { ...conn };
  showDialog.value = true;
}

function handleDelete(conn) {
  deleteConnection(conn.id);
  loadConnections();
}

function handleSave(conn) {
  saveConnection(conn);
  loadConnections();
}

function openKeyDialog(mode) {
  keyDialogMode.value = mode;
  showKeyDialog.value = true;
}

async function handleUpload() {
  if (!getPassword()) {
    warn("请先设置密钥");
    openKeyDialog("set");
    return;
  }

  const conns = getConnections();
  if (conns.length === 0) {
    warn("暂无连接可上传");
    return;
  }

  uploading.value = true;
  try {
    const plainText = JSON.stringify({
      connections: conns,
      llmSettings: getLLMSettings(),
    });
    const encrypted = await encryptData(plainText, getPassword());
    await uploadEncryptedData(encrypted);
    success("上传成功");
  } catch (e) {
    error("上传失败: " + (e.message || e.data?.message || "未知错误"));
  } finally {
    uploading.value = false;
  }
}

async function handleDownload() {
  if (!getPassword()) {
    warn("请先设置密钥");
    openKeyDialog("set");
    return;
  }

  downloading.value = true;
  try {
    const resp = await downloadEncryptedData();
    if (!resp.data || !resp.data.encrypted_data) {
      warn("云端暂无数据");
      return;
    }

    const decrypted = await decryptData(resp.data.encrypted_data, getPassword());
    const cloudPayload = JSON.parse(decrypted);
    const cloudConnections = Array.isArray(cloudPayload)
      ? cloudPayload
      : Array.isArray(cloudPayload?.connections)
        ? cloudPayload.connections
        : [];
    if (cloudPayload && !Array.isArray(cloudPayload) && cloudPayload.llmSettings) {
      saveLLMSettings(cloudPayload.llmSettings);
    }

    if (!Array.isArray(cloudConnections) || cloudConnections.length === 0) {
      warn("云端暂无数据");
      return;
    }

    const localConnections = getConnections();
    const localMap = new Map(localConnections.map((c) => [c.id, c]));
    const localNames = new Set(localConnections.map((c) => c.name));
    const cloudIds = new Set(cloudConnections.map((c) => c.id));

    const newConns = [];
    const conflicts = [];
    const localOnly = localConnections.filter((c) => !cloudIds.has(c.id));

    for (const conn of cloudConnections) {
      const local = localMap.get(conn.id);
      if (local) {
        if (!isConnectionEqual(local, conn)) {
          conflicts.push({ local, cloud: conn, diffs: getConnectionDiffs(local, conn) });
        }
      } else if (!localNames.has(conn.name)) {
        newConns.push(conn);
      }
    }

    for (const conn of newConns) {
      saveConnection(conn);
    }

    if (conflicts.length === 0 && localOnly.length === 0) {
      let msg = `下载完成`;
      if (newConns.length > 0) msg += `，新增 ${newConns.length} 个连接`;
      success(msg);
      loadConnections();
      return;
    }

    diffItems.value = conflicts;
    localOnlyItems.value = localOnly;
    pendingCloudConnections.value = cloudConnections;
    pendingNewConns.value = newConns;
    showDiffDialog.value = true;
    loadConnections();
  } catch (e) {
    if (e.message?.includes("OperationError") || e.name === "OperationError") {
      error("解密失败，密钥可能不正确");
    } else {
      error("下载失败: " + (e.message || e.data?.message || "未知错误"));
    }
  } finally {
    downloading.value = false;
  }
}

function handleMerge() {
  const cloudMap = new Map(pendingCloudConnections.value.map((c) => [c.id, c]));
  const merged = [...pendingCloudConnections.value];
  for (const conn of localOnlyItems.value) {
    if (!cloudMap.has(conn.id)) {
      merged.push(conn);
    }
  }
  saveConnections(merged);
  const conflictCount = diffItems.value.length;
  const localOnlyCount = localOnlyItems.value.length;
  showDiffDialog.value = false;
  let msg = "合并完成";
  if (conflictCount > 0) msg += `，更新 ${conflictCount} 个差异连接`;
  if (localOnlyCount > 0) msg += `，保留 ${localOnlyCount} 个本地独有连接`;
  success(msg);
  loadConnections();
}

function handleOverwrite() {
  saveConnections(pendingCloudConnections.value);
  showDiffDialog.value = false;
  success("已使用云端数据覆盖本地");
  loadConnections();
}

function handleSkipConflict() {
  showDiffDialog.value = false;
  success("已跳过差异连接，保留本地版本");
  loadConnections();
}

async function handleKeySaved() {
  success("密钥已设置");
}

async function handleKeyReset() {
  try {
    await resetCloudData();
    clearPassword();
    success("密钥已重置，云端数据已删除");
    showKeyDialog.value = false;
  } catch (e) {
    error("重置失败: " + (e.message || "未知错误"));
  }
}

onMounted(() => {
  loadConnections();
});
</script>

<template>
  <div class="connection-list">
    <div class="list-header">
      <span class="list-title">连接列表</span>
      <el-button-group class="header-actions">
        <el-button
          type="primary"
          size="small"
          :icon="Upload"
          :loading="uploading"
          @click="handleUpload"
          title="上传到云端"
        />
        <el-button
          type="primary"
          size="small"
          :icon="Download"
          :loading="downloading"
          @click="handleDownload"
          title="从云端下载"
        />
        <el-button
          size="small"
          type="primary"
          :icon="Key"
          @click="openKeyDialog(getPassword() ? 'change' : 'set')"
          title="密钥管理"
        />
        <el-button type="primary" size="small" :icon="Plus" @click="handleAdd" title="新建连接"/>
      </el-button-group>
    </div>

    <div class="list-items">
      <div
        v-for="conn in connections"
        :key="conn.id"
        class="connection-item"
        :class="{ active: conn.id === activeId }"
        @click="handleSelect(conn)"
        @dblclick="handleDoubleClick(conn)"
      >
        <div class="item-info">
          <div class="item-name">{{ conn.name }}</div>
          <div class="item-detail">
            {{ conn.username }}@{{ conn.host }}:{{ conn.port }}
          </div>
        </div>
        <div class="item-actions">
          <el-button
            size="small"
            :icon="Link"
            circle
            @click.stop="emit('connect', conn)"
            :disabled="status === 'connecting'"
          />
          <el-button
            size="small"
            :icon="Edit"
            circle
            @click.stop="handleEdit(conn)"
          />
          <el-popconfirm
            title="确认删除此连接？"
            @confirm="handleDelete(conn)"
          >
            <template #reference>
              <el-button
                size="small"
                :icon="Delete"
                circle
                type="danger"
                @click.stop
              />
            </template>
          </el-popconfirm>
        </div>
      </div>

      <div v-if="connections.length === 0" class="empty-hint">
        暂无保存的连接，点击上方"+"创建
      </div>
    </div>

    <WebsshConnectionDialog
      v-model="showDialog"
      :edit-connection="editingConnection"
      @save="handleSave"
    />

    <WebsshKeyDialog
      v-model="showKeyDialog"
      :mode="keyDialogMode"
      @saved="handleKeySaved"
      @reset="handleKeyReset"
    />

    <el-dialog
      v-model="showDiffDialog"
      title="云端数据与本地存在差异"
      width="600px"
      :close-on-click-modal="false"
    >
      <div v-if="diffItems.length > 0" class="diff-section">
        <div class="diff-hint">
          以下 {{ diffItems.length }} 个连接在本地和云端的数据不一致：
        </div>
        <div class="diff-list">
          <div v-for="(item, index) in diffItems" :key="index" class="diff-item">
            <div class="diff-item-header">
              {{ item.cloud.name }}（{{ item.cloud.host }}）
            </div>
            <el-table :data="item.diffs" size="small" border>
              <el-table-column label="字段" prop="label" width="100" />
              <el-table-column label="本地" min-width="140">
                <template #default="{ row }">
                  <span class="diff-local">{{ row.localValue }}</span>
                </template>
              </el-table-column>
              <el-table-column label="云端" min-width="140">
                <template #default="{ row }">
                  <span class="diff-cloud">{{ row.cloudValue }}</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </div>

      <div v-if="localOnlyItems.length > 0" class="diff-section">
        <div class="diff-hint">
          以下 {{ localOnlyItems.length }} 个连接仅存在于本地，云端不包含：
        </div>
        <el-table :data="localOnlyItems" size="small" border>
          <el-table-column label="名称" prop="name" min-width="100" />
          <el-table-column label="主机" min-width="130">
            <template #default="{ row }">{{ row.username }}@{{ row.host }}:{{ row.port }}</template>
          </el-table-column>
        </el-table>
      </div>

      <template #footer>
        <el-button @click="handleSkipConflict">跳过</el-button>
        <el-button type="primary" @click="handleMerge">合并（云端 + 本地独有）</el-button>
        <el-button type="danger" @click="handleOverwrite">覆盖本地全部数据</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.connection-list {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-bottom: 1px solid var(--el-border-color-light);
}

.header-actions {
  display: flex;
  align-items: center;
}

.list-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.list-items {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.connection-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
  margin-bottom: 4px;
  box-sizing: border-box;
  border: 1px solid transparent;
}

.connection-item:hover {
  background: var(--el-fill-color-light);
}

.connection-item.active {
  background: var(--el-color-primary-light-9);
  border: 1px solid var(--el-color-primary-light-7);
}

.item-info {
  flex: 1;
  min-width: 0;
}

.item-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-detail {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.empty-hint {
  text-align: center;
  color: #c0c4cc;
  font-size: 13px;
  padding: 24px 12px;
}

.diff-hint {
  margin-bottom: 12px;
  font-size: 14px;
  color: #606266;
}

.diff-section {
  margin-bottom: 20px;
}

.diff-section:last-of-type {
  margin-bottom: 0;
}

.diff-list {
  max-height: 400px;
  overflow-y: auto;
}

.diff-item {
  margin-bottom: 16px;
}

.diff-item:last-child {
  margin-bottom: 0;
}

.diff-item-header {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 8px;
}

.diff-local {
  color: #f56c6c;
}

.diff-cloud {
  color: #67c23a;
}
</style>
