import asyncio
import threading
from dataclasses import dataclass, field
from typing import Any

from pydantic import AliasChoices, BaseModel, ConfigDict, Field

from graph import AgentState


class NewConversationRequest(BaseModel):
    """
    Request body for creating a new LLM agent conversation.
    """
    model_config = ConfigDict(populate_by_name=True)
    ssh_session_id: str = Field(
        alias="sshSessionId",
        description="The SSH session ID to associate with this conversation",
    )


class NewConversationResponse(BaseModel):
    """
    Response body returned after creating a new conversation.
    """
    conversation_id: str = Field(
        alias="conversationId",
        description="The unique identifier for the newly created conversation",
    )


class CloseRequest(BaseModel):
    """
    Request body for closing all conversations associated with an SSH session.
    """
    model_config = ConfigDict(populate_by_name=True)
    ssh_session_id: str = Field(
        alias="sshSessionId",
        description="The SSH session ID whose conversations should be closed",
    )


class CloseResponse(BaseModel):
    """
    Response body returned after closing conversations.
    """
    closed_conversations: int = Field(
        alias="closedConversations",
        description="Number of conversations that were successfully closed",
    )


class ChatMessage(BaseModel):
    """
    A single message in the conversation history sent from the frontend.
    """
    role: str = Field(description="Message role: 'user', 'assistant', or 'system'")
    content: str = Field(default="", description="Text content of the message")


class ChatRequest(BaseModel):
    """
    Request body for sending a message to the LLM agent.
    The frontend is responsible for managing conversation history (messages array).
    Messages are set into the session state on arrival and cleared after the response completes.
    """
    model_config = ConfigDict(populate_by_name=True)
    conversation_id: str = Field(
        validation_alias=AliasChoices("conversation_id", "conversationId"),
        description="The conversation ID to send the message to",
    )
    message: str | None = Field(
        default=None,
        description="New user message text. If provided, it is appended to the conversation history.",
    )
    messages: list[ChatMessage] | None = Field(
        default=None,
        description="Full conversation history from the frontend. "
                    "Each item has 'role' and 'content'. "
                    "This replaces the session's message state on the backend for the duration of the request.",
    )
    approval_granted: bool | None = Field(
        default=None,
        description="Whether the user approved a pending dangerous command. "
                    "True=approve, False=reject, None=not an approval response.",
    )


class StopRequest(BaseModel):
    """
    Request body for stopping an in-progress LLM response.
    """
    model_config = ConfigDict(populate_by_name=True)
    conversation_id: str = Field(
        validation_alias=AliasChoices("conversation_id", "conversationId"),
        description="The conversation ID to stop",
    )


class StopResponse(BaseModel):
    """
    Response body returned after stopping a conversation.
    """
    stopped: bool = Field(description="Whether the conversation was successfully stopped")


@dataclass
class ConversationRuntime:
    graph: Any
    state: AgentState
    lock: asyncio.Lock
    stop_event: threading.Event = field(default_factory=threading.Event)