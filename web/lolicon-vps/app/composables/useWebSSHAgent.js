import { requestWithAuth, fetchWithAuth } from "~/composables/useAuth.js";

function parseSSEStream(reader, onEvent) {
  const decoder = new TextDecoder();
  let buffer = "";
  let currentEvent = "";

  function processLines(lines) {
    for (const line of lines) {
      if (line.startsWith("event: ")) {
        currentEvent = line.slice(7).trim();
      } else if (line.startsWith("data: ")) {
        const dataStr = line.slice(6).trim();
        if (!dataStr) continue;
        let payload;
        try {
          payload = JSON.parse(dataStr);
        } catch {
          continue;
        }
        onEvent(currentEvent, payload);
        currentEvent = "";
      } else if (line.trim() === "") {
        currentEvent = "";
      }
    }
  }

  return (async () => {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split("\n");
      buffer = lines.pop() || "";
      processLines(lines);
    }
    if (buffer.trim()) {
      processLines([buffer]);
    }
  })();
}

export function useWebSSHAgent() {
  const conversationId = ref(null);
  const messages = ref([]);
  const streaming = ref(false);
  const thinking = ref(false);
  const awaitingApproval = ref(false);
  const pendingToolCall = ref(null);

  function addMessage(role, content) {
    messages.value.push({ role, content, timestamp: Date.now() });
  }

  function updateAssistant(idx, content, extras = {}) {
    messages.value[idx] = { role: "assistant", content, timestamp: Date.now(), ...extras };
  }

  async function createConversation(sshSessionId) {
    const resp = await requestWithAuth("/webssh/llm/new", "POST", {
      body: { sshSessionId },
    });
    if (resp && resp.conversationId) {
      conversationId.value = resp.conversationId;
      messages.value = [];
      return resp.conversationId;
    }
    throw new Error(resp?.message || "Failed to create conversation");
  }

  async function streamChat(body, assistantIdx, assistantContent) {
    const resp = await fetchWithAuth("/webssh/llm/chat", {
      method: "POST",
      body: JSON.stringify(body),
    });

    if (!resp.ok) {
      throw new Error(`Chat request failed: ${resp.status}`);
    }

    await parseSSEStream(resp.body.getReader(), (event, payload) => {
      const inner = payload.payload || {};
      switch (event) {
        case "token":
        case "thinking":
          thinking.value = false;
          if (inner.text) {
            assistantContent.value += inner.text;
            updateAssistant(assistantIdx, assistantContent.value);
          }
          break;
        case "error":
          updateAssistant(assistantIdx, inner.message || "Unknown error", { isError: true });
          break;
        case "awaiting_approval":
          awaitingApproval.value = true;
          pendingToolCall.value = inner;
          break;
      }
    });
  }

  async function sendMessage(text) {
    if (!conversationId.value) {
      throw new Error("No active conversation");
    }
    addMessage("user", text);
    streaming.value = true;
    thinking.value = true;
    awaitingApproval.value = false;
    pendingToolCall.value = null;

    const assistantContent = ref("");
    messages.value.push({ role: "assistant", content: "", timestamp: Date.now() });
    const assistantIdx = messages.value.length - 1;

    try {
      await streamChat(
        { conversationId: conversationId.value, message: text },
        assistantIdx,
        assistantContent
      );
    } finally {
      streaming.value = false;
      thinking.value = false;
    }
  }

  async function sendApproval(granted) {
    if (!conversationId.value) return;
    streaming.value = true;
    awaitingApproval.value = false;

    addMessage("user", granted ? "[approved]" : "[rejected]");

    const assistantContent = ref("");
    messages.value.push({ role: "assistant", content: "", timestamp: Date.now() });
    const assistantIdx = messages.value.length - 1;

    try {
      await streamChat(
        { conversationId: conversationId.value, approval_granted: granted },
        assistantIdx,
        assistantContent
      );
    } finally {
      streaming.value = false;
    }
  }

  function reset() {
    conversationId.value = null;
    messages.value = [];
    streaming.value = false;
    thinking.value = false;
    awaitingApproval.value = false;
    pendingToolCall.value = null;
  }

  return {
    conversationId: readonly(conversationId),
    messages,
    streaming: readonly(streaming),
    thinking: readonly(thinking),
    awaitingApproval: readonly(awaitingApproval),
    pendingToolCall: readonly(pendingToolCall),
    createConversation,
    sendMessage,
    sendApproval,
    reset,
  };
}
