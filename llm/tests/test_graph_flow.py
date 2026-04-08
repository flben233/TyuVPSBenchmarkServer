from llm.agent_service.graph import CHECKPOINT_TTL_SECONDS, InMemoryCheckpointer, RedisCheckpointer, build_agent_graph


def test_dangerous_command_routes_to_awaiting_approval() -> None:
    checkpointer = InMemoryCheckpointer()
    workflow = build_agent_graph(checkpointer=checkpointer)

    initial_state = {
        "task_id": "task-danger-1",
        "user_id": "user-1",
        "messages": ["run: rm -rf /tmp/demo"],
    }

    result = workflow.invoke(initial_state)

    assert result["safety_status"] == "dangerous"
    assert result["awaiting_approval"] is True
    assert result["task_complete"] is False
    assert result["current_command"] == "rm -rf /tmp/demo"
    assert checkpointer.get("task-danger-1") is not None


def test_retry_limit_terminates_with_failure_response() -> None:
    checkpointer = InMemoryCheckpointer()

    def always_retry(_command: str) -> str:
        return "retry requested"

    workflow = build_agent_graph(
        checkpointer=checkpointer,
        command_executor=always_retry,
        max_retries=2,
    )

    result = workflow.invoke(
        {
            "task_id": "task-retry-1",
            "user_id": "user-1",
            "messages": ["run: ls -la"],
        }
    )

    assert result["task_complete"] is True
    assert result["retry_count"] == 2
    assert "retry limit" in result["final_response"].lower()


def test_approval_is_reset_after_approved_dangerous_execution() -> None:
    checkpointer = InMemoryCheckpointer()
    workflow = build_agent_graph(checkpointer=checkpointer)

    first = workflow.invoke(
        {
            "task_id": "task-approval-reset-1",
            "user_id": "user-1",
            "messages": ["run: rm -rf /tmp/demo"],
        }
    )
    assert first["awaiting_approval"] is True

    first["approval_granted"] = True
    approved = workflow.invoke(first)

    assert approved["task_complete"] is True
    assert approved["approval_granted"] is False

    approved_messages = list(approved["messages"])
    approved_messages.append("run: shutdown now")
    approved["messages"] = approved_messages

    next_result = workflow.invoke(approved)
    assert next_result["awaiting_approval"] is True


def test_go_client_is_called_for_safety_and_execute_with_task_id() -> None:
    class _FakeResponse:
        def __init__(self, payload):
            self._payload = payload

        def json(self):
            return self._payload

    class _StubGoClient:
        def __init__(self) -> None:
            self.calls = []

        def post_sync(self, path, payload, task_id):
            self.calls.append((path, payload, task_id))
            if path == "/api/agent/safety-check":
                return _FakeResponse(
                    {
                        "data": {
                            "risk_level": "safe",
                            "requires_approval": False,
                        }
                    }
                )
            if path == "/api/agent/execute":
                return _FakeResponse(
                    {
                        "data": {
                            "status": "running",
                            "message": "command dispatched",
                        }
                    }
                )
            raise AssertionError(f"unexpected path: {path}")

    checkpointer = InMemoryCheckpointer()
    go_client = _StubGoClient()
    workflow = build_agent_graph(checkpointer=checkpointer, go_client=go_client)

    result = workflow.invoke(
        {
            "task_id": "task-go-client-1",
            "user_id": "user-1",
            "messages": ["run: ls -la"],
        }
    )

    assert result["task_complete"] is True
    assert result["command_result"] == "command dispatched"
    assert go_client.calls[0] == (
        "/api/agent/safety-check",
        {"command": "ls -la"},
        "task-go-client-1",
    )
    assert go_client.calls[1] == (
        "/api/agent/execute",
        {"command": "ls -la", "approved": False},
        "task-go-client-1",
    )


def test_redis_checkpointer_sets_checkpoint_ttl_on_save() -> None:
    class _StubRedis:
        def __init__(self) -> None:
            self.last_set = None

        def set(self, key, value, ex=None):
            self.last_set = (key, value, ex)

    redis_client = _StubRedis()
    checkpointer = RedisCheckpointer(redis_client)

    checkpointer.put("checkpoint:task-ttl-1", {"task_id": "task-ttl-1", "messages": ["run: ls"]})

    assert redis_client.last_set is not None
    assert redis_client.last_set[0] == "checkpoint:task-ttl-1"
    assert redis_client.last_set[2] == CHECKPOINT_TTL_SECONDS
