import test from "node:test";
import assert from "node:assert/strict";

import {
  createAgentTaskMessage,
  createAgentMessage,
  createAgentApprovalResponse,
  isAgentUpdateMessage,
  isAgentApprovalMessage,
  isAgentErrorMessage,
  isAgentDoneMessage,
  parseAgentUpdateMessage,
  parseAgentApprovalMessage,
  parseAgentErrorMessage,
  parseAgentDoneMessage,
} from "../webssh-agent-protocol.js";

test("createAgentTaskMessage builds outbound payload", () => {
  assert.deepEqual(createAgentTaskMessage("check disk usage"), {
    type: "agent_task",
    message: "check disk usage",
  });
});

test("createAgentMessage builds outbound payload", () => {
  assert.deepEqual(createAgentMessage("task-1", "continue with apt fix"), {
    type: "agent_message",
    task_id: "task-1",
    message: "continue with apt fix",
  });
});

test("createAgentApprovalResponse builds outbound payload", () => {
  assert.deepEqual(createAgentApprovalResponse("task-1", true), {
    type: "agent_approval_response",
    task_id: "task-1",
    approved: true,
  });
});

test("agent inbound guards identify valid payloads", () => {
  assert.equal(isAgentUpdateMessage({ type: "agent_update", task_id: "t1", message: "step 1" }), true);
  assert.equal(isAgentApprovalMessage({ type: "agent_approval", task_id: "t1", question: "apply reboot?" }), true);
  assert.equal(isAgentApprovalMessage({ type: "agent_approval", task_id: "t1", message: "approval required" }), true);
  assert.equal(isAgentErrorMessage({ type: "agent_error", task_id: "t1", message: "permission denied" }), true);
  assert.equal(isAgentDoneMessage({ type: "agent_done", task_id: "t1", summary: "done" }), true);
  assert.equal(isAgentDoneMessage({ type: "agent_done", task_id: "t1", message: "done" }), true);
  assert.equal(isAgentErrorMessage({ type: "agent_error", message: "permission denied" }), true);
});

test("agent inbound parsers return normalized payload", () => {
  assert.deepEqual(parseAgentUpdateMessage('{"type":"agent_update","task_id":"t1","message":"running"}'), {
    type: "agent_update",
    taskId: "t1",
    message: "running",
    raw: {
      type: "agent_update",
      task_id: "t1",
      message: "running",
    },
  });

  assert.deepEqual(parseAgentApprovalMessage({ type: "agent_approval", task_id: "t2", question: "Proceed?" }), {
    type: "agent_approval",
    taskId: "t2",
    question: "Proceed?",
    raw: {
      type: "agent_approval",
      task_id: "t2",
      question: "Proceed?",
    },
  });

  assert.deepEqual(parseAgentErrorMessage({ type: "agent_error", task_id: "t2", message: "failed" }), {
    type: "agent_error",
    taskId: "t2",
    message: "failed",
    raw: {
      type: "agent_error",
      task_id: "t2",
      message: "failed",
    },
  });

  assert.deepEqual(parseAgentErrorMessage({ type: "agent_error", message: "failed" }), {
    type: "agent_error",
    taskId: "",
    message: "failed",
    raw: {
      type: "agent_error",
      message: "failed",
    },
  });

  assert.deepEqual(parseAgentDoneMessage({ type: "agent_done", task_id: "t2", summary: "all good" }), {
    type: "agent_done",
    taskId: "t2",
    summary: "all good",
    raw: {
      type: "agent_done",
      task_id: "t2",
      summary: "all good",
    },
  });

  assert.deepEqual(parseAgentApprovalMessage({ type: "agent_approval", task_id: "t2", message: "Proceed?" }), {
    type: "agent_approval",
    taskId: "t2",
    question: "Proceed?",
    raw: {
      type: "agent_approval",
      task_id: "t2",
      message: "Proceed?",
    },
  });

  assert.deepEqual(parseAgentDoneMessage({ type: "agent_done", task_id: "t2", message: "all good" }), {
    type: "agent_done",
    taskId: "t2",
    summary: "all good",
    raw: {
      type: "agent_done",
      task_id: "t2",
      message: "all good",
    },
  });
});

test("agent parsers return null on invalid payload", () => {
  assert.equal(parseAgentUpdateMessage("not-json"), null);
  assert.equal(parseAgentApprovalMessage({ type: "agent_approval", message: "missing task" }), null);
  assert.equal(parseAgentErrorMessage({ type: "agent_error", task_id: "t1" }), null);
  assert.equal(parseAgentDoneMessage({ type: "agent_done" }), null);
});
