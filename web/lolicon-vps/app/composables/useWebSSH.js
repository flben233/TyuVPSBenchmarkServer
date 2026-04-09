import {
  createAgentTaskMessage,
  createAgentMessage,
  createAgentApprovalResponse,
  parseAgentMessageStart,
  parseAgentToken,
  parseAgentMessageEnd,
  parseAgentState,
} from "~/utils/webssh-agent-protocol";
import {
  createStreamingUtf8Decoder,
  normalizeWebSSHOutputPayload,
  extractOutputPayload,
} from "~/utils/webssh-output";

export function useWebSSH() {
  const config = useAppConfig();
  const { token } = useAuth();

  const status = ref("disconnected");
  const errorMessage = ref("");
  let ws = null;
  let pingInterval = null;
  let outputDecoder = createStreamingUtf8Decoder();

  let outputCallback = null;
  let agentMessageStartCallback = null;
  let agentTokenCallback = null;
  let agentMessageEndCallback = null;
  let agentStateCallback = null;

  function onOutput(cb) {
    outputCallback = cb;
  }

  function onAgentMessageStart(cb) {
    agentMessageStartCallback = cb;
  }

  function onAgentToken(cb) {
    agentTokenCallback = cb;
  }

  function onAgentMessageEnd(cb) {
    agentMessageEndCallback = cb;
  }

  function onAgentState(cb) {
    agentStateCallback = cb;
  }

  function handleOutput(data) {
    if (!outputCallback) {
      return;
    }
    const decoded = normalizeWebSSHOutputPayload(data, outputDecoder);
    if (decoded) {
      outputCallback(decoded);
    }
  }

  function flushOutputDecoder() {
    if (!outputCallback) {
      outputDecoder = createStreamingUtf8Decoder();
      return;
    }
    const tail = outputDecoder.flush();
    if (tail) {
      outputCallback(tail);
    }
    outputDecoder = createStreamingUtf8Decoder();
  }

  function sendJsonMessage(payload) {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      return false;
    }
    ws.send(JSON.stringify(payload));
    return true;
  }

  function handleAgentEventMessage(msg) {
    const messageStart = parseAgentMessageStart(msg);
    if (messageStart) {
      if (agentMessageStartCallback) {
        agentMessageStartCallback(messageStart);
      }
      return true;
    }

    const token = parseAgentToken(msg);
    if (token) {
      if (agentTokenCallback) {
        agentTokenCallback(token);
      }
      return true;
    }

    const messageEnd = parseAgentMessageEnd(msg);
    if (messageEnd) {
      if (agentMessageEndCallback) {
        agentMessageEndCallback(messageEnd);
      }
      return true;
    }

    const state = parseAgentState(msg);
    if (state) {
      if (agentStateCallback) {
        agentStateCallback(state);
      }
      return true;
    }

    return false;
  }

  function getWebSocketUrl() {
    if (config.backendWsUrl) {
      return `${config.backendWsUrl}/webssh/ws?token=${token.value}`;
    }
    const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const hostname = window.location.hostname;
    const port = window.location.port || (window.location.protocol === "https:" ? "443" : "80");
    return `${wsProtocol}//${hostname}:${port}/api/webssh/ws?token=${token.value}`;
  }

  function connect(host, port, username, password, privateKey, cols, rows) {
    if (ws) {
      ws.close();
    }

    status.value = "connecting";
    errorMessage.value = "";

    const wsUrl = getWebSocketUrl();
    const currentWs = new WebSocket(wsUrl);
    currentWs.binaryType = "arraybuffer";
    ws = currentWs;

    currentWs.onopen = () => {
      if (ws !== currentWs) return;
      const connectMsg = {
        type: "connect",
        host,
        port,
        username,
        password,
        private_key: privateKey,
        cols,
        rows,
      };
      currentWs.send(JSON.stringify(connectMsg));
    };

    currentWs.onmessage = (event) => {
      if (ws !== currentWs) return;
      if (typeof event.data !== "string") {
        if (status.value === "connected") {
          handleOutput(event.data);
        }
        return;
      }

      let msg;
      try {
        msg = JSON.parse(event.data);
      } catch {
        if (status.value === "connected") {
          handleOutput(event.data);
        }
        return;
      }

      if (handleAgentEventMessage(msg)) {
        return;
      }

      const outputPayload = extractOutputPayload(msg);
      if (status.value === "connected" && outputPayload !== null) {
        handleOutput(outputPayload);
        return;
      }

      switch (msg.type) {
        case "connected":
          status.value = "connected";
          errorMessage.value = "";
          startPing();
          break;
        case "output":
          if (status.value === "connected") {
            handleOutput(msg.data);
          }
          break;
        case "error":
          errorMessage.value = msg.message;
          if (status.value === "connecting") {
            status.value = "error";
            currentWs.close();
            ws = null;
          }
          break;
        case "closed":
          flushOutputDecoder();
          status.value = "disconnected";
          stopPing();
          if (ws === currentWs) {
            ws = null;
          }
          break;
      }
    };

    currentWs.onerror = () => {
      if (ws !== currentWs) return;
      errorMessage.value = "WebSocket connection failed";
      status.value = "error";
    };

    currentWs.onclose = () => {
      if (ws !== currentWs) return;
      if (status.value !== "error") {
        status.value = "disconnected";
      }
      flushOutputDecoder();
      stopPing();
      ws = null;
    };
  }

  function sendInput(data) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(new TextEncoder().encode(data));
    }
  }

  function resize(cols, rows) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: "resize", cols, rows }));
    }
  }

  function sendAgentTask(message) {
    return sendJsonMessage(createAgentTaskMessage(message));
  }

  function sendAgentMessage(taskId, message) {
    return sendJsonMessage(createAgentMessage(taskId, message));
  }

  function sendAgentApproval(taskId, approved) {
    return sendJsonMessage(createAgentApprovalResponse(taskId, approved));
  }

  function disconnect() {
    stopPing();
    flushOutputDecoder();
    if (ws) {
      ws.close();
      ws = null;
    }
    status.value = "disconnected";
    errorMessage.value = "";
  }

  function startPing() {
    stopPing();
    pingInterval = setInterval(() => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "ping" }));
      }
    }, 30000);
  }

  function stopPing() {
    if (pingInterval) {
      clearInterval(pingInterval);
      pingInterval = null;
    }
  }

  onUnmounted(() => {
    disconnect();
  });

  return {
    status: readonly(status),
    errorMessage: readonly(errorMessage),
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
  };
}
