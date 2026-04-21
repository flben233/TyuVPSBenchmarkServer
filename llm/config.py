import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Settings:
    mcp_server_url: str
    redis_url: str
    daily_limit: int
    apis_config_path: str
    context_tail_keep: int
    compress_threshold_tokens: int
    max_chat_message_chars: int

    @classmethod
    def from_env(cls) -> "Settings":
        mcp_server_url = os.getenv("MCP_SERVER_URL", "").strip()
        redis_url = os.getenv("REDIS_URL", "redis://localhost:6379/0").strip()
        daily_limit = int(os.getenv("FREE_API_DAILY_LIMIT", "500"))
        apis_config_path = os.getenv("APIS_CONFIG_PATH", "apis.json").strip()
        context_tail_keep = int(os.getenv("CONTEXT_TAIL_KEEP", "6"))
        compress_threshold_tokens = int(os.getenv("COMPRESS_THRESHOLD_TOKENS", "6000"))
        max_chat_message_chars = int(os.getenv("MAX_CHAT_MESSAGE_CHARS", "4000"))

        if not mcp_server_url:
            raise RuntimeError("Missing required environment variable: MCP_SERVER_URL")

        return cls(
            mcp_server_url=mcp_server_url,
            redis_url=redis_url,
            daily_limit=daily_limit,
            apis_config_path=apis_config_path,
            context_tail_keep=max(1, context_tail_keep),
            compress_threshold_tokens=max(1, compress_threshold_tokens),
            max_chat_message_chars=max(1, max_chat_message_chars),
        )
