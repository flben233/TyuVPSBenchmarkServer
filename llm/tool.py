from typing import Any

from langchain_core.messages import (
    ToolMessage,
)
from langchain_core.tools import BaseTool


async def load_mcp_tools(mcp_url: str) -> list[BaseTool]:
    from langchain_mcp_adapters.client import MultiServerMCPClient

    client = MultiServerMCPClient(
        {"ssh": {"url": mcp_url, "transport": "streamable_http"}}
    )
    return await client.get_tools()



def _normalize_json_schema(schema: dict[str, Any] | None) -> dict[str, Any]:
    if not isinstance(schema, dict):
        return {"type": "object", "properties": {}}

    normalized = dict(schema)
    normalized.pop("$schema", None)
    if "type" not in normalized:
        normalized["type"] = "object"
    if normalized.get("type") == "object" and "properties" not in normalized:
        normalized["properties"] = {}
    return normalized


def _tool_to_openai_function(tool: BaseTool) -> dict[str, Any]:
    input_schema = _normalize_json_schema(tool.get_input_jsonschema())
    return {
        "type": "function",
        "function": {
            "name": tool.name,
            "description": tool.description or "",
            "parameters": input_schema,
        },
    }


def build_openai_tools(tools: list[BaseTool]) -> list[dict[str, Any]]:
    return [_tool_to_openai_function(tool) for tool in tools]


def inject_session_id(tool_call: dict, ssh_session_id: str) -> dict:
    args = dict(tool_call.get("args", {}))
    if "session_id" not in args and ssh_session_id:
        args["session_id"] = ssh_session_id
    return {**tool_call, "args": args}


def invoke_tool(tool_call: dict, tool_map: dict[str, BaseTool]) -> ToolMessage:
    name = tool_call["name"]
    args = tool_call.get("args", {})
    tool = tool_map.get(name)
    try:
        result = tool.invoke(args) if tool else f"Error: unknown tool '{name}'"
    except Exception as e:
        result = f"Error executing {name}: {e}"
    return ToolMessage(
        content=str(result),
        tool_call_id=tool_call["id"],
        name=name,
    )