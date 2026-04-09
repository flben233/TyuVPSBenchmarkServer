<script setup>
import { saveConnection } from "~/utils/webssh-storage";
import {
  shouldDedupeAgentState,
  setLatestPendingApproval,
  clearPendingApprovalForTask,
  buildApprovalsByMessageId,
} from "~/utils/webssh-agent-chat";

useHead({
  title: "WebSSH - Lolicon VPS",
});

const { userInfo } = useAuth();
const {
  status,
  errorMessage,
  connect,
  disconnect,
  sendInput,
  resize,
  sendAgentTask,
  sendAgentMessage,
  sendAgentApproval,
  onOutput,
  onAgentMessageStart,
  onAgentToken,
  onAgentMessageEnd,
  onAgentState,
} = useWebSSH();

const terminalRef = ref(null);
const selectedConnection = ref(null);
const agentMessages = ref([]);
const pendingApprovalByTaskId = ref({});
const composeTaskId = ref("");
const latestEventTaskId = ref("");
const lastAgentStateByTaskId = ref({});
const AGENT_STATE_DEDUPE_WINDOW_MS = 1500;

function pushAgentMessage(role, content, taskId = "", options = {}) {
  const entry = {
    id: crypto.randomUUID(),
    role,
    content,
    taskId,
    messageId: options.messageId || "",
    kind: options.kind || "text",
    state: options.state || "",
    streaming: options.streaming === true,
    finishReason: options.finishReason || "",
    timestamp: Date.now(),
  };
  agentMessages.value.push(entry);
  return entry;
}

function setPendingApproval(taskId, question, messageId) {
  pendingApprovalByTaskId.value = setLatestPendingApproval(
    pendingApprovalByTaskId.value,
    taskId,
    question,
    messageId
  );
}

function clearPendingApproval(taskId) {
  pendingApprovalByTaskId.value = clearPendingApprovalForTask(pendingApprovalByTaskId.value, taskId);
}

function resetPendingApprovals() {
  pendingApprovalByTaskId.value = {};
}

function upsertStateMessage(taskId, state, message) {
  const now = Date.now();
  const previous = lastAgentStateByTaskId.value[taskId] || null;
  const next = {
    taskId,
    state,
    message,
    timestamp: now,
  };

  const shouldMerge = shouldDedupeAgentState(previous, next, now, AGENT_STATE_DEDUPE_WINDOW_MS);
  if (shouldMerge && previous?.messageId) {
    const existing = findMessageByTaskAndMessageId(taskId, previous.messageId);
    if (existing) {
      existing.content = message;
      existing.timestamp = now;
      existing.state = state;
      lastAgentStateByTaskId.value = {
        ...lastAgentStateByTaskId.value,
        [taskId]: {
          ...next,
          messageId: previous.messageId,
        },
      };
      return existing;
    }
  }

  const created = pushAgentMessage("system", message, taskId, {
    state,
    messageId: `state:${taskId}:${now}`,
    kind: state === "awaiting_approval" ? "approval" : "text",
  });

  lastAgentStateByTaskId.value = {
    ...lastAgentStateByTaskId.value,
    [taskId]: {
      ...next,
      messageId: created.messageId,
    },
  };

  return created;
}

function findMessageByTaskAndMessageId(taskId, messageId) {
  for (let index = agentMessages.value.length - 1; index >= 0; index -= 1) {
    const item = agentMessages.value[index];
    if (item.taskId === taskId && item.messageId === messageId) {
      return item;
    }
  }
  return null;
}

function ensureStreamingAssistantMessage(taskId, messageId) {
  const existing = findMessageByTaskAndMessageId(taskId, messageId);
  if (existing) {
    return existing;
  }
  return pushAgentMessage("assistant", "", taskId, {
    messageId,
    streaming: true,
  });
}

function getStateMessage(state, message) {
  if (message) {
    return message;
  }
  switch (state) {
    case "thinking":
      return "Agent 正在思考";
    case "running_command":
      return "Agent 正在执行命令";
    case "awaiting_approval":
      return "Agent 等待审批";
    case "done":
      return "Agent 执行完成";
    case "error":
      return "Agent 执行失败";
    default:
      return "Agent 状态更新";
  }
}

watch(
  () => status.value,
  (nextStatus) => {
    if (nextStatus === "disconnected" || nextStatus === "error") {
      resetPendingApprovals();
      composeTaskId.value = "";
      latestEventTaskId.value = "";
      lastAgentStateByTaskId.value = {};
      agentMessages.value = [];
    }
  }
);

const statusText = computed(() => {
  switch (status.value) {
    case "connected":
      return `已连接: ${selectedConnection.value?.username}@${selectedConnection.value?.host}`;
    case "connecting":
      return "正在连接...";
    case "error":
      return errorMessage.value || "连接错误";
    default:
      return "未连接";
  }
});
const statusType = computed(() => {
  switch (status.value) {
    case "connected":
      return "success";
    case "connecting":
      return "warning";
    case "error":
      return "danger";
    default:
      return "info";
  }
});

function handleSelect(conn) {
  selectedConnection.value = conn;
}

function handleConnect(conn) {
  if (!userInfo.value) {
    ElMessage.warning("请先登录后再使用 WebSSH");
    return;
  }
  selectedConnection.value = conn;
  const terminal = terminalRef.value;
  const dims = terminal ? terminal.getDimensions() : { cols: 80, rows: 24 };
  terminal.clear();
  connect(
    conn.host,
    conn.port,
    conn.username,
    conn.authType === "password" ? conn.password : "",
    conn.authType === "privateKey" ? conn.privateKey : "",
    dims.cols,
    dims.rows
  );
  saveConnection({ ...conn, lastConnected: new Date().toISOString() });
}

