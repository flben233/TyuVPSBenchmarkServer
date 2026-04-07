<script setup>
import { Plus, Edit, Delete, Link } from "@element-plus/icons-vue";
import {
  getConnections,
  saveConnection,
  deleteConnection,
} from "~/utils/webssh-storage";
import {error} from "~/utils/message.js";

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

onMounted(() => {
  loadConnections();
});
</script>

<template>
  <div class="connection-list">
    <div class="list-header">
      <span class="list-title">连接列表</span>
      <el-button type="primary" size="small" @click="handleAdd" :icon="Plus">
        新建
      </el-button>
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
        暂无保存的连接，点击上方"新建"创建
      </div>
    </div>

    <WebsshConnectionDialog
      v-model="showDialog"
      :edit-connection="editingConnection"
      @save="handleSave"
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
