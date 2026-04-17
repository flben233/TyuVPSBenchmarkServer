<script setup>
import { ChatDotRound, Promotion, Loading, Plus, Check, Close, WarnTriangleFilled, ArrowDown, Delete, CircleCloseFilled, Setting } from "@element-plus/icons-vue";
import { error, success, warn } from "~/utils/message.js";
import { marked } from "marked";

marked.setOptions({
  breaks: true,
  gfm: true,
});

const props = defineProps({
  sshSessionId: String,
  connected: Boolean,
});

const {
  currentSessionId,
  conversationId,
  messages,
  streaming,
  thinking,
  awaitingApproval,
  pendingToolCall,
  sessions,
  newSession,
  switchSession,
  removeSession,
  llmSettings,
  updateLLMSettings,
  permanentAllowedCommands,
  saveWhitelist,
  sendMessage,
  sendApproval,
  stopChat,
} = useWebSSHAgent();

const inputText = ref("");
const chatContainer = ref(null);
const autoScroll = ref(true);
const creating = ref(false);
const expandedThinkings = ref(new Set());
const showSessionList = ref(false);
const showSettingsDialog = ref(false);
const settingsSubmitting = ref(false);

const initialized = computed(() => !!conversationId.value);

const sessionList = computed(() => {
  return Object.values(sessions.value).sort((a, b) => b.updatedAt - a.updatedAt);
});

const currentSessionName = computed(() => {
  if (!currentSessionId.value) return "";
  return sessions.value[currentSessionId.value]?.name || "New Chat";
});

function toggleThinking(idx) {
  const s = new Set(expandedThinkings.value);
  if (s.has(idx)) {
    s.delete(idx);
  } else {
    s.add(idx);
  }
  expandedThinkings.value = s;
}

function isThinkingExpanded(idx) {
  return expandedThinkings.value.has(idx);
}

function renderMarkdown(text) {
  if (!text) return "";
  return marked.parse(text);
}

function formatRelativeTime(ts) {
  if (!ts) return "";
  const diff = Date.now() - ts;
  if (diff < 60000) return "刚刚";
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`;
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`;
  return `${Math.floor(diff / 86400000)} 天前`;
}

watch(
  () => messages.value.length,
  () => {
    if (autoScroll.value) {
      nextTick(() => {
        if (chatContainer.value) {
          chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
        }
      });
    }
  }
);

watch(streaming, (val) => {
  if (!val && autoScroll.value) {
    nextTick(() => {
      if (chatContainer.value) {
        chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
      }
    });
  }
});

function handleScroll() {
  if (!chatContainer.value) return;
  const el = chatContainer.value;
  const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
  autoScroll.value = atBottom;
}

async function handleNewSession() {
  if (!props.sshSessionId) return;
  try {
    creating.value = true;
    await newSession(props.sshSessionId);
  } catch (e) {
    error("创建对话失败: " + (e.message || "unknown error"));
  } finally {
    creating.value = false;
  }
}

async function handleSwitchSession(id) {
  creating.value = true;
  try {
    await switchSession(id, props.sshSessionId);
  } catch (e) {
    error("切换会话失败: " + (e.message || "unknown error"));
  } finally {
    creating.value = false;
  }
  showSessionList.value = false;
}

function handleDeleteSession(id) {
  removeSession(id);
}

async function handleSend() {
  const text = inputText.value.trim();
  if (!text || streaming.value) return;
  inputText.value = "";
  try {
    await sendMessage(text);
  } catch (e) {
    error("发送失败: " + (e.message || "unknown error"));
  }
}

function handleKeyDown(e) {
  if (e.key === "Enter" && !e.shiftKey) {
    e.preventDefault();
    handleSend();
  }
}

function formatTime(ts) {
  return new Date(ts).toLocaleTimeString();
}

async function handleSaveLLMSettings(nextSettings) {
  if (nextSettings.enabled) {
    if (!nextSettings.apiBase || !nextSettings.apiKey || !nextSettings.model) {
      warn("启用自定义 API 时必须填写 API Base、API Key 和 Model");
      return;
    }
  }

  settingsSubmitting.value = true;
  try {
    updateLLMSettings(nextSettings);
    showSettingsDialog.value = false;
    success("设置已保存");
  } catch (e) {
    error("保存失败: " + (e.message || e.data?.message || "未知错误"));
  } finally {
    settingsSubmitting.value = false;
  }
}

async function handleSaveWhitelist(commands) {
  try {
    await saveWhitelist(commands);
  } catch (e) {
    error("白名单保存失败: " + (e.message || "未知错误"));
  }
}
</script>

