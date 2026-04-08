from __future__ import annotations

import asyncio
from contextlib import asynccontextmanager
from uuid import uuid4
from typing import AsyncIterator

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
    app.state.go_client = GoToolClient(
        base_url=settings.go_tool_server_url,
        internal_token=settings.go_internal_token,
        timeout_seconds=settings.go_client_timeout_seconds,
    )
    app.state.checkpointer = create_checkpointer(settings)
    app.state.graph = build_agent_graph(
        checkpointer=app.state.checkpointer,
        go_client=app.state.go_client,
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
