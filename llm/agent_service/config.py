from __future__ import annotations

import os

from pydantic import BaseModel, Field, field_validator


class Settings(BaseModel):
    go_tool_server_url: str = Field(default="http://localhost:8080", alias="GO_TOOL_SERVER_URL")
    go_internal_token: str = Field(default="", alias="GO_INTERNAL_TOKEN")
    redis_url: str = Field(default="redis://localhost:6379/0", alias="REDIS_URL")
    openai_api_key: str = Field(default="", alias="OPENAI_API_KEY")
    openai_api_base: str = Field(default="", alias="OPENAI_API_BASE")
    openai_model: str = Field(default="gpt-4o-mini", alias="OPENAI_MODEL")
    go_client_timeout_seconds: float = Field(default=10.0, alias="GO_CLIENT_TIMEOUT_SECONDS")

    @field_validator("go_tool_server_url", "go_internal_token", "redis_url")
    @classmethod
    def _must_not_be_empty(cls, value: str) -> str:
        if not value or not value.strip():
            raise ValueError("must not be empty")
        return value

    @classmethod
    def from_env(cls) -> "Settings":
        return cls(
            GO_TOOL_SERVER_URL=os.getenv("GO_TOOL_SERVER_URL", "http://localhost:8080"),
            GO_INTERNAL_TOKEN=os.getenv("GO_INTERNAL_TOKEN", ""),
            REDIS_URL=os.getenv("REDIS_URL", "redis://localhost:6379/0"),
            OPENAI_API_KEY=os.getenv("OPENAI_API_KEY", ""),
            OPENAI_API_BASE=os.getenv("OPENAI_API_BASE", ""),
            OPENAI_MODEL=os.getenv("OPENAI_MODEL", "gpt-4o-mini"),
            GO_CLIENT_TIMEOUT_SECONDS=os.getenv("GO_CLIENT_TIMEOUT_SECONDS", "10"),
        )
