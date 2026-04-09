import test from "node:test";
import assert from "node:assert/strict";

import * as protocol from "../webssh-agent-protocol.js";

const {
  createAgentTaskMessage,
  createAgentMessage,
  createAgentApprovalResponse,
  parseAgentMessageStart,
  parseAgentToken,
  parseAgentMessageEnd,
  parseAgentState,
} = protocol;

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

test("parses new streaming events", () => {
  assert.deepEqual(parseAgentMessageStart({ type: "agent_message_start", task_id: "t1", message_id: "m1" }), {
    taskId: "t1",
    messageId: "m1",
  });

  assert.deepEqual(parseAgentToken({ type: "agent_token", task_id: "t1", message_id: "m1", delta: "he" }), {
    taskId: "t1",
    messageId: "m1",
    delta: "he",
  });

  assert.deepEqual(parseAgentMessageEnd({ type: "agent_message_end", task_id: "t1", message_id: "m1", finish_reason: "stop" }), {
    taskId: "t1",
    messageId: "m1",
    finishReason: "stop",
  });

  assert.deepEqual(parseAgentState({ type: "agent_state", task_id: "t1", state: "thinking", message: "planning" }), {
    taskId: "t1",
    state: "thinking",
    message: "planning",
  });
});

test("does not treat legacy events as stream events", () => {
  const legacyEventPayloads = [
    { type: "agent_update", task_id: "t1", message: "status" },
    { type: "agent_approval", task_id: "t1", message: "approve?" },
    { type: "agent_error", task_id: "t1", message: "boom" },
    { type: "agent_done", task_id: "t1", message: "done" },
  ];

  for (const payload of legacyEventPayloads) {
    assert.equal(parseAgentMessageStart(payload), null);
    assert.equal(parseAgentToken(payload), null);
    assert.equal(parseAgentMessageEnd(payload), null);
    assert.equal(parseAgentState(payload), null);
  }
});

test("applies default values for optional fields", () => {
  assert.deepEqual(parseAgentMessageEnd({ type: "agent_message_end", task_id: "t1", message_id: "m1" }), {
    taskId: "t1",
    messageId: "m1",
    finishReason: "stop",
  });

  assert.deepEqual(parseAgentState({ type: "agent_state", task_id: "t1", state: "thinking" }), {
    taskId: "t1",
    state: "thinking",
    message: "",
  });
});

test("returns null for malformed or invalid streaming payloads", () => {
  assert.equal(parseAgentMessageStart("not-json"), null);
  assert.equal(parseAgentMessageStart({ type: "agent_message_start", task_id: "   ", message_id: "m1" }), null);
  assert.equal(parseAgentMessageStart({ type: "agent_message_start", task_id: "t1", message_id: "   " }), null);

  assert.equal(parseAgentToken({ type: "agent_token", task_id: "t1", message_id: "m1", delta: 12 }), null);
  assert.equal(parseAgentToken({ type: "agent_token", task_id: "   ", message_id: "m1", delta: "x" }), null);
  assert.equal(parseAgentToken({ type: "agent_token", task_id: "t1", message_id: "   ", delta: "x" }), null);

  assert.equal(parseAgentMessageEnd({ type: "agent_message_end", task_id: "   ", message_id: "m1" }), null);
  assert.equal(parseAgentMessageEnd({ type: "agent_message_end", task_id: "t1", message_id: "   " }), null);
  assert.equal(parseAgentMessageEnd({ type: "agent_message_end", task_id: "t1", message_id: "m1", finish_reason: 0 }), null);

  assert.equal(parseAgentState({ type: "agent_state", task_id: "   ", state: "thinking" }), null);
  assert.equal(parseAgentState({ type: "agent_state", task_id: "t1", state: "unknown" }), null);
  assert.equal(parseAgentState({ type: "agent_state", task_id: "t1", state: 1 }), null);
  assert.equal(parseAgentState({ type: "agent_state", task_id: "t1", state: "done", message: 1 }), null);
});

test("does not export legacy parser symbols", () => {
  assert.equal("parseAgentUpdateMessage" in protocol, false);
  assert.equal("parseAgentApprovalMessage" in protocol, false);
  assert.equal("parseAgentErrorMessage" in protocol, false);
  assert.equal("parseAgentDoneMessage" in protocol, false);
});
