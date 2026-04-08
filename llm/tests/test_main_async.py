import asyncio
from unittest.mock import patch

from llm.agent_service.main import invoke_graph_async


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
