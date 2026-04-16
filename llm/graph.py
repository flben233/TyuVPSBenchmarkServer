import json
import threading
import re
from typing import Any

from langchain_core.messages import (
    AIMessage,
    BaseMessage,
    HumanMessage,
    SystemMessage,
    ToolMessage,
)
from langchain_core.tools import BaseTool
from langgraph.graph import END, START, StateGraph
from openai import OpenAI

from state import AgentState, ensure_chunk_queue
from tool import build_openai_tools, inject_session_id, invoke_tool

DANGEROUS_PATTERNS = [
    re.compile(r"\brm\s+.*-.*f", re.IGNORECASE),
    re.compile(r"\bshutdown\b", re.IGNORECASE),
    re.compile(r"\breboot\b", re.IGNORECASE),
    re.compile(r"\bmkfs\b", re.IGNORECASE),
    re.compile(r"\bdd\s+if=", re.IGNORECASE),
    re.compile(r">\s*/dev/sd", re.IGNORECASE),
]

SYSTEM_PROMPT = (
    "You are a server operations assistant connected to a remote server via SSH. "
    "You help users manage their servers by executing commands when needed. "
    "Explain what you are going to do before running commands. "
    "Summarize the results after execution. "
    "Be cautious with destructive operations."
)


def _to_openai_messages(messages: list[BaseMessage]) -> list[dict[str, Any]]:
    result: list[dict[str, Any]] = []
    for message in messages:
        if isinstance(message, SystemMessage):
            result.append({"role": "system", "content": str(message.content or "")})
            continue

        if isinstance(message, HumanMessage):
            result.append({"role": "user", "content": str(message.content or "")})
            continue

        if isinstance(message, ToolMessage):
            result.append(
                {
                    "role": "tool",
                    "tool_call_id": message.tool_call_id,
                    "content": str(message.content or ""),
                }
            )
            continue

        if isinstance(message, AIMessage):
            tool_calls = []
            for tc in message.tool_calls:
                args = tc.get("args", {})
                tool_calls.append(
                    {
                        "id": tc.get("id") or "",
                        "type": "function",
                        "function": {
                            "name": tc.get("name") or "",
                            "arguments": json.dumps(args, ensure_ascii=False),
                        },
                    }
                )

            assistant_message: dict[str, Any] = {
                "role": "assistant",
                "content": str(message.content or ""),
            }
            if tool_calls:
                assistant_message["tool_calls"] = tool_calls
            result.append(assistant_message)
            continue

        result.append({"role": "user", "content": str(message.content or "")})

    return result


def _extract_delta_texts(delta: Any) -> tuple[str, str]:
    content = getattr(delta, "content", None) or ""
    reasoning = getattr(delta, "reasoning_content", None) or ""
    return str(content), str(reasoning)


def is_dangerous_command(command: str) -> bool:
    return any(p.search(command) for p in DANGEROUS_PATTERNS)


