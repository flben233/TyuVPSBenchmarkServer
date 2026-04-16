import { requestWithAuth, fetchWithAuth } from "~/composables/useAuth.js";
import { getSessions, saveSession, deleteSession as deleteStoredSession, createSessionId } from "~/utils/webssh-agent-session.js";

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
  const currentSessionId = ref(null);
  const conversationId = ref(null);
  const messages = ref([]);
  const streaming = ref(false);
  const thinking = ref(false);
  const awaitingApproval = ref(false);
  const pendingToolCall = ref(null);
  const sessions = ref({});

  let activeSshSessionId = null;

  sessions.value = getSessions();

  function getSessionName() {
    const userMsg = messages.value.find(
      (m) => m.role === "user" && m.content && !m.content.startsWith("[")
    );
    if (userMsg) {
      const text = userMsg.content;
      return text.length > 30 ? text.slice(0, 30) + "..." : text;
    }
    return "New Chat";
  }

  function persistCurrentSession() {
    if (!currentSessionId.value) return;
    saveSession({
      id: currentSessionId.value,
      conversationId: conversationId.value,
      messages: messages.value.map((m) => ({
        role: m.role,
        content: m.content,
        isError: m.isError || false,
        timestamp: m.timestamp,
      })),
      name: getSessionName(),
      sshSessionId: activeSshSessionId,
      createdAt:
        sessions.value[currentSessionId.value]?.createdAt || Date.now(),
      updatedAt: Date.now(),
    });
    sessions.value = getSessions();
  }

  function refreshSessions() {
    sessions.value = getSessions();
  }

  function addMessage(role, content) {
    messages.value.push({
      role,
      content,
      thinkingContent: "",
      timestamp: Date.now(),
    });
  }

  function updateAssistant(idx, extras) {
    const msg = { ...messages.value[idx], timestamp: Date.now(), ...extras };
    messages.value[idx] = msg;
  }

  async function createBackendConversation(sshSessionId) {
    activeSshSessionId = sshSessionId;
    const resp = await requestWithAuth("/webssh/llm/new", "POST", {
      body: { sshSessionId },
    });
    if (resp && resp.conversationId) {
      conversationId.value = resp.conversationId;
      return resp.conversationId;
    }
    throw new Error(resp?.message || "Failed to create conversation");
  }

  async function newSession(sshSessionId) {
    persistCurrentSession();
    currentSessionId.value = createSessionId();
    messages.value = [];
    conversationId.value = null;
    awaitingApproval.value = false;
    pendingToolCall.value = null;
    activeSshSessionId = sshSessionId;

    await createBackendConversation(sshSessionId);

    persistCurrentSession();
  }

  async function switchSession(id, sshSessionId) {
    persistCurrentSession();
    const session = sessions.value[id];
    if (!session) return;

    currentSessionId.value = id;
    messages.value = session.messages.map((m) => ({ ...m }));
    conversationId.value = null;
    activeSshSessionId = sshSessionId || session.sshSessionId;
    awaitingApproval.value = false;
    pendingToolCall.value = null;

    if (activeSshSessionId) {
      await createBackendConversation(activeSshSessionId);
      persistCurrentSession();
    }
  }

  function removeSession(id) {
    deleteStoredSession(id);
    sessions.value = getSessions();
    if (currentSessionId.value === id) {
      currentSessionId.value = null;
      messages.value = [];
      conversationId.value = null;
    }
  }

  let streamReader = null;

  function abortStream() {
    if (streamReader) {
      streamReader.cancel().catch(() => {});
      streamReader = null;
    }
  }

  async function stopChat() {
    if (!streaming.value || !conversationId.value) return;
    try {
      await requestWithAuth("/webssh/llm/stop", "POST", {
        body: { conversationId: conversationId.value },
      });
    } catch {
      abortStream();
    }
  }

  async function streamChat(body, assistantIdx) {
    let resp = await fetchWithAuth("/webssh/llm/chat", {
      method: "POST",
      body: JSON.stringify(body),
    });

    if (resp.status === 404 && activeSshSessionId) {
      await createBackendConversation(activeSshSessionId);
      body = { ...body, conversationId: conversationId.value };
      resp = await fetchWithAuth("/webssh/llm/chat", {
        method: "POST",
        body: JSON.stringify(body),
      });
    }

    if (!resp.ok) {
      throw new Error(`Chat request failed: ${resp.status}`);
    }

    let thinkingText = "";
    let tokenText = "";

    streamReader = resp.body.getReader();
    await parseSSEStream(streamReader, (event, payload) => {
      const inner = payload.payload || {};
      switch (event) {
        case "thinking":
          if (inner.text) {
            thinkingText += inner.text;
            updateAssistant(assistantIdx, {
              thinkingContent: thinkingText,
              content: tokenText,
            });
          }
          break;
        case "token":
          thinking.value = false;
          if (inner.text) {
            tokenText += inner.text;
            updateAssistant(assistantIdx, {
              thinkingContent: thinkingText,
              content: tokenText,
            });
          }
          break;
        case "error":
          updateAssistant(assistantIdx, {
            content: inner.message || "Unknown error",
            isError: true,
          });
          break;
        case "awaiting_approval":
          awaitingApproval.value = true;
          pendingToolCall.value = inner;
          break;
        case "stopped":
          break;
      }
    });
    streamReader = null;
  }

  async function sendMessage(text) {
    if (!conversationId.value) {
      throw new Error("No active conversation");
    }

    const history = messages.value.map((m) => ({
      role: m.role,
      content: m.content,
    }));
    addMessage("user", text);
    streaming.value = true;
    thinking.value = true;
    awaitingApproval.value = false;
    pendingToolCall.value = null;

    messages.value.push({
      role: "assistant",
      content: "",
      thinkingContent: "",
      timestamp: Date.now(),
    });
    const assistantIdx = messages.value.length - 1;

    try {
      await streamChat(
        {
          conversationId: conversationId.value,
          message: text,
          messages: history,
        },
        assistantIdx
      );
    } finally {
      streaming.value = false;
      thinking.value = false;
      persistCurrentSession();
    }
  }

  async function sendApproval(granted) {
    if (!conversationId.value) return;

    addMessage("user", granted ? "[approved]" : "[rejected]");
    const history = messages.value.map((m) => ({
      role: m.role,
      content: m.content,
    }));

    streaming.value = true;
    awaitingApproval.value = false;

    messages.value.push({
      role: "assistant",
      content: "",
      thinkingContent: "",
      timestamp: Date.now(),
    });
    const assistantIdx = messages.value.length - 1;

    try {
      await streamChat(
        {
          conversationId: conversationId.value,
          approval_granted: granted,
          messages: history,
        },
        assistantIdx
      );
    } finally {
      streaming.value = false;
      persistCurrentSession();
    }
  }

  return {
    currentSessionId: readonly(currentSessionId),
    conversationId: readonly(conversationId),
    messages,
    streaming: readonly(streaming),
    thinking: readonly(thinking),
    awaitingApproval: readonly(awaitingApproval),
    pendingToolCall: readonly(pendingToolCall),
    sessions: readonly(sessions),
    newSession,
    switchSession,
    removeSession,
    refreshSessions,
    sendMessage,
    sendApproval,
    stopChat,
  };
}
