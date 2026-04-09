from __future__ import annotations

from copy import deepcopy
from itertools import count
import json
from typing import Any, Callable, Protocol

from langgraph.graph import END, START, StateGraph
from redis import Redis

from .go_client import GoToolClient
from .state import AgentState, append_step, merge_with_defaults


DEFAULT_MAX_RETRIES = 3
CHECKPOINT_TTL_SECONDS = 30 * 60


class Checkpointer(Protocol):
    def put(self, key: str, state: AgentState) -> None: ...

    def get(self, task_id: str) -> AgentState | None: ...

    def get_by_key(self, key: str) -> AgentState | None: ...


class StreamEmitter(Protocol):
    def on_message_start(self, task_id: str, message_id: str) -> None: ...

    def on_token(self, task_id: str, message_id: str, delta: str) -> None: ...

    def on_message_end(self, task_id: str, message_id: str, finish_reason: str) -> None: ...

    def on_state(self, task_id: str, state: str, message: str) -> None: ...


class InMemoryCheckpointer:
    def __init__(self) -> None:
        self._store: dict[str, AgentState] = {}

    def put(self, key: str, state: AgentState) -> None:
        self._store[key] = deepcopy(state)

    def get(self, task_id: str) -> AgentState | None:
        return deepcopy(self._store.get(checkpoint_key(task_id)))

    def get_by_key(self, key: str) -> AgentState | None:
        return deepcopy(self._store.get(key))


class RedisCheckpointer:
    def __init__(self, redis_client: Redis) -> None:
        self._redis = redis_client

    def put(self, key: str, state: AgentState) -> None:
        self._redis.set(key, json.dumps(state), ex=CHECKPOINT_TTL_SECONDS)

    def get(self, task_id: str) -> AgentState | None:
        return self.get_by_key(checkpoint_key(task_id))

    def get_by_key(self, key: str) -> AgentState | None:
        raw = self._redis.get(key)
        if raw is None:
            return None
        if isinstance(raw, bytes):
            raw = raw.decode("utf-8")
        payload = json.loads(raw)
        if not isinstance(payload, dict):
            return None
        return merge_with_defaults(payload)


def checkpoint_key(task_id: str) -> str:
    return f"checkpoint:{task_id}"


def _save_checkpoint(checkpointer: Checkpointer, state: AgentState) -> None:
    task_id = state.get("task_id", "")
    if task_id:
        checkpointer.put(checkpoint_key(task_id), state)


def _default_command_executor(command: str) -> str:
    return f"simulated execution: {command}"


def _extract_data_payload(response_json: dict[str, Any], endpoint: str) -> dict[str, Any]:
    data = response_json.get("data")
    if not isinstance(data, dict):
        raise RuntimeError(f"go tool response missing data payload for {endpoint}")
    return data


def _invoke_safety_check(go_client: GoToolClient, task_id: str, command: str) -> dict[str, Any]:
    response = go_client.post_sync(
        "/api/agent/safety-check",
        {"command": command},
        task_id=task_id,
    )
    payload = response.json()
    if not isinstance(payload, dict):
        raise RuntimeError("go tool safety-check returned invalid payload")
    return _extract_data_payload(payload, "safety-check")


def _invoke_execute(
    go_client: GoToolClient,
    task_id: str,
    command: str,
    approved: bool,
) -> dict[str, Any]:
    response = go_client.post_sync(
        "/api/agent/execute",
        {
            "command": command,
            "approved": approved,
        },
        task_id=task_id,
    )
    payload = response.json()
    if not isinstance(payload, dict):
        raise RuntimeError("go tool execute returned invalid payload")
    return _extract_data_payload(payload, "execute")