def build_graph(
    *,
    openai_client: OpenAI,
    model: str,
    tools: list[BaseTool],
    stop_event: threading.Event | None = None,
    checkpointer=None,
):
    tool_map = {t.name: t for t in tools}
    openai_tools = build_openai_tools(tools)

    def route_start(state: AgentState) -> str:
        if state.get("awaiting_approval") and state.get("approval_granted") is not None:
            return "execute_approved" if state["approval_granted"] else "reject_tool"
        return "assistant"

    def assistant(state: AgentState) -> dict:
        msgs = [SystemMessage(content=SYSTEM_PROMPT)] + state.get("messages", [])
        openai_messages = _to_openai_messages(msgs)
        chunk_queue = ensure_chunk_queue(state)
        final_text_parts: list[str] = []
        final_tool_calls: list[dict[str, Any]] = []

        tool_call_builders: dict[int, dict[str, str]] = {}
        request_kwargs: dict[str, Any] = {
            "model": model,
            "messages": openai_messages,
            "stream": True,
        }
        if openai_tools:
            request_kwargs["tools"] = openai_tools
            request_kwargs["tool_choice"] = "auto"

        stream = openai_client.chat.completions.create(**request_kwargs)

        stopped = False
        for chunk in stream:
            if stop_event and stop_event.is_set():
                stopped = True
                break
            if not chunk.choices:
                continue
            delta = chunk.choices[0].delta
            token_text, thinking_text = _extract_delta_texts(delta)

            if thinking_text:
                chunk_queue.put({"kind": "thinking", "text": thinking_text})

            if token_text:
                final_text_parts.append(token_text)
                chunk_queue.put({"kind": "token", "text": token_text})

            delta_tool_calls = getattr(delta, "tool_calls", None) or []
            for delta_tool_call in delta_tool_calls:
                index = int(getattr(delta_tool_call, "index", 0) or 0)
                tool_builder = tool_call_builders.setdefault(
                    index,
                    {
                        "id": "",
                        "name": "",
                        "arguments": "",
                    },
                )

                tc_id = getattr(delta_tool_call, "id", None)
                if tc_id:
                    tool_builder["id"] = tc_id

                function = getattr(delta_tool_call, "function", None)
                if function is not None:
                    name_part = getattr(function, "name", None)
                    if name_part:
                        tool_builder["name"] += name_part
                    args_part = getattr(function, "arguments", None)
                    if args_part:
                        tool_builder["arguments"] += args_part

        for index in sorted(tool_call_builders.keys()):
            tool_builder = tool_call_builders[index]
            args: dict[str, Any]
            try:
                args = json.loads(tool_builder["arguments"] or "{}")
            except json.JSONDecodeError:
                args = {}
            tool_name = tool_builder["name"]
            if not tool_name:
                continue
            final_tool_calls.append(
                {
                    "id": tool_builder["id"] or f"tool_{index}",
                    "name": tool_name,
                    "args": args,
                    "type": "tool_call",
                }
            )

        response = AIMessage(content="".join(final_text_parts), tool_calls=final_tool_calls)

        updates: dict[str, Any] = {"messages": [response], "stopped": stopped}

        if stopped:
            return updates

        if response.tool_calls:
            command = response.tool_calls[0].get("args", {}).get("command", "")
            if is_dangerous_command(command):
                updates["pending_tool_call"] = response.tool_calls[0]
                updates["awaiting_approval"] = True
            else:
                updates["pending_tool_call"] = None
                updates["awaiting_approval"] = False
        else:
            updates["pending_tool_call"] = None
            updates["awaiting_approval"] = False

        return updates

    def route_after_assistant(state: AgentState) -> str:
        if state.get("stopped"):
            return "stopped"
        if state.get("awaiting_approval"):
            return "awaiting_approval"
        last = state["messages"][-1] if state.get("messages") else None
        if isinstance(last, AIMessage) and getattr(last, "tool_calls", None):
            return "execute_tools"
        state["messages"] = []
        return END

    def execute_tools(state: AgentState) -> dict:
        last = state["messages"][-1]
        if not isinstance(last, AIMessage) or not last.tool_calls:
            return {}
        results = []
        for tc in last.tool_calls:
            tc_with_session = inject_session_id(tc, state.get("ssh_session_id", ""))
            results.append(invoke_tool(tc_with_session, tool_map))
        return {
            "messages": results,
            "pending_tool_call": None,
            "awaiting_approval": False,
        }

    def execute_approved(state: AgentState) -> dict:
        pending = state.get("pending_tool_call")
        if not pending:
            return {}
        tc = inject_session_id(pending, state.get("ssh_session_id", ""))
        ai_msg = AIMessage(content="", tool_calls=[pending])
        tool_msg = invoke_tool(tc, tool_map)
        return {
            "messages": [ai_msg, tool_msg],
            "pending_tool_call": None,
            "awaiting_approval": False,
            "approval_granted": None,
        }

    def reject_tool(state: AgentState) -> dict:
        pending = state.get("pending_tool_call")
        if not pending:
            return {}
        ai_msg = AIMessage(content="", tool_calls=[pending])
        tool_msg = ToolMessage(
            content="User rejected this command. Acknowledge and suggest alternatives.",
            tool_call_id=pending["id"],
            name=pending["name"],
        )
        return {
            "messages": [ai_msg, tool_msg],
            "pending_tool_call": None,
            "awaiting_approval": False,
            "approval_granted": None,
        }

    builder = StateGraph(AgentState)

    builder.add_node("assistant", assistant)
    builder.add_node("execute_tools", execute_tools)
    builder.add_node("execute_approved", execute_approved)
    builder.add_node("reject_tool", reject_tool)

    builder.add_conditional_edges(
        START,
        route_start,
        {
            "assistant": "assistant",
            "execute_approved": "execute_approved",
            "reject_tool": "reject_tool",
        },
    )
    builder.add_conditional_edges(
        "assistant",
        route_after_assistant,
        {
            "execute_tools": "execute_tools",
            "awaiting_approval": END,
            "stopped": END,
            END: END,
        },
    )
    builder.add_edge("execute_tools", "assistant")
    builder.add_edge("execute_approved", "assistant")
    builder.add_edge("reject_tool", "assistant")

    return builder.compile(checkpointer=checkpointer)
