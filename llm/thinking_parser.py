from __future__ import annotations


class ThoughtParser:
    """Stateful streaming parser that splits <thought>...</thought> tags in content.

    Some OpenAI-compatible APIs (e.g. certain GLM / Qwen deployments) embed the
    model's reasoning inside ``<thought>...</thought>`` within the regular
    ``content`` field instead of using a dedicated ``reasoning_content`` field.
    Because streaming deltas can split a tag across chunks (e.g. ``"<thou"`` in
    one delta and ``"ght>"`` in the next), a naive ``str.replace`` approach is
    insufficient.

    This parser maintains two pieces of state:
      * ``_in_thought`` – whether we are currently inside a ``<thought>`` block.
      * ``_buf`` – a small buffer for the tail of the current chunk that *might*
        be the beginning of an opening or closing tag.  On the next ``feed()``
        call the buffer is prepended to the new text so the tag can be matched
        in full.

    Returns ``(content_text, thinking_text)`` from both ``feed()`` and
    ``flush()``, where *content_text* is everything outside the tags and
    *thinking_text* is everything inside.
    """

    _OPEN = "<thought>"
    _CLOSE = "</thought>"
    _MAX_TAG = max(len(_OPEN), len(_CLOSE))

    def __init__(self) -> None:
        self._in_thought = False
        self._buf = ""

    def feed(self, text: str) -> tuple[str, str]:
        """Process one streaming delta and return ``(content, thinking)``.

        Prepend the leftover buffer from the previous call, then scan for tag
        boundaries.  Any text that cannot yet be classified (because it might be
        the start of a tag) is stored back into ``_buf`` for the next call.
        """
        if not text:
            return "", ""
        # Prepend leftover from the previous delta that may be a partial tag
        text = self._buf + text
        self._buf = ""
        content_parts: list[str] = []
        thinking_parts: list[str] = []
        i = 0
        while i < len(text):
            if self._in_thought:
                # Look for the closing tag
                idx = text.find(self._CLOSE, i)
                if idx == -1:
                    thinking_parts.append(text[i:])
                    break
                # Found closing tag – emit the thinking text up to it
                thinking_parts.append(text[i:idx])
                self._in_thought = False
                i = idx + len(self._CLOSE)
            else:
                # Look for the opening tag
                idx = text.find(self._OPEN, i)
                if idx == -1:
                    content_parts.append(text[i:])
                    break
                # Found opening tag – emit the content text up to it
                content_parts.append(text[i:idx])
                self._in_thought = True
                i = idx + len(self._OPEN)
        return "".join(content_parts), "".join(thinking_parts)

    def flush(self) -> tuple[str, str]:
        """Drain the buffer after the stream ends.

        Any remaining buffered text is classified according to the current
        state: still inside ``<thought>`` → thinking; otherwise → content.
        """
        remaining = self._buf
        self._buf = ""
        if not remaining:
            return "", ""
        if self._in_thought:
            return "", remaining
        return remaining, ""


def split_thinking_content(text: str) -> tuple[str, str]:
    """Split complete text into visible content and hidden thinking text."""
    parser = ThoughtParser()
    content_text, thinking_text = parser.feed(text)
    flushed_content, flushed_thinking = parser.flush()
    return content_text + flushed_content, thinking_text + flushed_thinking