def build_agent_graph(
    checkpointer: Checkpointer | None = None,
    go_client: GoToolClient | None = None,
    command_executor: Callable[[str], str] | None = None,
    max_retries: int = DEFAULT_MAX_RETRIES,
    stream_emitter: StreamEmitter | None = None,
):
    active_checkpointer = checkpointer or InMemoryCheckpointer()
    execute_command_impl = command_executor or _default_command_executor
    message_counter = count(1)

    def _emit_state(state: AgentState, status: str, message: str) -> None:
        if stream_emitter is None:
            return
        stream_emitter.on_state(state.get("task_id", ""), status, message)

    def _emit_message_events(state: AgentState, text: str, finish_reason: str = "stop") -> None:
        if stream_emitter is None:
            return
        task_id = state.get("task_id", "")
        message_id = f"msg-{next(message_counter)}"
        stream_emitter.on_message_start(task_id, message_id)
        if text:
            stream_emitter.on_token(task_id, message_id, text)
        stream_emitter.on_message_end(task_id, message_id, finish_reason)

    def _safe_node_call(state: AgentState, fn: Callable[[AgentState], AgentState]) -> AgentState:
        try:
            return fn(state)
        except Exception as exc:
            merged = merge_with_defaults(state)
            _emit_state(merged, "error", str(exc))
            _emit_message_events(merged, str(exc), finish_reason="error")
            raise

    def receive_task(state: AgentState) -> AgentState:
        next_state = merge_with_defaults(state)
        append_step(next_state, "receive_task")
        _save_checkpoint(active_checkpointer, next_state)
        return next_state

    def parse_intent(state: AgentState) -> AgentState:
        next_state = merge_with_defaults(state)
        _emit_state(next_state, "thinking", "Parsing user intent.")
        latest_message = next_state.get("messages", [""])[-1] if next_state.get("messages") else ""
        intent = latest_message.strip()
        if intent.lower().startswith("run:"):
            intent = intent[4:].strip()
        next_state["current_command"] = intent
        next_state["task_complete"] = False
        next_state["final_response"] = ""
        append_step(next_state, "parse_intent")
        _save_checkpoint(active_checkpointer, next_state)
        return next_state

    def generate_command(state: AgentState) -> AgentState:
        next_state = merge_with_defaults(state)
        if not next_state.get("current_command"):
            latest_message = next_state.get("messages", [""])[-1] if next_state.get("messages") else ""
            next_state["current_command"] = latest_message.strip()
        append_step(next_state, "generate_command")
        _save_checkpoint(active_checkpointer, next_state)
        return next_state

    def safety_check(state: AgentState) -> AgentState:
        next_state = merge_with_defaults(state)
        command = next_state.get("current_command", "")
        command_lower = command.lower()
        task_id = next_state.get("task_id", "")
        dangerous_tokens = ("rm -rf", "shutdown", "reboot", "mkfs", "dd if=")
        is_dangerous = any(token in command_lower for token in dangerous_tokens)
        requires_approval = is_dangerous

        if go_client is not None and task_id and command:
            safety_data = _invoke_safety_check(go_client, task_id, command)
            risk_level = str(safety_data.get("risk_level", "warning"))
            next_state["safety_status"] = risk_level
            requires_approval = bool(safety_data.get("requires_approval", risk_level != "safe"))
        else:
            next_state["safety_status"] = "dangerous" if is_dangerous else "safe"

        approved_for_this_command = (
            next_state.get("approval_granted", False)
            and next_state.get("approved_command", "")
            in ("", next_state.get("current_command", ""))
        )
        next_state["awaiting_approval"] = requires_approval and not approved_for_this_command
        append_step(next_state, "safety_check")
        _save_checkpoint(active_checkpointer, next_state)
        return next_state

    def await_approval(state: AgentState) -> AgentState:
        next_state = merge_with_defaults(state)
        next_state["task_complete"] = False
        next_state["final_response"] = "Command requires explicit approval."
        append_step(next_state, "await_approval")
        _save_checkpoint(active_checkpointer, next_state)
        return next_state

    def execute_command(state: AgentState) -> AgentState:
        next_state = merge_with_defaults(state)
        _emit_state(next_state, "running_command", "Executing command.")
        next_state["awaiting_approval"] = False
        command = next_state.get("current_command", "")
        task_id = next_state.get("task_id", "")
        approved_for_this_command = (
            next_state.get("approval_granted", False)
            and next_state.get("approved_command", "")
            in ("", command)
        )

        if go_client is not None and task_id and command:
            execute_data = _invoke_execute(go_client, task_id, command, approved_for_this_command)
            next_state["command_result"] = str(execute_data.get("message", ""))
            next_state["awaiting_approval"] = str(execute_data.get("status", "")) == "awaiting_approval"
        else:
            next_state["command_result"] = execute_command_impl(command)

        next_state["approval_granted"] = False
        next_state["approved_command"] = ""
        append_step(next_state, "execute_command")
        _save_checkpoint(active_checkpointer, next_state)
        return next_state

    def analyze_result(state: AgentState) -> AgentState:
        next_state = merge_with_defaults(state)
        if next_state.get("awaiting_approval", False):
            next_state["task_complete"] = False
            append_step(next_state, "analyze_result")
            _save_checkpoint(active_checkpointer, next_state)
            return next_state

        result = next_state.get("command_result", "")
        retry_requested = "retry" in result.lower()
        retry_count = int(next_state.get("retry_count", 0))

        if retry_requested:
            retry_count += 1
            next_state["retry_count"] = retry_count
            if retry_count >= max_retries:
                next_state["task_complete"] = True
                next_state["final_response"] = f"Retry limit exceeded ({max_retries})."
            else:
                next_state["task_complete"] = False
        else:
            next_state["retry_count"] = 0
            next_state["task_complete"] = True

        append_step(next_state, "analyze_result")
        _save_checkpoint(active_checkpointer, next_state)
        return next_state

    def report_result(state: AgentState) -> AgentState:
        next_state = merge_with_defaults(state)
        if next_state.get("awaiting_approval", False):
            next_state["final_response"] = "Awaiting user approval for dangerous command."
        elif next_state.get("task_complete", False):
            if not next_state.get("final_response"):
                next_state["final_response"] = f"Task complete: {next_state.get('command_result', '')}".strip()
        else:
            next_state["final_response"] = "Task needs another execution cycle."

        if next_state.get("awaiting_approval", False):
            _emit_state(next_state, "awaiting_approval", next_state.get("final_response", ""))
            _emit_message_events(next_state, next_state.get("final_response", ""), finish_reason="stop")
        elif next_state.get("task_complete", False):
            _emit_state(next_state, "done", next_state.get("final_response", ""))
            _emit_message_events(next_state, next_state.get("final_response", ""), finish_reason="stop")
        else:
            _emit_state(next_state, "thinking", next_state.get("final_response", ""))

        append_step(next_state, "report_result")
        _save_checkpoint(active_checkpointer, next_state)
        return next_state

    graph = StateGraph(AgentState)
    graph.add_node("receive_task", lambda state: _safe_node_call(state, receive_task))
    graph.add_node("parse_intent", lambda state: _safe_node_call(state, parse_intent))
    graph.add_node("generate_command", lambda state: _safe_node_call(state, generate_command))
    graph.add_node("safety_check", lambda state: _safe_node_call(state, safety_check))
    graph.add_node("await_approval", lambda state: _safe_node_call(state, await_approval))
    graph.add_node("execute_command", lambda state: _safe_node_call(state, execute_command))
    graph.add_node("analyze_result", lambda state: _safe_node_call(state, analyze_result))
    graph.add_node("report_result", lambda state: _safe_node_call(state, report_result))

    graph.add_edge(START, "receive_task")
    graph.add_edge("receive_task", "parse_intent")
    graph.add_edge("parse_intent", "generate_command")
    graph.add_edge("generate_command", "safety_check")
    graph.add_conditional_edges(
        "safety_check",
        lambda state: "await_approval" if state.get("awaiting_approval", False) else "execute_command",
        {"await_approval": "await_approval", "execute_command": "execute_command"},
    )
    graph.add_edge("await_approval", "report_result")
    graph.add_edge("execute_command", "analyze_result")
    graph.add_conditional_edges(
        "analyze_result",
        lambda state: "report_result" if state.get("task_complete", False) else "generate_command",
        {"report_result": "report_result", "generate_command": "generate_command"},
    )
    graph.add_edge("report_result", END)

    return graph.compile()
