import test from "node:test";
import assert from "node:assert/strict";

import {
  shouldDedupeAgentState,
  setLatestPendingApproval,
  clearPendingApprovalForTask,
  buildApprovalsByMessageId,
} from "../webssh-agent-chat.js";
import {
  createStreamingUtf8Decoder,
  normalizeWebSSHOutputPayload,
  extractOutputPayload,
} from "../webssh-output.js";

test("dedupes repeated agent state messages in short window", () => {
  const previous = {
    taskId: "task-1",
    state: "thinking",
    message: "planning",
    timestamp: 10_000,
  };

  assert.equal(
    shouldDedupeAgentState(
      previous,
      { taskId: "task-1", state: "thinking", message: "planning" },
      10_900,
      1_000
    ),
    true
  );

  assert.equal(
    shouldDedupeAgentState(
      previous,
      { taskId: "task-1", state: "running_command", message: "planning" },
      10_900,
      1_000
    ),
    false
  );

  assert.equal(
    shouldDedupeAgentState(
      previous,
      { taskId: "task-1", state: "thinking", message: "planning" },
      11_500,
      1_000
    ),
    false
  );
});

test("streaming decoder correctly handles split multibyte utf-8", () => {
  const decoder = createStreamingUtf8Decoder();
  const encoder = new TextEncoder();
  const bytes = encoder.encode("A你B");

  const firstChunk = bytes.slice(0, 2);
  const secondChunk = bytes.slice(2);

  assert.equal(decoder.decode(firstChunk), "A");
  assert.equal(decoder.decode(secondChunk), "你B");
  assert.equal(decoder.flush(), "");
});

test("normalizeWebSSHOutputPayload supports payload variants", () => {
  assert.equal(normalizeWebSSHOutputPayload("plain"), "plain");
  assert.equal(normalizeWebSSHOutputPayload(new Uint8Array([65, 66, 67])), "ABC");
  assert.equal(normalizeWebSSHOutputPayload(new Uint8Array([228, 189, 160]).buffer), "你");
  assert.equal(normalizeWebSSHOutputPayload({ type: "Buffer", data: [65, 10] }), "A\n");
});

test("extractOutputPayload handles output message variants", () => {
  assert.equal(extractOutputPayload({ type: "output", data: "x" }), "x");
  assert.equal(extractOutputPayload({ type: "output", output: "y" }), "y");
  assert.equal(extractOutputPayload({ type: "stdout", data: "z" }), "z");
  assert.equal(extractOutputPayload({ type: "terminal_output", payload: "k" }), "k");
  assert.equal(extractOutputPayload({ type: "connected" }), null);
});

test("two sequential approvals for same task only keep latest actionable", () => {
  const messages = [
    { messageId: "m-old", taskId: "task-1", kind: "approval" },
    { messageId: "m-new", taskId: "task-1", kind: "approval" },
  ];

  let pendingByTask = {};
  pendingByTask = setLatestPendingApproval(pendingByTask, "task-1", "first", "m-old");
  pendingByTask = setLatestPendingApproval(pendingByTask, "task-1", "second", "m-new");

  assert.deepEqual(buildApprovalsByMessageId(messages, pendingByTask), {
    "m-new": {
      taskId: "task-1",
      question: "second",
      messageId: "m-new",
    },
  });
});

test("approval is cleared when task enters error state", () => {
  const messages = [{ messageId: "m-1", taskId: "task-1", kind: "approval" }];
  let pendingByTask = setLatestPendingApproval({}, "task-1", "need approval", "m-1");

  pendingByTask = clearPendingApprovalForTask(pendingByTask, "task-1");

  assert.deepEqual(buildApprovalsByMessageId(messages, pendingByTask), {});
});
