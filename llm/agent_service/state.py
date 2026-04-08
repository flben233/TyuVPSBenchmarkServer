from __future__ import annotations

from typing import TypedDict


class AgentState(TypedDict, total=False):
    task_id: str
    user_id: str
    messages: list[str]
    current_command: str
    command_result: str
    safety_status: str
    awaiting_approval: bool
    steps: list[str]
    task_complete: bool
    final_response: str
    approval_granted: bool
    approved_command: str
    retry_count: int


def default_agent_state(task_id: str, user_id: str = "anonymous") -> AgentState:
    return AgentState(
        task_id=task_id,
        user_id=user_id,
        messages=[],
        current_command="",
        command_result="",
        safety_status="unknown",
        awaiting_approval=False,
        steps=[],
        task_complete=False,
        final_response="",
        approval_granted=False,
        approved_command="",
        retry_count=0,
    )


def merge_with_defaults(state: AgentState) -> AgentState:
    merged = default_agent_state(
        task_id=state.get("task_id", ""),
        user_id=state.get("user_id", "anonymous"),
    )
    merged.update(state)
    return merged


def append_step(state: AgentState, step: str) -> AgentState:
    steps = list(state.get("steps", []))
    steps.append(step)
    state["steps"] = steps
    return state
