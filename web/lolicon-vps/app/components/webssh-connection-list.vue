<script setup>
import { Plus, Edit, Delete, Link, Upload, Download, Key } from "@element-plus/icons-vue";
import {
  getConnections,
  saveConnection,
  deleteConnection,
  getPassword,
  clearPassword,
} from "~/utils/webssh-storage";
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
    const plainText = JSON.stringify(conns);
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
    const cloudConnections = JSON.parse(decrypted);

    if (!Array.isArray(cloudConnections) || cloudConnections.length === 0) {
      warn("云端暂无数据");
      return;
    }

    const localConnections = getConnections();
    const localIds = new Set(localConnections.map((c) => c.id));
    const localNames = new Set(localConnections.map((c) => c.name));

    let added = 0;
    let skipped = 0;

    for (const conn of cloudConnections) {
      if (localIds.has(conn.id)) {
        skipped++;
        continue;
      }
      if (localNames.has(conn.name)) {
        skipped++;
        continue;
      }
      localConnections.push(conn);
      added++;
    }

    for (const conn of localConnections) {
      saveConnection(conn);
    }

    let msg = `下载完成`;
    if (added > 0) msg += `，新增 ${added} 个连接`;
    if (skipped > 0) msg += `，跳过 ${skipped} 个已存在连接`;
    success(msg);

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
</style>
