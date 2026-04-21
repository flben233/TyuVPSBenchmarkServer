import asyncio
import json
import threading
from contextlib import asynccontextmanager
from datetime import datetime, timezone
from queue import Empty
from typing import Any
from uuid import uuid4

from fastapi import FastAPI, HTTPException
from fastapi.responses import StreamingResponse
from openai import OpenAI

from config import Settings
from context_compression import (
    build_candidate_context_messages,
    compress_context_messages,
    to_message_dicts,
)
from free_api_pool import FreeAPIPool, FailoverClient, FreeAPIExhaustedError
from graph import (
    build_graph,
    ensure_chunk_queue,
)
from state import default_agent_state, append_user_message, clear_chunk_queue, set_messages_from_dicts
from tool import load_mcp_tools
from model import ConversationRuntime, NewConversationResponse, NewConversationRequest, CloseResponse, CloseRequest, \
    ChatRequest, StopRequest, StopResponse


def _validate_chat_message_lengths(request: ChatRequest, max_chars: int) -> None:
    if request.message and len(request.message) > max_chars:
        raise HTTPException(
            status_code=400,
            detail=f"message too long: max {max_chars} characters",
        )
    for message in request.messages or []:
        if len(message.content or "") > max_chars:
            raise HTTPException(
                status_code=400,
                detail=f"message too long: max {max_chars} characters",
            )


@asynccontextmanager
async def lifespan(_app: FastAPI):
    try:
        yield
    finally:
        conversation_runtimes.clear()
        ssh_session_to_conversation_ids.clear()


app = FastAPI(title="WebSSH LLM Agent", lifespan=lifespan)
settings: Settings | None = None
pool: FreeAPIPool | None = None
conversation_runtimes: dict[str, ConversationRuntime] = {}
ssh_session_to_conversation_ids: dict[str, set[str]] = {}


def _sse(event: str, payload: dict[str, Any]) -> str:
    return f"event: {event}\ndata: {json.dumps(payload)}\n\n"


def _now_iso() -> str:
    return datetime.now(timezone.utc).isoformat()


def _ensure_settings():
    global settings, pool
    if settings is None:
        settings = Settings.from_env()
    if pool is None:
        pool = FreeAPIPool()


@app.post("/new", response_model=NewConversationResponse, tags=["LLM Agent"], summary="Create a new conversation")
async def new_conversation(request: NewConversationRequest) -> NewConversationResponse:
    _ensure_settings()

    conversation_id = str(uuid4())
    selected_client: Any
    selected_model: str
    user_id = request.user_id or "anonymous"

    if request.llm_api is not None:
        api_base = (request.llm_api.api_base or "").strip()
        api_key = (request.llm_api.api_key or "").strip()
        model = (request.llm_api.model or "").strip()
        if api_base or api_key or model:
            if not api_base or not api_key or not model:
                raise HTTPException(
                    status_code=400,
                    detail="custom llm api requires apiBase, apiKey and model",
                )
            selected_client = OpenAI(base_url=api_base, api_key=api_key)
            selected_model = model
        else:
            raise HTTPException(
                status_code=400,
                detail="llmApi provided but all fields are empty",
            )
    else:
        try:
            failover = FailoverClient(pool, user_id)
            failover.acquire()
        except FreeAPIExhaustedError:
            if pool.has_endpoints:
                raise HTTPException(
                    status_code=429,
                    detail="Free API rate limit exceeded or all endpoints are in cooldown. "
                           "Please provide your own API key or try again later.",
                )
            raise HTTPException(
                status_code=503,
                detail="No free API endpoints configured. Please provide your own API key.",
            )
        selected_client = failover
        selected_model = ""

    tools = await load_mcp_tools(settings.mcp_server_url)
    state = default_agent_state(conversation_id, request.ssh_session_id)
    stop_event = threading.Event()
    graph = build_graph(
        openai_client=selected_client,
        model=selected_model,
        tools=tools,
        stop_event=stop_event,
    )
    runtime = ConversationRuntime(
        graph=graph,
        state=state,
        lock=asyncio.Lock(),
        stop_event=stop_event,
        user_id=user_id,
        openai_client=selected_client,
        model=selected_model,
        context_tail_keep=settings.context_tail_keep,
        compress_threshold_tokens=settings.compress_threshold_tokens,
        max_chat_message_chars=settings.max_chat_message_chars,
    )

    conversation_runtimes[conversation_id] = runtime
    ssh_session_to_conversation_ids.setdefault(request.ssh_session_id, set()).add(
        conversation_id
    )

    return NewConversationResponse(
        conversationId=conversation_id,
        maxChatMessageChars=settings.max_chat_message_chars,
    )


