function isRecord(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function parseJsonMaybe(input) {
  if (typeof input === "string") {
    try {
      return JSON.parse(input);
    } catch {
      return null;
    }
  }
  return input;
}

const AGENT_STATES = new Set([
  "thinking",
  "running_command",
  "awaiting_approval",
  "done",
  "error",
]);

function isNonEmptyString(value) {
  return typeof value === "string" && value.trim().length > 0;
}

export function createAgentTaskMessage(message) {
  return {
    type: "agent_task",
    message,
  };
}

export function createAgentMessage(taskId, message) {
  return {
    type: "agent_message",
    task_id: taskId,
    message,
  };
}

export function createAgentApprovalResponse(taskId, approved) {
  return {
    type: "agent_approval_response",
    task_id: taskId,
    approved,
  };
}

function isAgentMessageStart(payload) {
  return (
    isRecord(payload) &&
    payload.type === "agent_message_start" &&
    isNonEmptyString(payload.task_id) &&
    isNonEmptyString(payload.message_id)
  );
}

function isAgentToken(payload) {
  return (
    isRecord(payload) &&
    payload.type === "agent_token" &&
    isNonEmptyString(payload.task_id) &&
    isNonEmptyString(payload.message_id) &&
    typeof payload.delta === "string"
  );
}

function isAgentMessageEnd(payload) {
  return (
    isRecord(payload) &&
    payload.type === "agent_message_end" &&
    isNonEmptyString(payload.task_id) &&
    isNonEmptyString(payload.message_id) &&
    (typeof payload.finish_reason === "string" || typeof payload.finish_reason === "undefined")
  );
}

function isAgentState(payload) {
  return (
    isRecord(payload) &&
    payload.type === "agent_state" &&
    isNonEmptyString(payload.task_id) &&
    typeof payload.state === "string" &&
    AGENT_STATES.has(payload.state) &&
    (typeof payload.message === "string" || typeof payload.message === "undefined")
  );
}

export function parseAgentMessageStart(input) {
  const payload = parseJsonMaybe(input);
  if (!isAgentMessageStart(payload)) {
    return null;
  }
  return {
    taskId: payload.task_id,
    messageId: payload.message_id,
  };
}

export function parseAgentToken(input) {
  const payload = parseJsonMaybe(input);
  if (!isAgentToken(payload)) {
    return null;
  }
  return {
    taskId: payload.task_id,
    messageId: payload.message_id,
    delta: payload.delta,
  };
}

export function parseAgentMessageEnd(input) {
  const payload = parseJsonMaybe(input);
  if (!isAgentMessageEnd(payload)) {
    return null;
  }
  return {
    taskId: payload.task_id,
    messageId: payload.message_id,
    finishReason: typeof payload.finish_reason === "string" ? payload.finish_reason : "stop",
  };
}

export function parseAgentState(input) {
  const payload = parseJsonMaybe(input);
  if (!isAgentState(payload)) {
    return null;
  }
  return {
    taskId: payload.task_id,
    state: payload.state,
    message: typeof payload.message === "string" ? payload.message : "",
  };
}
