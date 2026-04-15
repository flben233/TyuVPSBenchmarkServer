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
    messages.value.push({ role, content, thinkingContent: "", timestamp: Date.now() });
  }

  function updateAssistant(idx, extras) {
    const msg = { ...messages.value[idx], timestamp: Date.now(), ...extras };
    messages.value[idx] = msg;
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

  async function streamChat(body, assistantIdx) {
    const resp = await fetchWithAuth("/webssh/llm/chat", {
      method: "POST",
      body: JSON.stringify(body),
    });

    if (!resp.ok) {
      throw new Error(`Chat request failed: ${resp.status}`);
    }

    let thinkingText = "";
    let tokenText = "";

    await parseSSEStream(resp.body.getReader(), (event, payload) => {
      const inner = payload.payload || {};
      switch (event) {
        case "thinking":
          if (inner.text) {
            thinkingText += inner.text;
            updateAssistant(assistantIdx, { thinkingContent: thinkingText, content: tokenText });
          }
          break;
        case "token":
          thinking.value = false;
          if (inner.text) {
            tokenText += inner.text;
            updateAssistant(assistantIdx, { thinkingContent: thinkingText, content: tokenText });
          }
          break;
        case "error":
          updateAssistant(assistantIdx, { content: inner.message || "Unknown error", isError: true });
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

    messages.value.push({ role: "assistant", content: "", thinkingContent: "", timestamp: Date.now() });
    const assistantIdx = messages.value.length - 1;

    try {
      await streamChat(
        { conversationId: conversationId.value, message: text },
        assistantIdx
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

    messages.value.push({ role: "assistant", content: "", thinkingContent: "", timestamp: Date.now() });
    const assistantIdx = messages.value.length - 1;

    try {
      await streamChat(
        { conversationId: conversationId.value, approval_granted: granted },
        assistantIdx
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
