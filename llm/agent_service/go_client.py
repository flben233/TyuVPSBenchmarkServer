from __future__ import annotations

from typing import Any, Optional

import httpx


DEFAULT_TIMEOUT_SECONDS = 10.0


class GoToolClient:
    def __init__(
        self,
        base_url: str,
        internal_token: str,
        timeout_seconds: float = DEFAULT_TIMEOUT_SECONDS,
        client: Optional[httpx.AsyncClient] = None,
    ) -> None:
        self.base_url = base_url.rstrip("/")
        self.internal_token = internal_token
        self.timeout_seconds = timeout_seconds
        self._client = client or httpx.AsyncClient(base_url=self.base_url, timeout=self.timeout_seconds)
        self._owns_client = client is None

    async def close(self) -> None:
        if self._owns_client:
            await self._client.aclose()

    async def post(self, path: str, payload: dict[str, Any], task_id: str) -> httpx.Response:
        url = self._build_url(path)
        headers = self._build_headers(task_id)
        try:
            response = await self._client.post(
                url,
                json=payload,
                headers=headers,
                timeout=self.timeout_seconds,
            )
            response.raise_for_status()
            return response
        except httpx.HTTPStatusError as exc:
            status_code = exc.response.status_code if exc.response is not None else "unknown"
            raise RuntimeError(f"go tool request failed with status {status_code} for {url}") from exc
        except httpx.RequestError as exc:
            raise RuntimeError(f"go tool request error for {url}: {exc}") from exc

    async def send_stream_event(self, payload: dict[str, Any], task_id: str) -> httpx.Response:
        return await self.post("/api/agent/stream-event", payload, task_id)

    def post_sync(self, path: str, payload: dict[str, Any], task_id: str) -> httpx.Response:
        url = self._build_url(path)
        headers = self._build_headers(task_id)
        try:
            response = httpx.post(
                url,
                json=payload,
                headers=headers,
                timeout=self.timeout_seconds,
            )
            response.raise_for_status()
            return response
        except httpx.HTTPStatusError as exc:
            status_code = exc.response.status_code if exc.response is not None else "unknown"
            raise RuntimeError(f"go tool request failed with status {status_code} for {url}") from exc
        except httpx.RequestError as exc:
            raise RuntimeError(f"go tool request error for {url}: {exc}") from exc

    def _build_url(self, path: str) -> str:
        if path.startswith("http://") or path.startswith("https://"):
            return path
        if not path.startswith("/"):
            path = f"/{path}"
        return f"{self.base_url}{path}"

    def _build_headers(self, task_id: str) -> dict[str, str]:
        return {
            "X-Internal-Token": self.internal_token,
            "X-Task-ID": task_id,
        }
