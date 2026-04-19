import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Settings:
    mcp_server_url: str
    redis_url: str
    daily_limit: int
    apis_config_path: str

    @classmethod
    def from_env(cls) -> "Settings":
        mcp_server_url = os.getenv("MCP_SERVER_URL", "").strip()
        redis_url = os.getenv("REDIS_URL", "redis://localhost:6379/0").strip()
        daily_limit = int(os.getenv("FREE_API_DAILY_LIMIT", "500"))
        apis_config_path = os.getenv("APIS_CONFIG_PATH", "apis.json").strip()

        if not mcp_server_url:
            raise RuntimeError("Missing required environment variable: MCP_SERVER_URL")

        return cls(
            mcp_server_url=mcp_server_url,
            redis_url=redis_url,
            daily_limit=daily_limit,
            apis_config_path=apis_config_path,
        )
