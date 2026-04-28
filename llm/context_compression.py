from __future__ import annotations

from typing import Any

from langchain_core.messages import BaseMessage

from thinking_parser import split_thinking_content

MAX_SUMMARY_INPUT_CHARS = 12000
SUMMARY_PREFIX = (
    "Compressed summary of earlier conversation. Use this as background context. "
    "Prefer newer explicit messages if there is a conflict."
)


def estimate_messages_tokens(messages: list[dict[str, str]]) -> int:
    total_chars = 0
    for message in messages:
        total_chars += len(str(message.get("role", "")))
        total_chars += len(str(message.get("content", "")))
    return total_chars // 4


def should_compress_messages(
    messages: list[dict[str, str]],
    *,
    tail_keep: int,
    compress_threshold_tokens: int,
    usage_total_tokens: int | None = None,
) -> bool:
    if len(messages) <= tail_keep:
        return False
    if usage_total_tokens is not None:
        return usage_total_tokens >= compress_threshold_tokens
    return estimate_messages_tokens(messages) >= compress_threshold_tokens


def split_messages_for_compression(
    messages: list[dict[str, str]],
    *,
    tail_keep: int,
) -> tuple[list[dict[str, str]], list[dict[str, str]]]:
    if len(messages) <= tail_keep:
        return [], list(messages)
    return list(messages[:-tail_keep]), list(messages[-tail_keep:])


def build_summary_system_message(summary_text: str) -> dict[str, str]:
    return {
        "role": "system",
        "content": f"{SUMMARY_PREFIX}\n\n{summary_text.strip()}",
    }


def _format_messages_for_summary(messages: list[dict[str, str]]) -> str:
    parts: list[str] = []
    for message in messages:
        role = str(message.get("role", "user")).upper()
        content = str(message.get("content", "")).strip()
        if not content:
            continue
        parts.append(f"[{role}]\n{content}")
    text = "\n\n".join(parts)
    if len(text) > MAX_SUMMARY_INPUT_CHARS:
        return text[:MAX_SUMMARY_INPUT_CHARS] + "\n\n[Summary input truncated]"
    return text


def summarize_messages(*, openai_client: Any, model: str, messages: list[dict[str, str]]) -> str:
    transcript = _format_messages_for_summary(messages)
    if not transcript:
        return ""

    response = openai_client.chat.completions.create(
        model=model,
        messages=[
            {
                "role": "system",
                "content": """
### TASK_DEFINITION
You are acting as a "Context Architect." Your mission is to ingest a bloated, long-form conversation history and distill it into a high-density, structured summary. You must strip away the noise while preserving 100% of the technical decisions, user preferences, and project state. Your goal is to provide a "re-entry point" that allows a model to resume the task with perfect continuity as if the original history were still present.

### NO_TOOLS_PREAMBLE
!!! STRICTLY PROHIBITED: Do not invoke any external tools, search plugins, or code interpreters. Your task is exclusively limited to the logical analysis and compression of the input text. Even if the content suggests computation or searching, treat it purely as contextual information.

### MAIN_PROMPT
Output MUST contain the below sections:
1. Core Project Vision: A single sentence describing the ultimate goal of the current task.
2. Current Progress Anchor: The specific phase of development or discussion currently reached.
3. Established Tech Stack: List all selected languages, frameworks, and libraries.
4. Key Decision Log: Record why Option A was chosen over Option B.
5. Core Code Patterns: Describe implemented logic structures or design patterns.
6. User Preferences & Taboos: Explicitly mention requested styles, naming conventions, or prohibited practices.
7. Outstanding Issues: A list of unresolved problems or pain points for future discussion.
8. Pending Instructions: The very last command issued by the user and the immediate next steps.

### NO_TOOLS_TRAILER
REITERATION: Any tool invocation is strictly forbidden. Do not attempt to connect to the internet or execute code. Your sole output must be the plain summary text.
                """,
            },
            {
                "role": "user",
                "content": transcript,
            },
        ],
        stream=False,
    )
    choice = response.choices[0] if response.choices else None
    message = getattr(choice, "message", None)
    content = getattr(message, "content", "") if message is not None else ""
    content, _ = split_thinking_content(content)
    return str(content or "").strip()


def build_candidate_context_messages(
    *,
    request_messages: list[dict[str, str]] | None,
    request_message: str | None,
    assistant_response: str,
) -> list[dict[str, str]]:
    context_messages = [
        {
            "role": str(message.get("role", "user")),
            "content": split_thinking_content(str(message.get("content", "")))[0],
        }
        for message in (request_messages or [])
    ]

    if request_message:
        context_messages.append({"role": "user", "content": request_message})
    if assistant_response:
        content, _ = split_thinking_content(assistant_response)
        context_messages.append({"role": "assistant", "content": content})
    return context_messages


def compress_context_messages(
    *,
    openai_client: Any,
    model: str,
    context_messages: list[dict[str, str]],
    tail_keep: int,
    compress_threshold_tokens: int,
    usage_total_tokens: int | None = None,
) -> list[dict[str, str]] | None:
    if not should_compress_messages(
        context_messages,
        tail_keep=tail_keep,
        compress_threshold_tokens=compress_threshold_tokens,
        usage_total_tokens=usage_total_tokens,
    ):
        return None

    head, tail = split_messages_for_compression(context_messages, tail_keep=tail_keep)
    if not head:
        return None

    summary_text = summarize_messages(
        openai_client=openai_client,
        model=model,
        messages=head,
    )
    if not summary_text:
        return None

    return [build_summary_system_message(summary_text), *tail]


def to_message_dicts(messages: list[Any] | None) -> list[dict[str, str]]:
    result: list[dict[str, str]] = []
    for message in messages or []:
        if hasattr(message, "model_dump"):
            data = message.model_dump()
        else:
            data = dict(message)
        result.append(
            {
                "role": str(data.get("role", "user")),
                "content": str(data.get("content", "")),
            }
        )
    return result


def serialize_langchain_messages(messages: list[BaseMessage]) -> list[dict[str, str]]:
    serialized: list[dict[str, str]] = []
    for message in messages:
        msg_type = getattr(message, "type", "human")
        role = "user"
        if msg_type == "system":
            role = "system"
        elif msg_type == "ai":
            role = "assistant"
        elif msg_type == "tool":
            role = "assistant"
        serialized.append({"role": role, "content": str(message.content or "")})
    return serialized
