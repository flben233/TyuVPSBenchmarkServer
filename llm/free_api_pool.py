from __future__ import annotations

import json
import logging
import os
import time
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

from openai import OpenAI

logger = logging.getLogger(__name__)


@dataclass
class APIEndpoint:
    """A single OpenAI-compatible API endpoint configuration."""

    api_base: str
    api_key: str
    model: str
    cooldown_seconds: int = 300

    _client: OpenAI | None = field(default=None, init=False, repr=False)

    @property
    def client(self) -> OpenAI:
        if self._client is None:
            self._client = OpenAI(
                base_url=self.api_base,
                api_key=self.api_key,
            )
        return self._client


@dataclass
class _CooldownState:
    until: float


class FreeAPIPool:
    """Manages a pool of free API endpoints with per-user rate limiting,
    automatic failover, and cooldown tracking via Redis.

    API endpoints are loaded from a JSON config file (``apis.json`` by default).
    Rate limiting uses a sliding-window counter stored in Redis.

    Environment variables
    ---------------------
    ``REDIS_URL``           Redis connection URL (default ``redis://localhost:6379/0``).
    ``FREE_API_DAILY_LIMIT`` Max requests per user per day (default ``20``).
    ``APIS_CONFIG_PATH``     Path to the JSON config file (default ``apis.json``).
    """

    def __init__(self) -> None:
        self._redis_url = os.getenv("REDIS_URL", "redis://localhost:6379/0")
        self._redis_password = os.getenv("REDIS_PASS", "")
        self._rate_limit = int(os.getenv("FREE_API_DAILY_LIMIT", "20"))
        self._endpoints: list[APIEndpoint] = []
        # In-memory cooldown tracking: endpoint index -> CooldownState
        self._cooldowns: dict[int, _CooldownState] = {}

        self._redis: Any = None
        self._load_config()

    # ------------------------------------------------------------------
    # Config loading
    # ------------------------------------------------------------------

    def _load_config(self) -> None:
        config_path = os.getenv("APIS_CONFIG_PATH", "apis.json")
        path = Path(config_path)
        if not path.is_absolute():
            path = Path(__file__).resolve().parent / path

        if not path.exists():
            logger.warning("API config file not found at %s – free API pool is empty", path)
            return

        try:
            raw = json.loads(path.read_text(encoding="utf-8"))
        except (json.JSONDecodeError, OSError) as exc:
            logger.error("Failed to read API config %s: %s", path, exc)
            return

        if not isinstance(raw, list):
            logger.error("API config %s must be a JSON array", path)
            return

        for idx, entry in enumerate(raw):
            try:
                ep = APIEndpoint(
                    api_base=entry["apiBase"],
                    api_key=entry["apiKey"],
                    model=entry["model"],
                    cooldown_seconds=int(entry.get("cooldownSeconds", 300)),
                )
                self._endpoints.append(ep)
            except (KeyError, TypeError, ValueError) as exc:
                logger.warning("Skipping API entry #%d: %s", idx, exc)

        logger.info("Loaded %d API endpoint(s) from %s", len(self._endpoints), path)

    def reload(self) -> None:
        self._endpoints.clear()
        self._cooldowns.clear()
        self._load_config()

    @property
    def has_endpoints(self) -> bool:
        return len(self._endpoints) > 0

    # ------------------------------------------------------------------
    # Redis connection (lazy)
    # ------------------------------------------------------------------

    def _get_redis(self):
        if self._redis is not None:
            return self._redis
        try:
            import redis
            kwargs = {"decode_responses": True}
            if self._redis_password:
                kwargs["password"] = self._redis_password
            self._redis = redis.from_url(self._redis_url, **kwargs)
            self._redis.ping()
            logger.info("Connected to Redis at %s", self._redis_url)
            return self._redis
        except Exception as exc:
            logger.error("Redis connection failed (%s) – rate limiting disabled", exc)
            self._redis = None
            return None

    # ------------------------------------------------------------------
    # Rate limiting – daily counter
    # ------------------------------------------------------------------

    def _daily_key(self, user_id: str) -> str:
        today = time.strftime("%Y-%m-%d", time.localtime())
        return f"free_api:daily:{user_id}:{today}"

    def check_rate_limit(self, user_id: str) -> bool:
        """Return True if the user is within their daily limit."""
        r = self._get_redis()
        if r is None:
            return True

        count = r.get(self._daily_key(user_id))
        if count is None:
            return True
        return int(count) < self._rate_limit

    def record_request(self, user_id: str) -> None:
        """Record one request for daily rate-limit tracking."""
        r = self._get_redis()
        if r is None:
            return

        key = self._daily_key(user_id)
        pipe = r.pipeline()
        pipe.incr(key)
        # Expire at end of day (next midnight + small buffer)
        pipe.expire(key, 86400 - int(time.time() % 86400) + 60)
        pipe.execute()

    def get_remaining(self, user_id: str) -> int:
        """Return how many requests the user has left today."""
        r = self._get_redis()
        if r is None:
            return self._rate_limit

        count = r.get(self._daily_key(user_id))
        if count is None:
            return self._rate_limit
        return max(0, self._rate_limit - int(count))

    # ------------------------------------------------------------------
    # Endpoint selection with cooldown
    # ------------------------------------------------------------------

    def _is_cooled_down(self, idx: int) -> bool:
        cd = self._cooldowns.get(idx)
        if cd is None:
            return True
        if time.time() >= cd.until:
            del self._cooldowns[idx]
            return True
        return False

    def set_cooldown(self, idx: int) -> None:
        """Put endpoint *idx* into cooldown."""
        if 0 <= idx < len(self._endpoints):
            ep = self._endpoints[idx]
            self._cooldowns[idx] = _CooldownState(until=time.time() + ep.cooldown_seconds)
            logger.warning(
                "API endpoint #%d (%s) entered cooldown for %ds",
                idx, ep.api_base, ep.cooldown_seconds,
            )

    def acquire(self) -> tuple[OpenAI, str, int] | None:
        """Pick the first available (non-cooled-down) endpoint.

        Returns ``(client, model, endpoint_index)`` or ``None`` if all are
        in cooldown or the pool is empty.
        """
        for idx, ep in enumerate(self._endpoints):
            if self._is_cooled_down(idx):
                return ep.client, ep.model, idx
        return None

    # ------------------------------------------------------------------
    # High-level: acquire with rate-limit check
    # ------------------------------------------------------------------

    def acquire_for_user(self, user_id: str) -> tuple[OpenAI, str, int] | None:
        """Try to acquire an endpoint for *user_id*.

        Returns ``(client, model, endpoint_index)`` or ``None`` when:
        - rate limit exceeded
        - all endpoints are in cooldown
        - pool is empty
        """
        if not self.has_endpoints:
            return None
        if not self.check_rate_limit(user_id):
            logger.info("User %s hit free API rate limit", user_id)
            return None
        return self.acquire()

    @property
    def num_endpoints(self) -> int:
        return len(self._endpoints)


