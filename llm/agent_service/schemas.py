from __future__ import annotations

from typing import Any, Dict, Optional

from pydantic import BaseModel, Field


class CreateTaskRequest(BaseModel):
    prompt: str = Field(min_length=1)
    metadata: Dict[str, Any] = Field(default_factory=dict)


class CreateTaskResponse(BaseModel):
    task_id: str
    status: str = "pending"


class TaskMessageRequest(BaseModel):
    message: str = Field(min_length=1)


class ApproveTaskRequest(BaseModel):
    approved: bool = True
    reason: Optional[str] = None


class AgentResponse(BaseModel):
    ok: bool = True
    message: str = "accepted"
    data: Dict[str, Any] = Field(default_factory=dict)
