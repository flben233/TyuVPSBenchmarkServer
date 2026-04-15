export function useWebSSH() {
  const config = useAppConfig();
  const { token } = useAuth();

  const status = ref("disconnected");
  const errorMessage = ref("");
  const sshSessionId = ref("");
  let ws = null;
  let pingInterval = null;

  let outputCallback = null;

  function onOutput(cb) {
    outputCallback = cb;
  }

  function handleOutput(data) {
    if (outputCallback) {
      const bytes = new Uint8Array(data);
      const decoder = new TextDecoder('utf-8');
      let decoded = decoder.decode(bytes);
      outputCallback(decoded);
    }
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
    doConnect(host, port, username, password, privateKey, cols, rows, false);
  }

  function doConnect(host, port, username, password, privateKey, cols, rows, isRetry) {
    if (ws) {
      ws.close();
    }

    status.value = "connecting";
    errorMessage.value = "";

    const wsUrl = getWebSocketUrl();
    const currentWs = new WebSocket(wsUrl);
    currentWs.binaryType = "arraybuffer";
    ws = currentWs;

    let opened = false;

    currentWs.onopen = () => {
      if (ws !== currentWs) return;
      opened = true;
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
      const msg = JSON.parse(event.data);
      switch (msg.type) {
        case "connected":
          status.value = "connected";
          errorMessage.value = "";
          sshSessionId.value = msg.message;
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
          status.value = "disconnected";
          stopPing();
          if (ws === currentWs) {
            ws = null;
          }
          break;
      }
    };

    currentWs.onerror = async (e) => {
      if (ws !== currentWs) return;
      if (!opened && !isRetry) {
        try {
          const { refreshToken } = useAuth();
          await refreshToken();
          doConnect(host, port, username, password, privateKey, cols, rows, true);
          return;
        } catch {
          // refresh failed, fall through to error
        }
      }
      errorMessage.value = "WebSocket connection failed";
      status.value = "error";
    };

    currentWs.onclose = () => {
      if (ws !== currentWs) return;
      if (status.value !== "error") {
        status.value = "disconnected";
      }
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

  function disconnect() {
    stopPing();
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
    sshSessionId: readonly(sshSessionId),
    connect,
    disconnect,
    sendInput,
    resize,
    onOutput,
  };
}
