import json
import threading
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
from command_checker import CommandChecker

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


class _ThoughtParser:
    """Stateful streaming parser that splits <thought>...</thought> tags in content.

    Some OpenAI-compatible APIs (e.g. certain GLM / Qwen deployments) embed the
    model's reasoning inside ``<thought>...</thought>`` within the regular
    ``content`` field instead of using a dedicated ``reasoning_content`` field.
    Because streaming deltas can split a tag across chunks (e.g. ``"<thou"`` in
    one delta and ``"ght>"`` in the next), a naive ``str.replace`` approach is
    insufficient.

    This parser maintains two pieces of state:
      * ``_in_thought`` – whether we are currently inside a ``<thought>`` block.
      * ``_buf`` – a small buffer for the tail of the current chunk that *might*
        be the beginning of an opening or closing tag.  On the next ``feed()``
        call the buffer is prepended to the new text so the tag can be matched
        in full.

    Returns ``(content_text, thinking_text)`` from both ``feed()`` and
    ``flush()``, where *content_text* is everything outside the tags and
    *thinking_text* is everything inside.
    """

    _OPEN = "<thought>"
    _CLOSE = "</thought>"
    _MAX_TAG = max(len(_OPEN), len(_CLOSE))

    def __init__(self) -> None:
        self._in_thought = False
        self._buf = ""

    def feed(self, text: str) -> tuple[str, str]:
        """Process one streaming delta and return ``(content, thinking)``.

        Prepend the leftover buffer from the previous call, then scan for tag
        boundaries.  Any text that cannot yet be classified (because it might be
        the start of a tag) is stored back into ``_buf`` for the next call.
        """
        if not text:
            return "", ""
        # Prepend leftover from the previous delta that may be a partial tag
        text = self._buf + text
        self._buf = ""
        content_parts: list[str] = []
        thinking_parts: list[str] = []
        i = 0
        while i < len(text):
            if self._in_thought:
                # Look for the closing tag
                idx = text.find(self._CLOSE, i)
                if idx == -1:
                    thinking_parts.append(text[i:])
                    break
                # Found closing tag – emit the thinking text up to it
                thinking_parts.append(text[i:idx])
                self._in_thought = False
                i = idx + len(self._CLOSE)
            else:
                # Look for the opening tag
                idx = text.find(self._OPEN, i)
                if idx == -1:
                    content_parts.append(text[i:])
                    break
                # Found opening tag – emit the content text up to it
                content_parts.append(text[i:idx])
                self._in_thought = True
                i = idx + len(self._OPEN)
        return "".join(content_parts), "".join(thinking_parts)

    def flush(self) -> tuple[str, str]:
        """Drain the buffer after the stream ends.

        Any remaining buffered text is classified according to the current
        state: still inside ``<thought>`` → thinking; otherwise → content.
        """
        remaining = self._buf
        self._buf = ""
        if not remaining:
            return "", ""
        if self._in_thought:
            return "", remaining
        return remaining, ""


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

        # Stateful parser for <thought>...</thought> tags in content deltas.
        # Only used when the API does NOT provide a dedicated reasoning_content field.
        thought_parser = _ThoughtParser()
        stopped = False
        for chunk in stream:
            if stop_event and stop_event.is_set():
                stopped = True
                break
            if not chunk.choices:
                continue
            delta = chunk.choices[0].delta

            raw_reasoning = getattr(delta, "reasoning_content", None) or ""
            raw_content = getattr(delta, "content", None) or ""
            raw_reasoning = str(raw_reasoning)
            raw_content = str(raw_content)

            # Branch 1: API provides a dedicated reasoning_content field
            # (e.g. DeepSeek).  Use it directly as thinking; content is content.
            if raw_reasoning:
                chunk_queue.put({"kind": "thinking", "text": raw_reasoning})
                if raw_content:
                    final_text_parts.append(raw_content)
                    chunk_queue.put({"kind": "token", "text": raw_content})
            # Branch 2: No reasoning_content – parse <thought> tags from content.
            # Some APIs (e.g. certain GLM/Qwen) embed reasoning inside these tags.
            elif raw_content:
                content_text, thinking_text = thought_parser.feed(raw_content)
                if thinking_text:
                    chunk_queue.put({"kind": "thinking", "text": thinking_text})
                if content_text:
                    final_text_parts.append(content_text)
                    chunk_queue.put({"kind": "token", "text": content_text})

            # Accumulate tool call deltas (id / name / arguments arrive piecemeal)
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

        # Flush any remaining buffered text from the thought parser
        content_text, thinking_text = thought_parser.flush()
        if thinking_text:
            chunk_queue.put({"kind": "thinking", "text": thinking_text})
        if content_text:
            final_text_parts.append(content_text)
            chunk_queue.put({"kind": "token", "text": content_text})

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
            checker = CommandChecker(
                user_allowed=state.get("allowed_commands"),
                session_allowed=state.get("session_allowed_commands"),
            )
            check_result = checker.check(command)
            if not check_result.all_allowed:
                updates["pending_tool_call"] = response.tool_calls[0]
                updates["disallowed_commands"] = check_result.disallowed
                updates["awaiting_approval"] = True
            else:
                updates["pending_tool_call"] = None
                updates["disallowed_commands"] = []
                updates["awaiting_approval"] = False
        else:
            updates["pending_tool_call"] = None
            updates["disallowed_commands"] = []
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
