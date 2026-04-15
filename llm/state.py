from queue import Empty, Queue
from queue import Empty, Queue
from typing import Annotated, Any

from langchain_core.messages import (
    BaseMessage,
    HumanMessage,
)
from langgraph.graph.message import add_messages
from typing_extensions import TypedDict


class AgentState(TypedDict, total=False):
    conversation_id: str
    ssh_session_id: str
    messages: Annotated[list[BaseMessage], add_messages]
    chunk_queue: Queue[dict[str, str] | str]
    pending_tool_call: dict[str, Any] | None
    awaiting_approval: bool
    approval_granted: bool | None


def default_agent_state(conversation_id: str, ssh_session_id: str) -> AgentState:
    return {
        "conversation_id": conversation_id,
        "ssh_session_id": ssh_session_id,
        "messages": [],
        "chunk_queue": Queue(),
        "pending_tool_call": None,
        "awaiting_approval": False,
        "approval_granted": None,
    }


def append_user_message(state: AgentState, message: str) -> AgentState:
    state["messages"] = [*state.get("messages", []), HumanMessage(content=message)]
    return state


def ensure_chunk_queue(state: AgentState) -> Queue[dict[str, str] | str]:
    queue = state.get("chunk_queue")
    if queue is None:
        queue = Queue()
        state["chunk_queue"] = queue
    return queue


def clear_chunk_queue(state: AgentState) -> None:
    queue = ensure_chunk_queue(state)
    while True:
        try:
            queue.get_nowait()
        except Empty:
            break