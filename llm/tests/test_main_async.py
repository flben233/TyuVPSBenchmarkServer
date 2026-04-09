import asyncio
from types import SimpleNamespace
from unittest.mock import patch

from fastapi import FastAPI

from llm.agent_service.main import GoBridgeStreamEmitter, invoke_graph_async, lifespan


def test_invoke_graph_async_uses_to_thread_wrapper() -> None:
    class _StubGraph:
        def invoke(self, state):
            return {"task_id": state["task_id"], "ok": True}

    graph = _StubGraph()
    captured = {}

    async def _fake_to_thread(func, *args, **kwargs):
        captured["func"] = func
        captured["args"] = args
        captured["kwargs"] = kwargs
        return func(*args, **kwargs)

    with patch("llm.agent_service.main.asyncio.to_thread", new=_fake_to_thread):
        result = asyncio.run(invoke_graph_async(graph, {"task_id": "task-main-1"}))

    assert result == {"task_id": "task-main-1", "ok": True}
    assert captured["func"] == graph.invoke
    assert captured["args"] == ({"task_id": "task-main-1"},)
    assert captured["kwargs"] == {}


def test_go_bridge_stream_emitter_schedules_event_without_waiting() -> None:
    captured = {}

    class _StubGoClient:
        async def send_stream_event(self, payload, task_id):
            return None

    class _StubFuture:
        def add_done_callback(self, callback):
            captured["callback"] = callback

    class _StubLoop:
        pass

    def _fake_run_coroutine_threadsafe(coro, loop):
        captured["coro"] = coro
        captured["loop"] = loop
        return _StubFuture()

    emitter = GoBridgeStreamEmitter(go_client=_StubGoClient(), loop=_StubLoop())

    with patch("llm.agent_service.main.asyncio.run_coroutine_threadsafe", new=_fake_run_coroutine_threadsafe):
        emitter.on_state("task-main-stream-1", "thinking", "Parsing user intent.")

    assert captured["loop"].__class__.__name__ == "_StubLoop"
    assert captured["callback"]
    captured["coro"].close()


def test_go_bridge_stream_emitter_builds_expected_payloads() -> None:
    captured_calls = []

    class _StubGoClient:
        async def send_stream_event(self, payload, task_id):
            captured_calls.append({"payload": payload, "task_id": task_id})
            return None

    class _StubFuture:
        def add_done_callback(self, callback):
            callback(self)

        def result(self):
            return None

    class _StubLoop:
        pass

    def _fake_run_coroutine_threadsafe(coro, _loop):
        asyncio.run(coro)
        return _StubFuture()

    emitter = GoBridgeStreamEmitter(go_client=_StubGoClient(), loop=_StubLoop())

    with patch("llm.agent_service.main.asyncio.run_coroutine_threadsafe", new=_fake_run_coroutine_threadsafe):
        emitter.on_message_start("task-1", "msg-1")
        emitter.on_token("task-1", "msg-1", "hel")
        emitter.on_message_end("task-1", "msg-1", "stop")
        emitter.on_state("task-1", "thinking", "Planning")

    assert captured_calls == [
        {
            "task_id": "task-1",
            "payload": {"type": "agent_message_start", "task_id": "task-1", "message_id": "msg-1"},
        },
        {
            "task_id": "task-1",
            "payload": {
                "type": "agent_token",
                "task_id": "task-1",
                "message_id": "msg-1",
                "delta": "hel",
            },
        },
        {
            "task_id": "task-1",
            "payload": {
                "type": "agent_message_end",
                "task_id": "task-1",
                "message_id": "msg-1",
                "finish_reason": "stop",
            },
        },
        {
            "task_id": "task-1",
            "payload": {
                "type": "agent_state",
                "task_id": "task-1",
                "state": "thinking",
                "message": "Planning",
            },
        },
    ]


def test_go_bridge_stream_emitter_handles_closed_loop_schedule_failure() -> None:
    class _StubGoClient:
        async def send_stream_event(self, payload, task_id):
            return None

    class _StubLoop:
        pass

    def _raising_run_coroutine_threadsafe(coro, _loop):
        coro.close()
        raise RuntimeError("Event loop is closed")

    emitter = GoBridgeStreamEmitter(go_client=_StubGoClient(), loop=_StubLoop())

    with patch(
        "llm.agent_service.main.asyncio.run_coroutine_threadsafe",
        new=_raising_run_coroutine_threadsafe,
    ), patch("llm.agent_service.main.logger.warning") as warning_mock:
        emitter.on_state("task-loop-closed", "thinking", "message")

    warning_mock.assert_called_once()


def test_go_bridge_stream_emitter_logs_when_future_result_fails() -> None:
    class _StubGoClient:
        async def send_stream_event(self, payload, task_id):
            return None

    class _StubFuture:
        def add_done_callback(self, callback):
            callback(self)

        def result(self):
            raise RuntimeError("bridge send failed")

    class _StubLoop:
        pass

    def _fake_run_coroutine_threadsafe(coro, _loop):
        coro.close()
        return _StubFuture()

    emitter = GoBridgeStreamEmitter(go_client=_StubGoClient(), loop=_StubLoop())

    with patch("llm.agent_service.main.asyncio.run_coroutine_threadsafe", new=_fake_run_coroutine_threadsafe), patch(
        "llm.agent_service.main.logger.warning"
    ) as warning_mock:
        emitter.on_message_end("task-callback-fail", "msg-1", "error")

    warning_mock.assert_called_once()


def test_lifespan_wires_stream_emitter_into_graph_builder() -> None:
    app = FastAPI()
    captured = {}

    class _FakeGoClient:
        def __init__(self, base_url: str, internal_token: str, timeout_seconds: float) -> None:
            captured["go_client_init"] = {
                "base_url": base_url,
                "internal_token": internal_token,
                "timeout_seconds": timeout_seconds,
            }

        async def close(self) -> None:
            return None

    def _fake_build_agent_graph(**kwargs):
        captured["graph_kwargs"] = kwargs
        return object()

    async def run_test() -> None:
        with patch(
            "llm.agent_service.main.Settings.from_env",
            return_value=SimpleNamespace(
                go_tool_server_url="http://go-tool.local",
                go_internal_token="token-123",
                go_client_timeout_seconds=3.5,
                redis_url="redis://localhost:6379/0",
            ),
        ), patch("llm.agent_service.main.GoToolClient", new=_FakeGoClient), patch(
            "llm.agent_service.main.create_checkpointer", return_value="checkpointer"
        ), patch("llm.agent_service.main.build_agent_graph", new=_fake_build_agent_graph):
            async with lifespan(app):
                assert app.state.graph is not None

    asyncio.run(run_test())

    assert captured["graph_kwargs"]["checkpointer"] == "checkpointer"
    assert isinstance(captured["graph_kwargs"]["stream_emitter"], GoBridgeStreamEmitter)
