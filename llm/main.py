import asyncio
import json
from contextlib import asynccontextmanager
from datetime import datetime, timezone
from queue import Empty
from typing import Any
from uuid import uuid4

from fastapi import FastAPI, HTTPException
from fastapi.responses import StreamingResponse
from openai import OpenAI

from config import Settings
from graph import (
    build_graph,
    ensure_chunk_queue,
)
from state import default_agent_state, append_user_message, clear_chunk_queue
from tool import load_mcp_tools
from model import ConversationRuntime, NewConversationResponse, NewConversationRequest, CloseResponse, CloseRequest, \
    ChatRequest


@asynccontextmanager
async def lifespan(_app: FastAPI):
    try:
        yield
    finally:
        conversation_runtimes.clear()
        ssh_session_to_conversation_ids.clear()


app = FastAPI(title="WebSSH LLM Agent", lifespan=lifespan)
settings: Settings | None = None
openai_client: OpenAI | None = None
conversation_runtimes: dict[str, ConversationRuntime] = {}
ssh_session_to_conversation_ids: dict[str, set[str]] = {}


def _sse(event: str, payload: dict[str, Any]) -> str:
    return f"event: {event}\ndata: {json.dumps(payload)}\n\n"


def _now_iso() -> str:
    return datetime.now(timezone.utc).isoformat()


@app.post("/new", response_model=NewConversationResponse)
async def new_conversation(request: NewConversationRequest) -> NewConversationResponse:
    global settings, openai_client
    if settings is None:
        settings = Settings.from_env()
    if openai_client is None:
        openai_client = OpenAI(
            base_url=settings.openai_api_base,
            api_key=settings.openai_api_key,
        )

    conversation_id = str(uuid4())

    tools = await load_mcp_tools(settings.mcp_server_url)
    graph = build_graph(openai_client=openai_client, model=settings.openai_model, tools=tools)

    conversation_runtimes[conversation_id] = ConversationRuntime(
        graph=graph,
        state=default_agent_state(conversation_id, request.ssh_session_id),
        lock=asyncio.Lock(),
    )
    ssh_session_to_conversation_ids.setdefault(request.ssh_session_id, set()).add(
        conversation_id
    )

    return NewConversationResponse(conversationId=conversation_id)


@app.post("/close", response_model=CloseResponse)
async def close_conversation_state(request: CloseRequest) -> CloseResponse:
    conversation_ids = ssh_session_to_conversation_ids.pop(request.ssh_session_id, set())

    closed_count = 0
    for conversation_id in conversation_ids:
        runtime = conversation_runtimes.pop(conversation_id, None)
        if runtime is None:
            continue
        closed_count += 1

    return CloseResponse(closedConversations=closed_count)


@app.post("/chat")
async def chat(request: ChatRequest) -> StreamingResponse:
    runtime = conversation_runtimes.get(request.conversation_id)
    if not runtime:
        raise HTTPException(status_code=404, detail="conversation not found")

    async def event_stream():
        message_id = str(uuid4())
        async with runtime.lock:
            if request.message:
                runtime.state = append_user_message(runtime.state, request.message)
            if request.approval_granted is not None:
                runtime.state["approval_granted"] = request.approval_granted

            chunk_queue = ensure_chunk_queue(runtime.state)
            clear_chunk_queue(runtime.state)
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

            yield _sse(
                "message_end",
                {
                    "conversationId": request.conversation_id,
                    "messageId": message_id,
                    "timestamp": _now_iso(),
                    "payload": {},
                },
            )

            if runtime.state.get("awaiting_approval"):
                pending = runtime.state.get("pending_tool_call")
                yield _sse(
                    "awaiting_approval",
                    {
                        "conversationId": request.conversation_id,
                        "messageId": message_id,
                        "timestamp": _now_iso(),
                        "payload": pending or {},
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
