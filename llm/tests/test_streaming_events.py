"""Task 4 streaming events tests.

TDD evidence (red -> green):
1) python -m pytest llm/tests/test_streaming_events.py -q
2) python -m pytest llm/tests/test_streaming_events.py -q
"""

from dataclasses import dataclass

from llm.agent_service.graph import InMemoryCheckpointer, build_agent_graph


@dataclass(frozen=True)
class StateEvent:
    task_id: str
    state: str
    message: str


@dataclass(frozen=True)
class MessageStartEvent:
    task_id: str
    message_id: str


@dataclass(frozen=True)
class TokenEvent:
    task_id: str
    message_id: str
    delta: str


@dataclass(frozen=True)
class MessageEndEvent:
    task_id: str
    message_id: str
    finish_reason: str


class _EventRecorder:
    def __init__(self) -> None:
        self.events: list[StateEvent | MessageStartEvent | TokenEvent | MessageEndEvent] = []

    def on_message_start(self, task_id: str, message_id: str) -> None:
        self.events.append(MessageStartEvent(task_id=task_id, message_id=message_id))

    def on_token(self, task_id: str, message_id: str, delta: str) -> None:
        self.events.append(TokenEvent(task_id=task_id, message_id=message_id, delta=delta))

    def on_message_end(self, task_id: str, message_id: str, finish_reason: str) -> None:
        self.events.append(
            MessageEndEvent(task_id=task_id, message_id=message_id, finish_reason=finish_reason)
        )

    def on_state(self, task_id: str, state: str, message: str) -> None:
        self.events.append(StateEvent(task_id=task_id, state=state, message=message))


def _event_kinds(recorder: _EventRecorder) -> list[str]:
    return [event.__class__.__name__ for event in recorder.events]


def test_graph_event_sequence_for_normal_completion() -> None:
    recorder = _EventRecorder()
    workflow = build_agent_graph(checkpointer=InMemoryCheckpointer(), stream_emitter=recorder)

    result = workflow.invoke(
        {
            "task_id": "task-stream-1",
            "user_id": "user-1",
            "messages": ["run: echo hello"],
        }
    )

    assert result["task_complete"] is True

    assert _event_kinds(recorder) == [
        "StateEvent",
        "StateEvent",
        "StateEvent",
        "MessageStartEvent",
        "TokenEvent",
        "MessageEndEvent",
    ]

    thinking_event = recorder.events[0]
    assert isinstance(thinking_event, StateEvent)
    assert thinking_event.task_id == "task-stream-1"
    assert thinking_event.state == "thinking"

    running_event = recorder.events[1]
    assert isinstance(running_event, StateEvent)
    assert running_event.task_id == "task-stream-1"
    assert running_event.state == "running_command"

    done_event = recorder.events[2]
    assert isinstance(done_event, StateEvent)
    assert done_event.task_id == "task-stream-1"
    assert done_event.state == "done"

    done_start = recorder.events[3]
    done_token = recorder.events[4]
    done_end = recorder.events[5]
    assert isinstance(done_start, MessageStartEvent)
    assert isinstance(done_token, TokenEvent)
    assert isinstance(done_end, MessageEndEvent)
    assert done_start.task_id == "task-stream-1"
    assert done_token.task_id == "task-stream-1"
    assert done_end.task_id == "task-stream-1"
    assert done_token.message_id == done_start.message_id
    assert done_end.message_id == done_start.message_id
    assert done_end.finish_reason == "stop"
    assert done_token.delta.startswith("Task complete:")

    done_state_events = [
        event for event in recorder.events if isinstance(event, StateEvent) and event.state == "done"
    ]
    assert len(done_state_events) == 1


def test_graph_event_sequence_for_approval_required_pause() -> None:
    recorder = _EventRecorder()
    workflow = build_agent_graph(checkpointer=InMemoryCheckpointer(), stream_emitter=recorder)

    result = workflow.invoke(
        {
            "task_id": "task-stream-approval-1",
            "user_id": "user-1",
            "messages": ["run: rm -rf /tmp/demo"],
        }
    )

    assert result["awaiting_approval"] is True
    assert _event_kinds(recorder) == [
        "StateEvent",
        "StateEvent",
        "MessageStartEvent",
        "TokenEvent",
        "MessageEndEvent",
    ]

    awaiting_state_events = [
        event for event in recorder.events if isinstance(event, StateEvent) and event.state == "awaiting_approval"
    ]
    assert len(awaiting_state_events) == 1
    assert awaiting_state_events[0].task_id == "task-stream-approval-1"

    token_events = [event for event in recorder.events if isinstance(event, TokenEvent)]
    assert len(token_events) == 1
    assert token_events[0].delta == "Awaiting user approval for dangerous command."


def test_graph_event_sequence_for_error_path() -> None:
    recorder = _EventRecorder()

    def _raising_executor(_command: str) -> str:
        raise RuntimeError("forced command failure")

    workflow = build_agent_graph(
        checkpointer=InMemoryCheckpointer(),
        command_executor=_raising_executor,
        stream_emitter=recorder,
    )

    try:
        workflow.invoke(
            {
                "task_id": "task-stream-error-1",
                "user_id": "user-1",
                "messages": ["run: echo boom"],
            }
        )
        raise AssertionError("expected RuntimeError from command executor")
    except RuntimeError as exc:
        assert "forced command failure" in str(exc)

    assert _event_kinds(recorder) == [
        "StateEvent",
        "StateEvent",
        "StateEvent",
        "MessageStartEvent",
        "TokenEvent",
        "MessageEndEvent",
    ]

    error_state_events = [event for event in recorder.events if isinstance(event, StateEvent) and event.state == "error"]
    assert len(error_state_events) == 1
    assert error_state_events[0].task_id == "task-stream-error-1"
    assert "forced command failure" in error_state_events[0].message

    token_events = [event for event in recorder.events if isinstance(event, TokenEvent)]
    assert len(token_events) == 1
    assert "forced command failure" in token_events[0].delta

    message_start_events = [event for event in recorder.events if isinstance(event, MessageStartEvent)]
    message_end_events = [event for event in recorder.events if isinstance(event, MessageEndEvent)]
    assert len(message_start_events) == 1
    assert len(message_end_events) == 1
    assert message_end_events[0].message_id == message_start_events[0].message_id
    assert message_end_events[0].finish_reason == "error"