@app.post("/close", response_model=CloseResponse, tags=["LLM Agent"], summary="Close all conversations for an SSH session")
async def close_conversation_state(request: CloseRequest) -> CloseResponse:
    conversation_ids = ssh_session_to_conversation_ids.pop(request.ssh_session_id, set())

    closed_count = 0
    for conversation_id in conversation_ids:
        runtime = conversation_runtimes.pop(conversation_id, None)
        if runtime is None:
            continue
        closed_count += 1

    return CloseResponse(closedConversations=closed_count)


@app.post("/stop", response_model=StopResponse, tags=["LLM Agent"], summary="Stop an in-progress LLM response")
async def stop_chat(request: StopRequest) -> StopResponse:
    """
    Signal the LLM agent to stop generating the current response.
    The graph will break out of the streaming loop and route directly to END.
    """
    runtime = conversation_runtimes.get(request.conversation_id)
    if not runtime:
        raise HTTPException(status_code=404, detail="conversation not found")
    _validate_chat_message_lengths(request, runtime.max_chat_message_chars)
    runtime.stop_event.set()
    return StopResponse(stopped=True)


@app.post("/chat", tags=["LLM Agent"], summary="Send a message and stream the LLM response")
async def chat(request: ChatRequest) -> StreamingResponse:
    """
    Send a message to the LLM agent and receive a streaming SSE response.

    The frontend manages the full conversation history and passes it via the
    ``messages`` field on every request.  The backend sets these messages into
    the session state before invoking the graph, and clears them after the
    response completes so that no history is retained server-side.

    **SSE event types:**

    | Event | Description |
    |---|---|
    | `message_start` | Response stream begins |
    | `thinking` | AI reasoning token |
    | `token` | AI response token |
    | `message_end` | Response stream ends |
    | `stopped` | Response was stopped by user via /stop |
    | `awaiting_approval` | AI wants to run a dangerous command, waiting for user approval |
    | `error` | An error occurred |
    | `done` | Stream finished |
    """
    runtime = conversation_runtimes.get(request.conversation_id)
    if not runtime:
        raise HTTPException(status_code=404, detail="conversation not found")

    async def event_stream():
        message_id = str(uuid4())
        streamed_response_text = ""
        async with runtime.lock:
            runtime.stop_event.clear()
            if request.messages is not None:
                # 这里设置上，graph在处理完后会清除，保证每次请求都是全量的消息历史
                runtime.state = set_messages_from_dicts(
                    runtime.state,
                    [m.model_dump() for m in request.messages],
                )
            if request.message:
                runtime.state = append_user_message(runtime.state, request.message)
            if request.approval_granted is not None:
                runtime.state["approval_granted"] = request.approval_granted
            if request.allowed_commands is not None:
                runtime.state["allowed_commands"] = request.allowed_commands
            if request.session_allowed_commands is not None:
                runtime.state["session_allowed_commands"] = request.session_allowed_commands

            chunk_queue = ensure_chunk_queue(runtime.state)
            clear_chunk_queue(runtime.state)
            runtime.state["stopped"] = False
            graph_task = asyncio.create_task(asyncio.to_thread(runtime.graph.invoke, runtime.state))

            yield _sse(
                "message_start",
                {
                    "conversationId": request.conversation_id,
                    "messageId": message_id,
                    "timestamp": _now_iso(),
                    "payload": {},
                },
            )

            while True:
                drained = False
                while True:
                    try:
                        token_item = chunk_queue.get_nowait()
                        drained = True
                    except Empty:
                        break

                    if isinstance(token_item, dict):
                        token_kind = str(token_item.get("kind", "token"))
                        token_text = str(token_item.get("text", ""))
                    else:
                        token_kind = "token"
                        token_text = str(token_item)

                    event_name = "thinking" if token_kind == "thinking" else "token"
                    if event_name == "token" and token_text:
                        streamed_response_text += token_text

                    yield _sse(
                        event_name,
                        {
                            "conversationId": request.conversation_id,
                            "messageId": message_id,
                            "timestamp": _now_iso(),
                            "payload": {"text": token_text},
                        },
                    )

                if graph_task.done():
                    break

                if not drained:
                    await asyncio.sleep(0.02)

            try:
                runtime.state = graph_task.result()
            except Exception as exc:
                yield _sse(
                    "error",
                    {
                        "conversationId": request.conversation_id,
                        "messageId": message_id,
                        "timestamp": _now_iso(),
                        "payload": {"message": str(exc)},
                    },
                )
                yield _sse(
                    "done",
                    {
                        "conversationId": request.conversation_id,
                        "messageId": message_id,
                        "timestamp": _now_iso(),
                        "payload": {},
                    },
                )
                return

            assistant_response = streamed_response_text

            done_payload: dict[str, Any] = {}
            candidate_context_messages = build_candidate_context_messages(
                request_messages=to_message_dicts(request.messages),
                request_message=request.message,
                assistant_response=assistant_response,
            )

            compressed_context_messages = None
            if candidate_context_messages:
                try:
                    compression_model = runtime.model or getattr(runtime.openai_client, "_model", "")
                    usage_total_tokens = None
                    last_usage = runtime.state.get("last_usage") or {}
                    if isinstance(last_usage, dict):
                        raw_total = last_usage.get("total_tokens")
                        if raw_total is not None:
                            usage_total_tokens = int(raw_total)
                    if usage_total_tokens is not None and usage_total_tokens >= runtime.compress_threshold_tokens:
                        yield _sse(
                            "compressing",
                            {
                                "conversationId": request.conversation_id,
                                "messageId": message_id,
                                "timestamp": _now_iso(),
                                "payload": {},
                            },
                        )
                    compressed_context_messages = compress_context_messages(
                        openai_client=runtime.openai_client,
                        model=compression_model,
                        context_messages=candidate_context_messages,
                        tail_keep=runtime.context_tail_keep,
                        compress_threshold_tokens=runtime.compress_threshold_tokens,
                        usage_total_tokens=usage_total_tokens,
                    )
                except Exception:
                    compressed_context_messages = None

            if compressed_context_messages:
                done_payload = {
                    "contextCompressed": True,
                    "contextMessages": compressed_context_messages,
                }

            yield _sse(
                "message_end",
                {
                    "conversationId": request.conversation_id,
                    "messageId": message_id,
                    "timestamp": _now_iso(),
                    "payload": {},
                },
            )

            if runtime.state.get("stopped"):
                yield _sse(
                    "stopped",
                    {
                        "conversationId": request.conversation_id,
                        "messageId": message_id,
                        "timestamp": _now_iso(),
                        "payload": {},
                    },
                )

            if runtime.state.get("awaiting_approval"):
                pending = runtime.state.get("pending_tool_call") or {}
                disallowed = runtime.state.get("disallowed_commands") or []
                yield _sse(
                    "awaiting_approval",
                    {
                        "conversationId": request.conversation_id,
                        "messageId": message_id,
                        "timestamp": _now_iso(),
                        "payload": {
                            **pending,
                            "disallowed_commands": disallowed,
                        },
                    },
                )

            yield _sse(
                "done",
                {
                    "conversationId": request.conversation_id,
                    "messageId": message_id,
                    "timestamp": _now_iso(),
                    "payload": done_payload,
                },
            )

    return StreamingResponse(
        event_stream(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
        },
    )


if __name__ == '__main__':
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=18000)