function handleDisconnect() {
  disconnect();
}

function handleInput(data) {
  sendInput(data);
}

function handleResize({ cols, rows }) {
  resize(cols, rows);
}

function handleAgentTaskSubmit(message) {
  if (!sendAgentTask(message)) {
    ElMessage.warning("当前连接不可用，无法发送任务");
    return;
  }
  pushAgentMessage("user", message);
}

function handleAgentMessageSubmit({ taskId, message }) {
  if (!sendAgentMessage(taskId, message)) {
    ElMessage.warning("当前连接不可用，无法发送消息");
    return;
  }
  composeTaskId.value = taskId;
  pushAgentMessage("user", message, taskId);
}

function handleAgentApprovalSubmit({ taskId, approved }) {
  if (!sendAgentApproval(taskId, approved)) {
    ElMessage.warning("当前连接不可用，无法发送审批");
    return;
  }
  pushAgentMessage("system", approved ? "审批结果: 已批准" : "审批结果: 已拒绝", taskId);
  clearPendingApproval(taskId);
}

onMounted(() => {
  onOutput((data) => {
    if (terminalRef.value) {
      terminalRef.value.write(data);
    }
  });

  onAgentMessageStart((event) => {
    latestEventTaskId.value = event.taskId;
    ensureStreamingAssistantMessage(event.taskId, event.messageId);
  });

  onAgentToken((event) => {
    latestEventTaskId.value = event.taskId;
    const message = ensureStreamingAssistantMessage(event.taskId, event.messageId);
    message.content += event.delta;
  });

  onAgentMessageEnd((event) => {
    latestEventTaskId.value = event.taskId;
    const message = ensureStreamingAssistantMessage(event.taskId, event.messageId);
    message.streaming = false;
    message.finishReason = event.finishReason;
  });

  onAgentState((event) => {
    latestEventTaskId.value = event.taskId;
    const stateMessage = getStateMessage(event.state, event.message);
    if (event.state === "awaiting_approval") {
      const approvalMessage = upsertStateMessage(event.taskId, event.state, stateMessage);
      setPendingApproval(event.taskId, stateMessage, approvalMessage.messageId);
      return;
    }

    if (event.state === "done" || event.state === "error") {
      clearPendingApproval(event.taskId);
    }

    upsertStateMessage(event.taskId, event.state, stateMessage);
  });
});

const activePanelTaskId = computed(() => {
  if (composeTaskId.value) {
    return composeTaskId.value;
  }
  return latestEventTaskId.value;
});

const approvalsByMessageId = computed(() => {
  return buildApprovalsByMessageId(agentMessages.value, pendingApprovalByTaskId.value);
});
</script>

<template>
  <div class="webssh-page">
    <div class="webssh-sidebar">
      <WebsshConnectionList
        :active-id="selectedConnection?.id"
        :status="status"
        @select="handleSelect"
        @connect="handleConnect"
      />
    </div>
    <div class="webssh-main">
      <div class="webssh-toolbar">
        <div class="toolbar-left">
          <el-tag :type="statusType" effect="dark">
            {{ statusText }}
          </el-tag>
        </div>
        <div class="toolbar-right">
          <el-button
            v-if="status === 'connected'"
            type="danger"
            size="small"
            @click="handleDisconnect"
          >
            断开连接
          </el-button>
          <el-button
            v-else-if="selectedConnection && status === 'disconnected'"
            type="primary"
            size="small"
            @click="handleConnect(selectedConnection)"
          >
            连接
          </el-button>
        </div>
      </div>
      <div class="webssh-terminal-area">
        <div class="webssh-terminal-pane">
          <WebsshTerminal
            ref="terminalRef"
            :status="status"
            :error-message="errorMessage"
            @input="handleInput"
            @resize="handleResize"
          />
        </div>
        <div class="webssh-agent-pane">
          <WebsshAgentPanel
            :status="status"
            :messages="agentMessages"
            :pending-approval-by-message-id="approvalsByMessageId"
            :active-task-id="activePanelTaskId"
            @submit-task="handleAgentTaskSubmit"
            @submit-message="handleAgentMessageSubmit"
            @submit-approval="handleAgentApprovalSubmit"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.webssh-page {
  display: flex;
  width: 100%;
  height: 100vh;
  overflow: hidden;
}

.webssh-sidebar {
  width: 280px;
  flex-shrink: 0;
  border-right: 1px solid var(--el-border-color-light);
  background: #fff;
  height: 100%;
}

.webssh-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.webssh-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-light);
  background: #fff;
}

.webssh-terminal-area {
  flex: 1;
  padding: 8px;
  overflow: hidden;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  gap: 8px;
}

.webssh-terminal-pane {
  min-width: 0;
  min-height: 0;
}

.webssh-agent-pane {
  min-height: 0;
}

@media screen and (max-width: 768px) {
  .webssh-page {
    flex-direction: column;
  }
  .webssh-sidebar {
    width: 100%;
    height: 200px;
    border-right: none;
    border-bottom: 1px solid var(--el-border-color-light);
  }

  .webssh-terminal-area {
    grid-template-columns: 1fr;
    grid-template-rows: minmax(260px, 1fr) 280px;
  }
}
</style>
