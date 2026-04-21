import { requestWithAuth, fetchWithAuth } from "~/composables/useAuth.js";
import { getSessions, saveSession, deleteSession as deleteStoredSession, createSessionId } from "~/utils/webssh-agent-session.js";
import { getLLMSettings, saveLLMSettings } from "~/utils/webssh-llm-settings.js";

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
  const contextMessages = ref([]);
  const streaming = ref(false);
  const thinking = ref(false);
  const awaitingApproval = ref(false);
  const pendingToolCall = ref(null);
  const sessions = ref({});
  const llmSettings = ref(getLLMSettings());
  const permanentAllowedCommands = ref([]);
  const sessionAllowedCommands = ref([]);

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
      contextMessages: contextMessages.value.map((m) => ({
        role: m.role,
        content: m.content,
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

  function addContextMessage(role, content) {
    contextMessages.value.push({ role, content });
  }

  function updateAssistant(idx, extras) {
    const msg = { ...messages.value[idx], timestamp: Date.now(), ...extras };
    messages.value[idx] = msg;
  }

  function updateContextAssistant(content) {
    for (let i = contextMessages.value.length - 1; i >= 0; i -= 1) {
      if (contextMessages.value[i].role === "assistant") {
        contextMessages.value[i] = {
          ...contextMessages.value[i],
          content,
        };
        return;
      }
    }
  }

  async function fetchWhitelist() {
    try {
      const resp = await requestWithAuth("/webssh/llm/whitelist/get", "GET");
      if (resp && resp.data && Array.isArray(resp.data.commands)) {
        permanentAllowedCommands.value = resp.data.commands;
      }
    } catch {
      // ignore – defaults will be used
    }
  }

  async function saveWhitelist(commands) {
    await requestWithAuth("/webssh/llm/whitelist/save", "POST", {
      body: { commands },
    });
    permanentAllowedCommands.value = commands;
  }

  function allowCommandsForSession(commands) {
    const current = new Set(sessionAllowedCommands.value);
    for (const cmd of commands) {
      current.add(cmd);
    }
    sessionAllowedCommands.value = [...current];
  }

  function _buildWhitelistBody() {
    return {
      allowed_commands: permanentAllowedCommands.value,
      session_allowed_commands: sessionAllowedCommands.value,
    };
  }

  async function createBackendConversation(sshSessionId) {
    activeSshSessionId = sshSessionId;
    const custom = llmSettings.value;
    const requestBody = { sshSessionId };
    if (custom.enabled) {
      if (!custom.apiBase || !custom.apiKey || !custom.model) {
        throw new Error("Custom LLM API requires API Base, API Key and Model");
      }
      requestBody.llmApi = {
        apiBase: custom.apiBase,
        apiKey: custom.apiKey,
        model: custom.model,
      };
    }
    const resp = await requestWithAuth("/webssh/llm/new", "POST", {
      body: requestBody,
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
    contextMessages.value = [];
    conversationId.value = null;
    awaitingApproval.value = false;
    pendingToolCall.value = null;
    sessionAllowedCommands.value = [];
    activeSshSessionId = sshSessionId;

    await fetchWhitelist();
    await createBackendConversation(sshSessionId);

    persistCurrentSession();
  }

  async function switchSession(id, sshSessionId) {
    persistCurrentSession();
    const session = sessions.value[id];
    if (!session) return;

    currentSessionId.value = id;
    messages.value = session.messages.map((m) => ({ ...m }));
    contextMessages.value = (session.contextMessages || session.messages || []).map((m) => ({
      role: m.role,
      content: m.content,
    }));
    conversationId.value = null;
    activeSshSessionId = sshSessionId || session.sshSessionId;
    awaitingApproval.value = false;
    pendingToolCall.value = null;
    sessionAllowedCommands.value = [];

    if (activeSshSessionId) {
      await fetchWhitelist();
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
      contextMessages.value = [];
      conversationId.value = null;
    }
  }

  function updateLLMSettings(nextSettings) {
    saveLLMSettings(nextSettings);
    llmSettings.value = getLLMSettings();
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
            updateContextAssistant(tokenText);
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
            updateContextAssistant(tokenText);
          }
          break;
        case "error":
          updateAssistant(assistantIdx, {
            content: inner.message || "Unknown error",
            isError: true,
          });
          updateContextAssistant(inner.message || "Unknown error");
          break;
        case "awaiting_approval":
          awaitingApproval.value = true;
          pendingToolCall.value = inner;
          break;
        case "done":
          console.log(inner)
          if (Array.isArray(inner.contextMessages) && inner.contextMessages.length > 0) {
            contextMessages.value = inner.contextMessages.map((m) => ({
              role: m.role,
              content: m.content,
            }));
          }
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

    const history = contextMessages.value.map((m) => ({
      role: m.role,
      content: m.content,
    }));
    addMessage("user", text);
    addContextMessage("user", text);
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
    addContextMessage("assistant", "");
    const assistantIdx = messages.value.length - 1;

    try {
      await streamChat(
        {
          conversationId: conversationId.value,
          message: text,
          messages: history,
          ..._buildWhitelistBody(),
        },
        assistantIdx
      );
    } finally {
      streaming.value = false;
      thinking.value = false;
      persistCurrentSession();
    }
  }

  async function sendApproval(granted, addToSession = false) {
    if (!conversationId.value) return;

    if (addToSession && pendingToolCall.value?.disallowed_commands) {
      allowCommandsForSession(pendingToolCall.value.disallowed_commands);
    }

    addMessage("user", granted ? "[approved]" : "[rejected]");
    addContextMessage("user", granted ? "[approved]" : "[rejected]");
    const history = contextMessages.value.map((m) => ({
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
    addContextMessage("assistant", "");
    const assistantIdx = messages.value.length - 1;

    try {
      await streamChat(
        {
          conversationId: conversationId.value,
          approval_granted: granted,
          messages: history,
          ..._buildWhitelistBody(),
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
    contextMessages: readonly(contextMessages),
    streaming: readonly(streaming),
    thinking: readonly(thinking),
    awaitingApproval: readonly(awaitingApproval),
    pendingToolCall: readonly(pendingToolCall),
    sessions: readonly(sessions),
    llmSettings: readonly(llmSettings),
    permanentAllowedCommands: readonly(permanentAllowedCommands),
    sessionAllowedCommands: readonly(sessionAllowedCommands),
    newSession,
    switchSession,
    removeSession,
    updateLLMSettings,
    fetchWhitelist,
    saveWhitelist,
    allowCommandsForSession,
    refreshSessions,
    sendMessage,
    sendApproval,
    stopChat,
  };
}
