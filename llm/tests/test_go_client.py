import asyncio

import httpx

from llm.agent_service.go_client import GoToolClient


def test_post_adds_internal_and_task_headers() -> None:
    captured = {}

    def handler(request: httpx.Request) -> httpx.Response:
        captured["internal"] = request.headers.get("X-Internal-Token")
        captured["task"] = request.headers.get("X-Task-ID")
        return httpx.Response(200, json={"ok": True})

    transport = httpx.MockTransport(handler)
    async def run_test() -> None:
        async with httpx.AsyncClient(transport=transport) as client:
            go_client = GoToolClient(
                base_url="http://go-tool.local",
                internal_token="token-123",
                client=client,
            )
            await go_client.post("/api/tool/run", {"input": "hello"}, task_id="task-42")

    asyncio.run(run_test())

    assert captured["internal"] == "token-123"
    assert captured["task"] == "task-42"


def test_post_normalizes_relative_path_without_leading_slash() -> None:
    captured = {}

    def handler(request: httpx.Request) -> httpx.Response:
        captured["url"] = str(request.url)
        return httpx.Response(200, json={"ok": True})

    transport = httpx.MockTransport(handler)

    async def run_test() -> None:
        async with httpx.AsyncClient(transport=transport) as client:
            go_client = GoToolClient(
                base_url="http://go-tool.local/",
                internal_token="token-123",
                client=client,
            )
            await go_client.post("api/tool/run", {"input": "hello"}, task_id="task-42")

    asyncio.run(run_test())

    assert captured["url"] == "http://go-tool.local/api/tool/run"


def test_post_raises_runtime_error_on_non_2xx() -> None:
    def handler(_request: httpx.Request) -> httpx.Response:
        return httpx.Response(500, json={"error": "upstream failure"})

    transport = httpx.MockTransport(handler)

    async def run_test() -> None:
        async with httpx.AsyncClient(transport=transport) as client:
            go_client = GoToolClient(
                base_url="http://go-tool.local",
                internal_token="token-123",
                client=client,
            )
            await go_client.post("/api/tool/run", {"input": "hello"}, task_id="task-42")

    try:
        asyncio.run(run_test())
        raise AssertionError("expected RuntimeError for non-2xx response")
    except RuntimeError as exc:
        assert "500" in str(exc)


def test_post_raises_runtime_error_on_request_failure() -> None:
    def handler(request: httpx.Request) -> httpx.Response:
        raise httpx.ConnectError("connection failed", request=request)

    transport = httpx.MockTransport(handler)

    async def run_test() -> None:
        async with httpx.AsyncClient(transport=transport) as client:
            go_client = GoToolClient(
                base_url="http://go-tool.local",
                internal_token="token-123",
                client=client,
            )
            await go_client.post("/api/tool/run", {"input": "hello"}, task_id="task-42")

    try:
        asyncio.run(run_test())
        raise AssertionError("expected RuntimeError for request failure")
    except RuntimeError as exc:
        assert "request" in str(exc).lower()
