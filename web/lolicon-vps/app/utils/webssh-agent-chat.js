export function shouldDedupeAgentState(previous, next, now, windowMs = 1200) {
  if (!previous || !next) {
    return false;
  }
  if (previous.taskId !== next.taskId) {
    return false;
  }
  if (previous.state !== next.state) {
    return false;
  }
  if (previous.message !== next.message) {
    return false;
  }
  return now - previous.timestamp <= windowMs;
}

export function setLatestPendingApproval(pendingByTask, taskId, question, messageId) {
  return {
    ...pendingByTask,
    [taskId]: {
      taskId,
      question,
      messageId,
    },
  };
}

export function clearPendingApprovalForTask(pendingByTask, taskId) {
  if (!taskId || !pendingByTask[taskId]) {
    return pendingByTask;
  }
  const next = { ...pendingByTask };
  delete next[taskId];
  return next;
}

export function buildApprovalsByMessageId(messages, pendingByTask) {
  const mapping = {};
  for (const message of messages) {
    if (!message?.messageId || message.kind !== "approval") {
      continue;
    }
    const pending = pendingByTask[message.taskId];
    if (pending && pending.messageId === message.messageId) {
      mapping[message.messageId] = pending;
    }
  }
  return mapping;
}
