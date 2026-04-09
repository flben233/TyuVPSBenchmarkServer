<script setup>
const props = defineProps({
  status: {
    type: String,
    default: "disconnected",
  },
  messages: {
    type: Array,
    default: () => [],
  },
  pendingApprovalByMessageId: {
    type: Object,
    default: () => ({}),
  },
  activeTaskId: {
    type: String,
    default: "",
  },
});

const emit = defineEmits(["submit-task", "submit-message", "submit-approval"]);

const taskInput = ref("");
const messageTaskId = ref("");
const messageInput = ref("");
const isTaskIdDirty = ref(false);

watch(
  () => props.activeTaskId,
  (value) => {
    if (isTaskIdDirty.value) {
      return;
    }
    messageTaskId.value = value || "";
  },
  { immediate: true }
);

const connected = computed(() => props.status === "connected");

function submitTask() {
  const message = taskInput.value.trim();
  if (!message) {
    return;
  }
  emit("submit-task", message);
  taskInput.value = "";
}

function submitMessage() {
  const taskId = messageTaskId.value.trim();
  const message = messageInput.value.trim();
  if (!taskId || !message) {
    return;
  }
  emit("submit-message", { taskId, message });
  isTaskIdDirty.value = false;
  messageInput.value = "";
}

function submitApproval(taskId, approved) {
  if (!taskId) {
    return;
  }
  emit("submit-approval", {
    taskId,
    approved,
  });
}

function formatTime(timestamp) {
  if (!timestamp) {
    return "";
  }
  const date = new Date(timestamp);
  return Number.isNaN(date.getTime()) ? "" : date.toLocaleTimeString();
}

function roleLabel(role) {
  switch (role) {
    case "user":
      return "你";
    case "assistant":
      return "Agent";
    default:
      return "系统";
  }
}

function isPendingApprovalMessage(message) {
  return Boolean(message.kind === "approval" && message.messageId && props.pendingApprovalByMessageId[message.messageId]);
}

function getPendingApprovalForMessage(message) {
  if (!message?.messageId) {
    return null;
  }
  return props.pendingApprovalByMessageId[message.messageId] || null;
}

function handleTaskIdInput() {
  isTaskIdDirty.value = true;
}
</script>

<template>
  <section class="agent-panel">
    <header class="agent-panel-header">
      <h3>AI Agent</h3>
      <el-tag size="small" :type="connected ? 'success' : 'info'">
        {{ connected ? "在线" : "离线" }}
      </el-tag>
    </header>

    <div class="agent-block">
      <div class="block-title">创建任务</div>
      <el-input
        v-model="taskInput"
        type="textarea"
        :rows="3"
        placeholder="用自然语言描述你要执行的任务"
        :disabled="!connected"
      />
      <el-button class="block-action" type="primary" :disabled="!connected || !taskInput.trim()" @click="submitTask">
        发送任务
      </el-button>
    </div>

    <div class="agent-block">
      <div class="block-title">继续对话</div>
      <el-input v-model="messageTaskId" placeholder="任务 ID" :disabled="!connected" @input="handleTaskIdInput" />
      <el-input
        v-model="messageInput"
        type="textarea"
        :rows="2"
        placeholder="发送给 Agent 的补充信息"
        :disabled="!connected"
      />
      <el-button
        class="block-action"
        :disabled="!connected || !messageTaskId.trim() || !messageInput.trim()"
        @click="submitMessage"
      >
        发送消息
      </el-button>
    </div>

    <div class="agent-block chat-block">
      <div class="block-title">会话</div>
      <div v-if="messages.length === 0" class="chat-empty">暂无 Agent 消息</div>
      <div v-else class="chat-list">
        <div
          v-for="item in messages"
          :key="item.id"
          class="chat-item"
          :class="[`role-${item.role || 'system'}`]"
        >
          <div class="chat-meta">
            <span>{{ roleLabel(item.role) }}</span>
            <span>{{ formatTime(item.timestamp) }}</span>
          </div>
          <div class="chat-bubble">
            <div class="chat-content">{{ item.content || (item.streaming ? "正在生成..." : "") }}</div>
            <div v-if="item.streaming" class="chat-streaming">正在生成...</div>
            <div v-if="isPendingApprovalMessage(item)" class="approval-card">
              <div class="approval-card-title">待审批</div>
              <div class="approval-question">{{ getPendingApprovalForMessage(item)?.question || item.content }}</div>
              <div class="approval-actions">
                <el-button size="small" type="success" :disabled="!connected" @click="submitApproval(item.taskId, true)">批准</el-button>
                <el-button size="small" type="danger" :disabled="!connected" @click="submitApproval(item.taskId, false)">拒绝</el-button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.agent-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 10px;
  border-radius: 4px;
  border: 1px solid var(--el-border-color-light);
  background: #ffffff;
  overflow: hidden;
}

.agent-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.agent-panel-header h3 {
  margin: 0;
  font-size: 14px;
}

.agent-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.block-title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.block-action {
  align-self: flex-start;
}

.chat-block {
  min-height: 0;
  flex: 1;
}

.chat-empty {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.chat-list {
  min-height: 0;
  height: 100%;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-right: 2px;
}

.chat-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.chat-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.chat-bubble {
  max-width: 92%;
  border-radius: 10px;
  padding: 8px 10px;
  border: 1px solid var(--el-border-color-light);
  background: #f5f7fa;
}

.chat-content {
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 13px;
  color: var(--el-text-color-primary);
}

.chat-streaming {
  margin-top: 6px;
  font-size: 12px;
  color: var(--el-color-primary);
}

.approval-card {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px dashed var(--el-border-color);
}

.approval-card-title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
}

.approval-question {
  margin-bottom: 8px;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 13px;
  color: var(--el-text-color-primary);
}

.approval-actions {
  display: flex;
  gap: 8px;
}

.role-user {
  align-items: flex-end;
}

.role-user .chat-bubble {
  background: #ecf5ff;
  border-color: #c6e2ff;
}

.role-assistant {
  align-items: flex-start;
}

.role-assistant .chat-bubble {
  background: #f0f9eb;
  border-color: #d9ecff;
}

.role-system {
  align-items: center;
}

.role-system .chat-bubble {
  max-width: 100%;
  background: #f4f4f5;
}

@media screen and (max-width: 768px) {
  .chat-bubble {
    max-width: 100%;
  }
}
</style>
