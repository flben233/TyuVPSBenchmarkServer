from __future__ import annotations

import asyncio
import logging
from contextlib import asynccontextmanager
from uuid import uuid4
from typing import Any, AsyncIterator

from fastapi import FastAPI
from fastapi import HTTPException
from redis import Redis

from .config import Settings
from .graph import InMemoryCheckpointer, RedisCheckpointer, build_agent_graph, checkpoint_key
from .go_client import GoToolClient
from .state import default_agent_state
from .schemas import (
    AgentResponse,
    ApproveTaskRequest,
    CreateTaskRequest,
    CreateTaskResponse,
    TaskMessageRequest,
)


logger = logging.getLogger(__name__)


class GoBridgeStreamEmitter:
    def __init__(self, go_client: GoToolClient, loop: asyncio.AbstractEventLoop) -> None:
        self._go_client = go_client
        self._loop = loop

    def on_message_start(self, task_id: str, message_id: str) -> None:
        self._schedule(task_id, {"type": "agent_message_start", "task_id": task_id, "message_id": message_id})

    def on_token(self, task_id: str, message_id: str, delta: str) -> None:
        self._schedule(
            task_id,
            {
                "type": "agent_token",
                "task_id": task_id,
                "message_id": message_id,
                "delta": delta,
            },
        )

    def on_message_end(self, task_id: str, message_id: str, finish_reason: str) -> None:
        self._schedule(
            task_id,
            {
                "type": "agent_message_end",
                "task_id": task_id,
                "message_id": message_id,
                "finish_reason": finish_reason,
            },
        )

    def on_state(self, task_id: str, state: str, message: str) -> None:
        self._schedule(
            task_id,
            {
                "type": "agent_state",
                "task_id": task_id,
                "state": state,
                "message": message,
            },
        )

    def _schedule(self, task_id: str, payload: dict[str, Any]) -> None:
        event_type = str(payload.get("type", "unknown"))
        coro = self._go_client.send_stream_event(payload=payload, task_id=task_id)
        try:
            future = asyncio.run_coroutine_threadsafe(coro, self._loop)
        except RuntimeError:
            coro.close()
            logger.warning(
                "Failed to schedule stream event; task_id=%s event_type=%s",
                task_id,
                event_type,
                exc_info=True,
            )
            return

        def _ignore_result(done_future):
            try:
                done_future.result()
            except Exception:
                logger.warning(
                    "Stream event delivery failed; task_id=%s event_type=%s",
                    task_id,
                    event_type,
                    exc_info=True,
                )
                return

        future.add_done_callback(_ignore_result)


def create_checkpointer(settings: Settings):
    try:
        redis_client = Redis.from_url(settings.redis_url, decode_responses=True)
        redis_client.ping()
        return RedisCheckpointer(redis_client)
    except Exception:
        return InMemoryCheckpointer()


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    settings = Settings.from_env()
    app.state.settings = settings
    app.state.loop = asyncio.get_running_loop()
    app.state.go_client = GoToolClient(
        base_url=settings.go_tool_server_url,
        internal_token=settings.go_internal_token,
        timeout_seconds=settings.go_client_timeout_seconds,
    )
    app.state.stream_emitter = GoBridgeStreamEmitter(go_client=app.state.go_client, loop=app.state.loop)
    app.state.checkpointer = create_checkpointer(settings)
    app.state.graph = build_agent_graph(
        checkpointer=app.state.checkpointer,
        go_client=app.state.go_client,
        stream_emitter=app.state.stream_emitter,
    )
    try:
        yield
    finally:
        await app.state.go_client.close()


app = FastAPI(title="LLM Agent Service", lifespan=lifespan)


async def invoke_graph_async(graph, state: dict) -> dict:
    return await asyncio.to_thread(graph.invoke, state)


@app.get("/health")
async def health() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/api/agent/tasks", response_model=CreateTaskResponse)
async def create_task(request: CreateTaskRequest) -> CreateTaskResponse:
    task_id = str(uuid4())
    state = default_agent_state(task_id=task_id)
    state["messages"] = [request.prompt]
    result = await invoke_graph_async(app.state.graph, state)
    status = "awaiting_approval" if result.get("awaiting_approval", False) else "running"
    if result.get("task_complete", False):
        status = "completed"
    return CreateTaskResponse(task_id=task_id, status=status)


@app.post("/api/agent/tasks/{task_id}/message", response_model=AgentResponse)
async def send_task_message(task_id: str, request: TaskMessageRequest) -> AgentResponse:
    existing_state = app.state.checkpointer.get(task_id)
    if existing_state is None:
        raise HTTPException(status_code=404, detail="task not found")

    messages = list(existing_state.get("messages", []))
    messages.append(request.message)
    existing_state["messages"] = messages
    existing_state["approval_granted"] = False
    existing_state["approved_command"] = ""
    result = await invoke_graph_async(app.state.graph, existing_state)

    return AgentResponse(
        ok=True,
        message="message accepted",
        data={
            "task_id": task_id,
            "awaiting_approval": result.get("awaiting_approval", False),
            "task_complete": result.get("task_complete", False),
            "final_response": result.get("final_response", ""),
        },
    )


@app.post("/api/agent/tasks/{task_id}/approve", response_model=AgentResponse)
async def approve_task(task_id: str, request: ApproveTaskRequest) -> AgentResponse:
    existing_state = app.state.checkpointer.get(task_id)
    if existing_state is None:
        raise HTTPException(status_code=404, detail="task not found")

    existing_state["approval_granted"] = bool(request.approved)
    existing_state["approved_command"] = existing_state.get("current_command", "") if request.approved else ""
    if not request.approved:
        existing_state["awaiting_approval"] = False
        existing_state["task_complete"] = True
        existing_state["final_response"] = request.reason or "Approval denied by user."
        app.state.checkpointer.put(checkpoint_key(task_id), existing_state)
        return AgentResponse(ok=True, message="approval rejected", data={"task_id": task_id})

    result = await invoke_graph_async(app.state.graph, existing_state)
    return AgentResponse(
        ok=True,
        message="approval accepted",
        data={
            "task_id": task_id,
            "awaiting_approval": result.get("awaiting_approval", False),
            "task_complete": result.get("task_complete", False),
            "final_response": result.get("final_response", ""),
        },
    )
