from unittest.mock import MagicMock

from context_compression import (
    build_candidate_context_messages,
    compress_context_messages,
)
from tool import TOOL_MESSAGE_TRUNCATED_NOTICE, truncate_tool_message_content


def test_truncate_tool_message_content_appends_notice_when_too_long():
    text = "x" * 5000
    result = truncate_tool_message_content(text)

    assert len(result) <= 4000
    assert result.endswith(TOOL_MESSAGE_TRUNCATED_NOTICE)


def test_compress_context_messages_returns_summary_and_tail():
    mock_client = MagicMock()
    mock_response = MagicMock()
    mock_choice = MagicMock()
    mock_choice.message.content = "summary text"
    mock_response.choices = [mock_choice]
    mock_client.chat.completions.create.return_value = mock_response

    context_messages = [
        {"role": "user", "content": "hello " * 5000},
        {"role": "assistant", "content": "world"},
        {"role": "user", "content": "tail-1"},
        {"role": "assistant", "content": "tail-2"},
        {"role": "user", "content": "tail-3"},
        {"role": "assistant", "content": "tail-4"},
        {"role": "user", "content": "tail-5"},
        {"role": "assistant", "content": "tail-6"},
    ]

    compressed = compress_context_messages(
        openai_client=mock_client,
        model="demo-model",
        context_messages=context_messages,
        tail_keep=6,
        compress_threshold_tokens=6000,
    )

    assert compressed is not None
    assert compressed[0]["role"] == "system"
    assert "summary text" in compressed[0]["content"]
    assert compressed[1:] == context_messages[-6:]


def test_build_candidate_context_messages_appends_latest_round():
    candidate = build_candidate_context_messages(
        request_messages=[{"role": "system", "content": "summary"}],
        request_message="next question",
        assistant_response="next answer",
    )

    assert candidate == [
        {"role": "system", "content": "summary"},
        {"role": "user", "content": "next question"},
        {"role": "assistant", "content": "next answer"},
    ]


def test_compress_context_messages_uses_usage_tokens_when_provided():
    mock_client = MagicMock()
    mock_response = MagicMock()
    mock_choice = MagicMock()
    mock_choice.message.content = "summary text"
    mock_response.choices = [mock_choice]
    mock_client.chat.completions.create.return_value = mock_response

    compressed = compress_context_messages(
        openai_client=mock_client,
        model="demo-model",
        context_messages=[
            {"role": "user", "content": "a"},
            {"role": "assistant", "content": "b"},
            {"role": "user", "content": "c"},
            {"role": "assistant", "content": "d"},
            {"role": "user", "content": "e"},
            {"role": "assistant", "content": "f"},
            {"role": "user", "content": "g"},
        ],
        tail_keep=2,
        compress_threshold_tokens=6000,
        usage_total_tokens=7000,
    )

    assert compressed is not None
    assert compressed[0]["role"] == "system"
