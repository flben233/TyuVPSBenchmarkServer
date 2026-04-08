<script setup>
import { saveConnection } from "~/utils/webssh-storage";

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
  onAgentUpdate,
  onAgentApproval,
  onAgentError,
  onAgentDone,
} = useWebSSH();

const terminalRef = ref(null);
const selectedConnection = ref(null);
const agentTimeline = ref([]);
const pendingApproval = ref(null);
const activeAgentTaskId = ref("");

function pushAgentTimeline(label, content, taskId = "") {
  agentTimeline.value.unshift({
    id: crypto.randomUUID(),
    label,
    content,
    taskId,
    timestamp: Date.now(),
  });
}

watch(
  () => status.value,
  (nextStatus) => {
    if (nextStatus === "disconnected" || nextStatus === "error") {
      pendingApproval.value = null;
      activeAgentTaskId.value = "";
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
  pushAgentTimeline("任务提交", message);
}

function handleAgentMessageSubmit({ taskId, message }) {
  if (!sendAgentMessage(taskId, message)) {
    ElMessage.warning("当前连接不可用，无法发送消息");
    return;
  }
  activeAgentTaskId.value = taskId;
  pushAgentTimeline("用户消息", message, taskId);
}

function handleAgentApprovalSubmit({ taskId, approved }) {
  if (!sendAgentApproval(taskId, approved)) {
    ElMessage.warning("当前连接不可用，无法发送审批");
    return;
  }
  pushAgentTimeline("审批响应", approved ? "已批准" : "已拒绝", taskId);
  pendingApproval.value = null;
}

onMounted(() => {
  onOutput((data) => {
    if (terminalRef.value) {
      terminalRef.value.write(data);
    }
  });

  onAgentUpdate((event) => {
    activeAgentTaskId.value = event.taskId;
    pushAgentTimeline("执行更新", event.message, event.taskId);
  });

  onAgentApproval((event) => {
    activeAgentTaskId.value = event.taskId;
    pendingApproval.value = event;
    pushAgentTimeline("审批请求", event.question, event.taskId);
  });

  onAgentError((event) => {
    activeAgentTaskId.value = event.taskId;
    pushAgentTimeline("执行错误", event.message, event.taskId);
  });

  onAgentDone((event) => {
    activeAgentTaskId.value = event.taskId;
    pendingApproval.value = null;
    pushAgentTimeline("执行完成", event.summary, event.taskId);
  });
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
            :timeline="agentTimeline"
            :pending-approval="pendingApproval"
            :active-task-id="activeAgentTaskId"
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
