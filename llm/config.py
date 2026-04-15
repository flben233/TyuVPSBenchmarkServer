import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Settings:
    mcp_server_url: str
    openai_api_base: str
    openai_api_key: str
    openai_model: str

    @classmethod
    def from_env(cls) -> "Settings":
        mcp_server_url = os.getenv("MCP_SERVER_URL", "").strip()
        openai_api_base = os.getenv("OPENAI_API_BASE", "").strip()
        openai_api_key = os.getenv("OPENAI_API_KEY", "").strip()
        openai_model = os.getenv("OPENAI_MODEL", "").strip()

        missing = []
        if not mcp_server_url:
            missing.append("MCP_SERVER_URL")
        if not openai_api_base:
            missing.append("OPENAI_API_BASE")
        if not openai_api_key:
            missing.append("OPENAI_API_KEY")
        if not openai_model:
            missing.append("OPENAI_MODEL")

        if missing:
            raise RuntimeError(
                "Missing required environment variables: " + ", ".join(missing)
            )

        return cls(
            mcp_server_url=mcp_server_url,
            openai_api_base=openai_api_base,
            openai_api_key=openai_api_key,
            openai_model=openai_model,
        )
