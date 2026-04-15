<script setup>
import { saveConnection } from "~/utils/webssh-storage";
import { ChatDotRound } from "@element-plus/icons-vue";

useHead({
  title: "WebSSH - Lolicon VPS",
});

const { userInfo } = useAuth();
const { status, errorMessage, sshSessionId, connect, disconnect, sendInput, resize, onOutput } = useWebSSH();

const terminalRef = ref(null);
const selectedConnection = ref(null);
const showAgentPanel = ref(false);
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

onMounted(() => {
  onOutput((data) => {
    if (terminalRef.value) {
      terminalRef.value.write(data);
    }
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
            :type="showAgentPanel ? 'primary' : 'default'"
            size="small"
            :icon="ChatDotRound"
            @click="showAgentPanel = !showAgentPanel"
          >
            AI
          </el-button>
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
      <div class="webssh-content">
        <div class="webssh-terminal-area">
          <WebsshTerminal
            ref="terminalRef"
            :status="status"
            :error-message="errorMessage"
            @input="handleInput"
            @resize="handleResize"
          />
        </div>
        <transition name="slide">
          <div v-if="showAgentPanel && status === 'connected'" class="webssh-agent-panel">
            <WebsshAgentChat
              :ssh-session-id="sshSessionId"
              :connected="status === 'connected'"
            />
          </div>
        </transition>
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

.toolbar-right {
  display: flex;
  gap: 8px;
}

.webssh-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.webssh-terminal-area {
  flex: 1;
  padding: 8px;
  overflow: hidden;
  min-width: 0;
}

.webssh-agent-panel {
  width: 380px;
  flex-shrink: 0;
  border-left: 1px solid var(--el-border-color-light);
  background: #fafafa;
}

.slide-enter-active,
.slide-leave-active {
  transition: width 0.25s ease, opacity 0.25s ease;
}

.slide-enter-from,
.slide-leave-to {
  width: 0;
  opacity: 0;
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
}
</style>
