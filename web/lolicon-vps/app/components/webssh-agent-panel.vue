<script setup>
const props = defineProps({
  status: {
    type: String,
    default: "disconnected",
  },
  timeline: {
    type: Array,
    default: () => [],
  },
  pendingApproval: {
    type: Object,
    default: null,
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

watch(
  () => props.activeTaskId,
  (value) => {
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
  messageInput.value = "";
}

function submitApproval(approved) {
  if (!props.pendingApproval?.taskId) {
    return;
  }
  emit("submit-approval", {
    taskId: props.pendingApproval.taskId,
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
      <el-input v-model="messageTaskId" placeholder="任务 ID" :disabled="!connected" />
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

    <div v-if="pendingApproval" class="agent-block approval-block">
      <div class="block-title">待审批</div>
      <p class="approval-question">{{ pendingApproval.question }}</p>
      <div class="approval-actions">
        <el-button size="small" type="success" @click="submitApproval(true)">批准</el-button>
        <el-button size="small" type="danger" @click="submitApproval(false)">拒绝</el-button>
      </div>
    </div>

    <div class="agent-block timeline-block">
      <div class="block-title">执行时间线</div>
      <div v-if="timeline.length === 0" class="timeline-empty">暂无 Agent 事件</div>
      <div v-else class="timeline-list">
        <div v-for="item in timeline" :key="item.id" class="timeline-item">
          <div class="timeline-meta">
            <span>{{ item.label }}</span>
            <span>{{ formatTime(item.timestamp) }}</span>
          </div>
          <div class="timeline-content">{{ item.content }}</div>
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

.approval-block {
  background: #f5f7fa;
  padding: 8px;
  border-radius: 4px;
}

.approval-question {
  margin: 0;
  font-size: 13px;
  color: var(--el-text-color-primary);
}

.approval-actions {
  display: flex;
  gap: 8px;
}

.timeline-block {
  min-height: 0;
  flex: 1;
}

.timeline-empty {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.timeline-list {
  min-height: 0;
  height: 100%;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.timeline-item {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  padding: 6px 8px;
}

.timeline-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 4px;
}

.timeline-content {
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 13px;
  color: var(--el-text-color-primary);
}
</style>