<template>
  <div class="agent-chat">
    <div class="chat-header">
      <template v-if="initialized">
        <el-popover
          placement="bottom-start"
          :width="260"
          trigger="click"
          v-model:visible="showSessionList"
        >
          <template #reference>
            <span class="session-selector">
              <span class="session-selector-name">{{ currentSessionName }}</span>
              <el-icon class="session-selector-arrow"><ArrowDown /></el-icon>
            </span>
          </template>
          <div class="session-popover">
            <div class="session-popover-title">会话列表</div>
            <div class="session-popover-list">
              <div
                v-for="s in sessionList"
                :key="s.id"
                class="session-popover-item"
                :class="{ active: s.id === currentSessionId }"
                @click="handleSwitchSession(s.id)"
              >
                <div class="session-popover-item-info">
                  <span class="session-popover-item-name">{{ s.name }}</span>
                  <span class="session-popover-item-time">{{ formatRelativeTime(s.updatedAt) }}</span>
                </div>
                <el-icon
                  class="session-popover-item-delete"
                  @click.stop="handleDeleteSession(s.id)"
                >
                  <Delete />
                </el-icon>
              </div>
              <div v-if="sessionList.length === 0" class="session-popover-empty">
                暂无保存的会话
              </div>
            </div>
          </div>
        </el-popover>
      </template>
      <span v-else class="chat-title">
        <el-icon><ChatDotRound /></el-icon>
        AI Assistant
      </span>
      <div class="header-actions">
        <el-button
          size="small"
          text
          :icon="Setting"
          @click="showSettingsDialog = true"
          title="LLM API 设置"
        />
        <el-button
          v-if="connected && sshSessionId"
          size="small"
          text
          :icon="Plus"
          :loading="creating"
          @click="handleNewSession"
          title="新建对话"
        />
      </div>
    </div>

    <div v-if="!initialized" class="chat-placeholder">
      <template v-if="connected && sshSessionId">
        <el-button type="primary" @click="handleNewSession" :loading="creating">
          新建 AI 对话
        </el-button>
        <div v-if="sessionList.length > 0" class="saved-sessions">
          <p class="saved-sessions-title">历史会话</p>
          <div
            v-for="s in sessionList"
            :key="s.id"
            class="saved-session-item"
            @click="handleSwitchSession(s.id)"
          >
            <div class="saved-session-item-info">
              <span class="saved-session-item-name">{{ s.name }}</span>
              <span class="saved-session-item-time">{{ formatRelativeTime(s.updatedAt) }}</span>
            </div>
          </div>
        </div>
        <p v-else class="placeholder-hint">与 AI 助手对话，帮助管理此 SSH 会话</p>
      </template>
      <template v-else>
        <p class="placeholder-hint">请先连接 SSH 会话以使用 AI 助手</p>
      </template>
    </div>

    <template v-else>
      <div ref="chatContainer" class="chat-messages" @scroll="handleScroll">
        <div
          v-for="(msg, idx) in messages"
          :key="idx"
          class="chat-message"
          :class="[`msg-${msg.role}`, { 'msg-error': msg.isError }]"
        >
          <div class="msg-avatar">
            <div v-if="msg.role === 'user'" class="avatar avatar-user">You</div>
            <div v-else class="avatar avatar-ai">AI</div>
          </div>
          <div class="msg-body">
            <div class="msg-meta">
              <span class="msg-role">{{ msg.role === 'user' ? 'You' : 'AI' }}</span>
              <span class="msg-time">{{ formatTime(msg.timestamp) }}</span>
            </div>
            <template v-if="msg.role === 'assistant'">
              <div
                v-if="msg.thinkingContent"
                class="thinking-block"
                :class="{ 'thinking-expanded': isThinkingExpanded(idx) }"
              >
                <div class="thinking-header" @click="toggleThinking(idx)">
                  <el-icon class="is-loading" v-if="streaming && thinking && idx === messages.length - 1"><Loading /></el-icon>
                  <el-icon v-else><ArrowDown /></el-icon>
                  <span>{{ isThinkingExpanded(idx) ? '隐藏思考过程' : '查看思考过程' }}</span>
                </div>
                <div v-if="isThinkingExpanded(idx)" class="thinking-content">
                  <pre>{{ msg.thinkingContent }}</pre>
                </div>
              </div>
              <div
                v-if="msg.content"
                class="msg-content msg-markdown"
                v-html="renderMarkdown(msg.content)"
              />
              <div
                v-else-if="streaming && thinking && idx === messages.length - 1 && !msg.thinkingContent"
                class="msg-content"
              >
                <span class="thinking-indicator"><el-icon class="is-loading"><Loading /></el-icon> 正在思考...</span>
              </div>
              <div v-else-if="!msg.content && !msg.thinkingContent && !msg.isError" class="msg-content">
                <span class="thinking-indicator"><el-icon class="is-loading"><Loading /></el-icon></span>
              </div>
            </template>
            <div v-else class="msg-content">
              <span>{{ msg.content }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="awaitingApproval && pendingToolCall" class="approval-bar">
        <el-icon color="#e6a23c"><WarnTriangleFilled /></el-icon>
        <div class="approval-info">
          <div class="approval-text">
            <span>以下指令需要审批: <strong>{{ (pendingToolCall.disallowed_commands || []).join(', ') }}</strong></span>
            <span class="approval-cmd" v-if="pendingToolCall.args?.command">
              完整命令: <code>{{ pendingToolCall.args.command }}</code>
            </span>
          </div>
          <div class="approval-actions">
            <el-button type="success" size="small" :icon="Check" @click="sendApproval(true)" :disabled="streaming">
              允许一次
            </el-button>
            <el-button type="warning" size="small" :icon="Check" @click="sendApproval(true, true)" :disabled="streaming">
              本次会话允许
            </el-button>
            <el-button type="danger" size="small" :icon="Close" @click="sendApproval(false)" :disabled="streaming">
              拒绝
            </el-button>
          </div>
        </div>
      </div>

      <div class="chat-input-area">
        <el-input
          v-model="inputText"
          type="textarea"
          :autosize="{ minRows: 1, maxRows: 4 }"
          placeholder="发送消息... (Enter 发送, Shift+Enter 换行)"
          :disabled="streaming || awaitingApproval"
          resize="none"
          @keydown="handleKeyDown"
        />
        <el-button
          v-if="streaming"
          type="danger"
          :icon="CircleCloseFilled"
          circle
          @click="stopChat"
          class="stop-btn"
        />
        <el-button
          v-else
          type="primary"
          :icon="Promotion"
          circle
          :disabled="!inputText.trim() || awaitingApproval"
          @click="handleSend"
          class="send-btn"
        />
      </div>
    </template>
  </div>

  <WebsshLlmSettingsDialog
    v-model="showSettingsDialog"
    :settings="llmSettings"
    :allowed-commands="permanentAllowedCommands"
    :submitting="settingsSubmitting"
    @save="handleSaveLLMSettings"
    @save-whitelist="handleSaveWhitelist"
  />
</template>

<style scoped>
.agent-chat {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #fff;
  border-radius: 6px;
  border: 1px solid var(--el-border-color-lighter);
  box-sizing: border-box;
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.chat-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.session-selector {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.2s;
  max-width: 280px;
}

.session-selector:hover {
  background: #f5f7fa;
}

.session-selector-name {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-selector-arrow {
  font-size: 12px;
  color: #909399;
  flex-shrink: 0;
}

.session-popover {
  margin: -12px;
}

.session-popover-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  padding: 10px 14px 8px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.session-popover-list {
  max-height: 300px;
  overflow-y: auto;
  padding: 4px 0;
}

.session-popover-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 14px;
  cursor: pointer;
  transition: background 0.15s;
}

.session-popover-item:hover {
  background: #f5f7fa;
}

.session-popover-item.active {
  background: var(--el-color-primary-light-9);
}

.session-popover-item-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.session-popover-item-name {
  font-size: 13px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-popover-item-time {
  font-size: 11px;
  color: #c0c4cc;
}

.session-popover-item-delete {
  font-size: 14px;
  color: #c0c4cc;
  flex-shrink: 0;
  margin-left: 8px;
  transition: color 0.15s;
}

.session-popover-item-delete:hover {
  color: #f56c6c;
}

.session-popover-empty {
  padding: 16px 14px;
  text-align: center;
  font-size: 13px;
  color: #c0c4cc;
}

.chat-placeholder {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 24px;
}

.placeholder-hint {
  font-size: 13px;
  color: #909399;
  text-align: center;
  max-width: 240px;
  line-height: 1.6;
}

.saved-sessions {
  width: 100%;
  max-width: 300px;
}

.saved-sessions-title {
  font-size: 12px;
  color: #909399;
  margin: 0 0 8px;
  text-align: center;
}

.saved-session-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s;
}

.saved-session-item:hover {
  background: #f5f7fa;
}

.saved-session-item-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.saved-session-item-name {
  font-size: 13px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.saved-session-item-time {
  font-size: 11px;
  color: #c0c4cc;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.chat-message {
  display: flex;
  gap: 10px;
  max-width: 100%;
}

.msg-user {
  flex-direction: row-reverse;
}

.msg-avatar {
  flex-shrink: 0;
  margin-top: 2px;
}

.avatar {
  width: 30px;
  height: 30px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
}

.avatar-user {
  background: var(--el-color-primary-light-8);
  color: var(--el-color-primary-dark-2);
}

.avatar-ai {
  background: #e8f5e9;
  color: #388e3c;
}

.msg-body {
  min-width: 0;
  flex: 1;
}

.msg-user .msg-body {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.msg-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.msg-role {
  font-size: 12px;
  font-weight: 600;
  color: #606266;
}

.msg-time {
  font-size: 11px;
  color: #c0c4cc;
}

.msg-content {
  font-size: 13px;
  line-height: 1.6;
  color: #303133;
}

.msg-user .msg-content {
  background: var(--el-color-primary-light-9);
  border: 1px solid var(--el-color-primary-light-7);
  padding: 6px 12px;
  border-radius: 6px 6px 2px 6px;
  display: inline-block;
  word-break: break-word;
}

.msg-assistant .msg-content.msg-markdown {
  background: #f9fafb;
  border: 1px solid var(--el-border-color-lighter);
  padding: 8px 12px;
  border-radius: 6px 6px 6px 2px;
  word-break: break-word;
}

.msg-error .msg-content {
  background: #fef0f0;
  border: 1px solid #fbc4c4;
  padding: 6px 12px;
  border-radius: 6px;
  color: #f56c6c;
}

.thinking-block {
  margin-bottom: 8px;
  border-radius: 6px;
  background: #fafafa;
  border: 1px solid var(--el-border-color-lighter);
  overflow: hidden;
}

.thinking-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  cursor: pointer;
  font-size: 12px;
  color: #909399;
  user-select: none;
  transition: background 0.2s;
}

.thinking-header:hover {
  background: #f5f5f5;
}

.thinking-header .el-icon {
  transition: transform 0.25s;
}

.thinking-block.thinking-expanded .thinking-header .el-icon {
  transform: rotate(180deg);
}

.thinking-content {
  max-height: 200px;
  overflow-y: auto;
  padding: 0 12px 8px;
  border-top: 1px solid var(--el-border-color-extra-light);
}

.thinking-content pre {
  margin: 0;
  padding: 8px 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 12px;
  line-height: 1.5;
  color: #909399;
  font-family: inherit;
}

.thinking-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #909399;
  font-size: 13px;
}

.approval-bar {
  padding: 10px 14px;
  background: #fdf6ec;
  border-top: 1px solid #f5dab1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.approval-info {
  gap: 6px;
  font-size: 13px;
  color: #606266;
  flex: 1;
  min-width: 0;
}

.approval-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.approval-cmd code {
  background: rgba(0, 0, 0, 0.06);
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 12px;
  font-family: 'Cascadia Mono', Consolas, monospace;
  word-break: break-all;
}

.approval-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
  margin-top: 6px;
}

.chat-input-area {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  padding: 10px 14px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.chat-input-area :deep(.el-textarea) {
  flex: 1;
}

.chat-input-area :deep(.el-textarea__inner) {
  font-size: 13px;
  padding: 8px 12px;
  border-radius: 6px;
}

.send-btn {
  flex-shrink: 0;
  margin-bottom: 1px;
}

.stop-btn {
  flex-shrink: 0;
  margin-bottom: 1px;
}
</style>

<style>
.msg-markdown p {
  margin: 0 0 8px;
}
.msg-markdown p:last-child {
  margin-bottom: 0;
}
.msg-markdown code {
  background: #f0f2f5;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
  font-family: 'Cascadia Mono', Consolas, monospace;
}
.msg-markdown pre {
  margin: 8px 0;
  padding: 10px 12px;
  background: #1e1e2e;
  color: #cdd6f4;
  border-radius: 6px;
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.5;
}
.msg-markdown pre code {
  background: none;
  padding: 0;
  color: inherit;
}
.msg-markdown ul,
.msg-markdown ol {
  margin: 4px 0;
  padding-left: 20px;
}
.msg-markdown li {
  margin: 2px 0;
}
.msg-markdown blockquote {
  margin: 8px 0;
  padding: 4px 12px;
  border-left: 3px solid var(--el-color-primary-light-5);
  color: #606266;
  background: #f9fafc;
  border-radius: 0 4px 4px 0;
}
.msg-markdown h1,
.msg-markdown h2,
.msg-markdown h3 {
  margin: 8px 0 4px;
  font-weight: 600;
}
.msg-markdown h1 { font-size: 16px; }
.msg-markdown h2 { font-size: 15px; }
.msg-markdown h3 { font-size: 14px; }
.msg-markdown a {
  color: var(--el-color-primary);
  text-decoration: none;
}
.msg-markdown a:hover {
  text-decoration: underline;
}
.msg-markdown table {
  border-collapse: collapse;
  margin: 8px 0;
  font-size: 12px;
}
.msg-markdown th,
.msg-markdown td {
  border: 1px solid var(--el-border-color-lighter);
  padding: 4px 8px;
}
.msg-markdown th {
  background: #f5f7fa;
}
</style>