class FreeAPIExhaustedError(Exception):
    """Raised when all free API endpoints are in cooldown or unavailable."""


class FailoverClient:
    """Drop-in replacement for ``openai.OpenAI`` with automatic failover.

    Mimics the ``client.chat.completions.create()`` interface so that
    ``graph.py`` can use it without any changes.  On 429 / rate-limit
    errors the current endpoint is put into cooldown and the next
    available one is tried transparently.

    Usage::

        fc = FailoverClient(pool, user_id)
        fc.acquire()                          # initial acquisition + rate check
        stream = fc.chat.completions.create(  # transparent failover on 429
            model=..., messages=..., stream=True,
        )
    """

    def __init__(self, pool: FreeAPIPool, user_id: str) -> None:
        self._pool = pool
        self._user_id = user_id
        self._client: OpenAI | None = None
        self._model: str = ""
        self._idx: int = -1

    def acquire(self) -> None:
        """Initial acquisition with rate-limit check.

        Raises ``FreeAPIExhaustedError`` when no endpoint is available.
        """
        result = self._pool.acquire_for_user(self._user_id)
        if result is None:
            raise FreeAPIExhaustedError(
                "Free API rate limit exceeded or all endpoints are in cooldown"
            )
        self._client, self._model, self._idx = result
        self._pool.record_request(self._user_id)
        logger.info(
            "User %s acquired endpoint #%d (%s) model=%s",
            self._user_id, self._idx, self._client.base_url, self._model,
        )

    @property
    def base_url(self) -> str:
        return self._client.base_url if self._client else ""

    # ---- Mimic openai.OpenAI interface: client.chat.completions.create() ----

    @property
    def chat(self) -> FailoverClient:
        return self

    @property
    def completions(self) -> FailoverClient:
        return self

    def create(self, **kwargs: Any) -> Any:
        """Drop-in for ``client.chat.completions.create()``.

        On retryable errors (429 / rate-limit / quota) the current endpoint
        is put into cooldown and the next available one is acquired.
        Retries at most ``num_endpoints`` times, then re-raises.
        """
        max_retries = max(self._pool.num_endpoints, 1)
        for attempt in range(max_retries):
            try:
                kwargs["model"] = self._model
                return self._client.chat.completions.create(**kwargs)
            except Exception as exc:
                if not self._is_retryable(exc):
                    raise
                logger.warning(
                    "Endpoint #%d failed (attempt %d/%d): %s",
                    self._idx, attempt + 1, max_retries, exc,
                )
                self._pool.set_cooldown(self._idx)
                next_result = self._pool.acquire()
                if next_result is None:
                    raise FreeAPIExhaustedError(
                        "All free API endpoints are in cooldown"
                    ) from exc
                self._client, self._model, self._idx = next_result
                logger.info(
                    "Failover to endpoint #%d (%s) model=%s",
                    self._idx, self._client.base_url, self._model,
                )
        raise FreeAPIExhaustedError("All free API endpoints exhausted after retries")

    @staticmethod
    def _is_retryable(exc: Exception) -> bool:
        status = getattr(exc, "status_code", None)
        if status == 429:
            return True
        msg = str(exc).lower()
        return "429" in str(exc) or "rate" in msg or "quota" in msg
