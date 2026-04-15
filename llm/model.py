import asyncio
from dataclasses import dataclass
from typing import Any

from pydantic import AliasChoices, BaseModel, ConfigDict, Field

from graph import AgentState


class NewConversationRequest(BaseModel):
    model_config = ConfigDict(populate_by_name=True)
    ssh_session_id: str = Field(alias="sshSessionId")


class NewConversationResponse(BaseModel):
    conversation_id: str = Field(alias="conversationId")


class CloseRequest(BaseModel):
    model_config = ConfigDict(populate_by_name=True)
    ssh_session_id: str = Field(alias="sshSessionId")


class CloseResponse(BaseModel):
    closed_conversations: int = Field(alias="closedConversations")


class ChatRequest(BaseModel):
    model_config = ConfigDict(populate_by_name=True)
    conversation_id: str = Field(
        validation_alias=AliasChoices("conversation_id", "conversationId")
    )
    message: str | None = None
    approval_granted: bool | None = None


@dataclass
class ConversationRuntime:
    graph: Any
    state: AgentState
    lock: asyncio.Lock