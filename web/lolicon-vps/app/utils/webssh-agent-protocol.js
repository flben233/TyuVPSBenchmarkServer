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

export function isAgentUpdateMessage(payload) {
  return (
    isRecord(payload) &&
    payload.type === "agent_update" &&
    typeof payload.task_id === "string" &&
    typeof payload.message === "string"
  );
}

export function isAgentApprovalMessage(payload) {
  const hasQuestion = typeof payload?.question === "string";
  const hasMessageFallback = typeof payload?.message === "string";
  return (
    isRecord(payload) &&
    payload.type === "agent_approval" &&
    typeof payload.task_id === "string" &&
    (hasQuestion || hasMessageFallback)
  );
}

export function isAgentErrorMessage(payload) {
  return (
    isRecord(payload) &&
    payload.type === "agent_error" &&
    (typeof payload.task_id === "string" || typeof payload.task_id === "undefined") &&
    typeof payload.message === "string"
  );
}

export function isAgentDoneMessage(payload) {
  const hasSummary = typeof payload?.summary === "string";
  const hasMessageFallback = typeof payload?.message === "string";
  return (
    isRecord(payload) &&
    payload.type === "agent_done" &&
    typeof payload.task_id === "string" &&
    (hasSummary || hasMessageFallback)
  );
}

export function parseAgentUpdateMessage(input) {
  const payload = parseJsonMaybe(input);
  if (!isAgentUpdateMessage(payload)) {
    return null;
  }
  return {
    type: payload.type,
    taskId: payload.task_id,
    message: payload.message,
    raw: payload,
  };
}

export function parseAgentApprovalMessage(input) {
  const payload = parseJsonMaybe(input);
  if (!isAgentApprovalMessage(payload)) {
    return null;
  }
  const question = typeof payload.question === "string" ? payload.question : payload.message;
  return {
    type: payload.type,
    taskId: payload.task_id,
    question,
    raw: payload,
  };
}

export function parseAgentErrorMessage(input) {
  const payload = parseJsonMaybe(input);
  if (!isAgentErrorMessage(payload)) {
    return null;
  }
  return {
    type: payload.type,
    taskId: typeof payload.task_id === "string" ? payload.task_id : "",
    message: payload.message,
    raw: payload,
  };
}

export function parseAgentDoneMessage(input) {
  const payload = parseJsonMaybe(input);
  if (!isAgentDoneMessage(payload)) {
    return null;
  }
  const summary = typeof payload.summary === "string" ? payload.summary : payload.message;
  return {
    type: payload.type,
    taskId: payload.task_id,
    summary,
    raw: payload,
  };
}
