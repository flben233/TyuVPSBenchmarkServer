<script setup>
import { ChatDotRound, Promotion, Loading, Close, Check, WarnTriangleFilled } from "@element-plus/icons-vue";
import { error, warn, success } from "~/utils/message.js";

const props = defineProps({
  sshSessionId: String,
  connected: Boolean,
});

const {
  conversationId,
  messages,
  streaming,
  thinking,
  awaitingApproval,
  pendingToolCall,
  createConversation,
  sendMessage,
  sendApproval,
  reset,
} = useWebSSHAgent();

const inputText = ref("");
const chatContainer = ref(null);
const autoScroll = ref(true);
const creating = ref(false);
let userScrolled = false;

const initialized = computed(() => !!conversationId.value);

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

async function handleInit() {
  if (!props.sshSessionId) return;
  try {
    creating.value = true;
    await createConversation(props.sshSessionId);
  } catch (e) {
    error("Failed to create conversation: " + (e.message || "unknown error"));
  } finally {
    creating.value = false;
  }
}

async function handleSend() {
  const text = inputText.value.trim();
  if (!text || streaming.value) return;
  inputText.value = "";
  try {
    await sendMessage(text);
  } catch (e) {
    error("Failed to send message: " + (e.message || "unknown error"));
  }
}

function handleKeyDown(e) {
  if (e.key === "Enter" && !e.shiftKey) {
    e.preventDefault();
    handleSend();
  }
}

function handleNewConversation() {
  reset();
}

function formatTime(ts) {
  return new Date(ts).toLocaleTimeString();
}
</script>

<template>
  <div class="agent-chat">
    <div class="chat-header">
      <span class="chat-title">
        <el-icon><ChatDotRound /></el-icon>
        AI Assistant
      </span>
      <div class="chat-header-actions">
        <el-button
          v-if="initialized"
          size="small"
          :icon="Close"
          @click="handleNewConversation"
          title="New conversation"
        />
      </div>
    </div>

    <div v-if="!initialized" class="chat-placeholder">
      <template v-if="connected && sshSessionId">
        <el-button type="primary" @click="handleInit" :loading="creating">
          Start AI Conversation
        </el-button>
        <p class="placeholder-hint">Start a conversation with the AI assistant to help manage this SSH session</p>
      </template>
      <template v-else>
        <p class="placeholder-hint">Please connect to an SSH session first to use the AI assistant</p>
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
          <div class="msg-meta">
            <span class="msg-role">{{ msg.role === 'user' ? 'You' : 'AI' }}</span>
            <span class="msg-time">{{ formatTime(msg.timestamp) }}</span>
          </div>
          <div class="msg-content">
            <pre class="msg-text" v-if="msg.role === 'assistant'">{{ msg.content || (streaming && idx === messages.length - 1 ? '' : '...') }}</pre>
            <span v-else>{{ msg.content }}</span>
          </div>
        </div>
        <div v-if="thinking && streaming" class="chat-message msg-assistant">
          <div class="msg-meta">
            <span class="msg-role">AI</span>
          </div>
          <div class="msg-content">
            <span class="thinking-indicator"><el-icon class="is-loading"><Loading /></el-icon> Thinking...</span>
          </div>
        </div>
      </div>

      <div v-if="awaitingApproval && pendingToolCall" class="approval-bar">
        <div class="approval-info">
          <el-icon color="#e6a23c"><WarnTriangleFilled /></el-icon>
          <span>AI is requesting permission to execute: <strong>{{ pendingToolCall.name || 'command' }}</strong></span>
        </div>
        <div class="approval-actions">
          <el-button type="success" size="small" :icon="Check" @click="sendApproval(true)" :disabled="streaming">
            Allow
          </el-button>
          <el-button type="danger" size="small" :icon="Close" @click="sendApproval(false)" :disabled="streaming">
            Deny
          </el-button>
        </div>
      </div>

      <div class="chat-input-area">
        <el-input
          v-model="inputText"
          type="textarea"
          :rows="2"
          placeholder="Ask the AI assistant..."
          :disabled="streaming || awaitingApproval"
          resize="none"
          @keydown="handleKeyDown"
        />
        <el-button
          type="primary"
          :icon="Promotion"
          :loading="streaming"
          :disabled="!inputText.trim() || streaming || awaitingApproval"
          @click="handleSend"
          class="send-btn"
        />
      </div>
    </template>
  </div>
</template>

<style scoped>
.agent-chat {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #fafafa;
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid var(--el-border-color-light);
  background: #fff;
}

.chat-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.chat-header-actions {
  display: flex;
  gap: 4px;
}

.chat-placeholder {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 24px;
}

.placeholder-hint {
  font-size: 13px;
  color: #909399;
  text-align: center;
  max-width: 240px;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.chat-message {
  padding: 8px 12px;
  border-radius: 8px;
  max-width: 92%;
  word-break: break-word;
}

.msg-user {
  align-self: flex-end;
  background: var(--el-color-primary-light-9);
  border: 1px solid var(--el-color-primary-light-7);
}

.msg-assistant {
  align-self: flex-start;
  background: #fff;
  border: 1px solid var(--el-border-color-lighter);
}

.msg-error {
  background: #fef0f0;
  border-color: #fbc4c4;
}

.msg-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.msg-role {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  color: #909399;
}

.msg-time {
  font-size: 11px;
  color: #c0c4cc;
}

.msg-content {
  font-size: 13px;
  line-height: 1.5;
  color: #303133;
}

.msg-text {
  margin: 0;
  white-space: pre-wrap;
  font-family: inherit;
  font-size: 13px;
  line-height: 1.5;
}

.thinking-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #909399;
  font-size: 13px;
}

.approval-bar {
  padding: 8px 14px;
  background: #fdf6ec;
  border-top: 1px solid #f5dab1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.approval-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #606266;
  flex: 1;
  min-width: 0;
}

.approval-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.chat-input-area {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  padding: 10px 14px;
  border-top: 1px solid var(--el-border-color-light);
  background: #fff;
}

.chat-input-area :deep(.el-textarea__inner) {
  font-size: 13px;
}

.send-btn {
  flex-shrink: 0;
}
</style>
