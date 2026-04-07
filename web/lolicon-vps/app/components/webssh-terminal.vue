<script setup>
import { Monitor, Loading, CircleCloseFilled } from "@element-plus/icons-vue";
import "@xterm/xterm/css/xterm.css";
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';

const props = defineProps({
  status: String,
  errorMessage: String,
});

const emit = defineEmits(["input", "resize"]);

const terminalRef = ref(null);
let terminal = null;
let fitAddon = null;

onMounted(async () => {
  terminal = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: "'Sarasa Mono SC', 'Noto Sans Mono CJK SC', 'Microsoft YaHei Mono', 'Cascadia Mono', Consolas, monospace",
    theme: {
      background: "#1e1e2e",
      foreground: "#cdd6f4",
      cursor: "#f5e0dc",
      selectionBackground: "#585b70",
      black: "#45475a",
      red: "#f38ba8",
      green: "#a6e3a1",
      yellow: "#f9e2af",
      blue: "#89b4fa",
      magenta: "#f5c2e7",
      cyan: "#94e2d5",
      white: "#bac2de",
      brightBlack: "#585b70",
      brightRed: "#f38ba8",
      brightGreen: "#a6e3a1",
      brightYellow: "#f9e2af",
      brightBlue: "#89b4fa",
      brightMagenta: "#f5c2e7",
      brightCyan: "#94e2d5",
      brightWhite: "#a6adc8",
    },
  });

  fitAddon = new FitAddon();
  terminal.loadAddon(fitAddon);
  terminal.loadAddon(new WebLinksAddon());

  terminal.open(terminalRef.value);
  fitAddon.fit();

  terminal.onData((data) => {
    emit("input", data);
  });

  terminal.onResize(({ cols, rows }) => {
    emit("resize", { cols, rows });
  });

  window.addEventListener("resize", handleResize);
});

function handleResize() {
  if (fitAddon) {
    fitAddon.fit();
  }
}

function write(data) {
  if (terminal) {
    terminal.write(data);
  }
}

function clear() {
  if (terminal) {
    terminal.clear();
    terminal.reset();
  }
}

function getDimensions() {
  if (terminal) {
    return { cols: terminal.cols, rows: terminal.rows };
  }
  return { cols: 80, rows: 24 };
}

defineExpose({ write, clear, getDimensions });

onUnmounted(() => {
  window.removeEventListener("resize", handleResize);
  if (terminal) {
    terminal.dispose();
  }
});
</script>

<template>
  <div class="webssh-terminal-wrapper">
    <div
      v-if="status === 'disconnected'"
      class="terminal-placeholder"
    >
      <div class="placeholder-content">
        <el-icon size="48" color="#909399"><Monitor /></el-icon>
        <p>选择一个连接或创建新连接以开始</p>
      </div>
    </div>
    <div
      v-if="status === 'connecting'"
      class="terminal-placeholder"
    >
      <div class="placeholder-content">
        <el-icon size="48" class="is-loading" color="#39c5bb"><Loading /></el-icon>
        <p>正在连接...</p>
      </div>
    </div>
    <div
      v-if="status === 'error' && errorMessage"
      class="terminal-placeholder"
    >
      <div class="placeholder-content error">
        <el-icon size="48" color="#f56c6c"><CircleCloseFilled /></el-icon>
        <p>{{ errorMessage }}</p>
      </div>
    </div>
    <div ref="terminalRef" class="terminal-container" :class="{ hidden: status !== 'connected' }"></div>
  </div>
</template>

<style scoped>
.webssh-terminal-wrapper {
  width: 100%;
  height: 100%;
  position: relative;
  background: #1e1e2e;
  border-radius: 4px;
  overflow: hidden;
}

.terminal-container {
  width: 100%;
  height: 100%;
  padding: 4px;
  box-sizing: border-box;
  overflow: hidden;
}

.terminal-container.hidden {
  visibility: hidden;
}

.terminal-placeholder {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.placeholder-content {
  text-align: center;
}

.placeholder-content p {
  margin-top: 16px;
  font-size: 14px;
  color: #909399;
}

.placeholder-content.error p {
  color: #f56c6c;
  max-width: 400px;
}
</style>
