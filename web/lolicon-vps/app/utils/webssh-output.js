function asUint8Array(payload) {
  if (payload instanceof Uint8Array) {
    return payload;
  }

  if (payload instanceof ArrayBuffer) {
    return new Uint8Array(payload);
  }

  if (typeof Buffer !== "undefined" && Buffer.isBuffer(payload)) {
    return new Uint8Array(payload.buffer, payload.byteOffset, payload.byteLength);
  }

  if (
    payload &&
    typeof payload === "object" &&
    payload.type === "Buffer" &&
    Array.isArray(payload.data)
  ) {
    return new Uint8Array(payload.data);
  }

  return null;
}

export function createStreamingUtf8Decoder() {
  const decoder = new TextDecoder("utf-8");

  return {
    decode(chunk) {
      const bytes = asUint8Array(chunk);
      if (!bytes) {
        return typeof chunk === "string" ? chunk : "";
      }
      return decoder.decode(bytes, { stream: true });
    },
    flush() {
      return decoder.decode();
    },
  };
}

export function normalizeWebSSHOutputPayload(payload, decodeState = null) {
  if (typeof payload === "string") {
    return payload;
  }

  const bytes = asUint8Array(payload);
  if (!bytes) {
    return "";
  }

  if (decodeState && typeof decodeState.decode === "function") {
    return decodeState.decode(bytes);
  }

  return new TextDecoder("utf-8").decode(bytes);
}

export function extractOutputPayload(msg) {
  if (!msg || typeof msg !== "object") {
    return null;
  }

  const type = typeof msg.type === "string" ? msg.type : "";
  if (!type) {
    return null;
  }

  const outputTypes = new Set(["output", "stdout", "terminal_output"]);
  if (!outputTypes.has(type)) {
    return null;
  }

  if (Object.prototype.hasOwnProperty.call(msg, "data")) {
    return msg.data;
  }
  if (Object.prototype.hasOwnProperty.call(msg, "output")) {
    return msg.output;
  }
  if (Object.prototype.hasOwnProperty.call(msg, "payload")) {
    return msg.payload;
  }

  return null;
}
