from __future__ import annotations


class ThoughtParser:
    """Stateful streaming parser that splits thinking tags in content.

    Some OpenAI-compatible APIs embed the model's reasoning inside tags within
    the regular ``content`` field instead of using a dedicated
    ``reasoning_content`` field.  Two common tag formats are supported:

    * ``<thought>...</thought>`` (e.g. certain GLM deployments)
    * ``<think>...</think>`` (e.g. Qwen3 series)

    Because streaming deltas can split a tag across chunks (e.g. ``"<thou"`` in
    one delta and ``"ght>"`` in the next), a naive ``str.replace`` approach is
    insufficient.

    This parser maintains two pieces of state:
      * ``_in_thought`` – whether we are currently inside a thinking block.
      * ``_active_close`` – the closing tag that matches the current opening tag.
      * ``_buf`` – a small buffer for the tail of the current chunk that *might*
        be the beginning of an opening or closing tag.  On the next ``feed()``
        call the buffer is prepended to the new text so the tag can be matched
        in full.

    Returns ``(content_text, thinking_text)`` from both ``feed()`` and
    ``flush()``, where *content_text* is everything outside the tags and
    *thinking_text* is everything inside.
    """

    _OPEN_TAGS = ("<thought>", "<think>")
    _CLOSE_MAP = {
        "<thought>": "</thought>",
        "<think>": "</think>",
    }
    _ALL_TAGS = _OPEN_TAGS + tuple(_CLOSE_MAP.values())
    _MAX_TAG = max(len(t) for t in _ALL_TAGS)

    def __init__(self) -> None:
        self._in_thought = False
        self._active_close: str = ""
        self._buf = ""

    def feed(self, text: str) -> tuple[str, str]:
        """Process one streaming delta and return ``(content, thinking)``.

        Prepend the leftover buffer from the previous call, then scan for tag
        boundaries.  Any text that cannot yet be classified (because it might be
        the start of a tag) is stored back into ``_buf`` for the next call.
        """
        if not text:
            return "", ""
        text = self._buf + text
        self._buf = ""
        content_parts: list[str] = []
        thinking_parts: list[str] = []
        i = 0
        while i < len(text):
            if self._in_thought:
                close_tag = self._active_close
                idx = text.find(close_tag, i)
                if idx != -1:
                    thinking_parts.append(text[i:idx])
                    self._in_thought = False
                    self._active_close = ""
                    i = idx + len(close_tag)
                else:
                    # No complete closing tag found.  Check if the tail might
                    # contain a partial closing tag (e.g. "</th" at the end).
                    last_lt = text.rfind("<", i)
                    if last_lt >= i:
                        thinking_parts.append(text[i:last_lt])
                        self._buf = text[last_lt:]
                    else:
                        thinking_parts.append(text[i:])
                    break
            else:
                # Find the earliest opening tag
                earliest_idx = -1
                matched_open = ""
                for open_tag in self._OPEN_TAGS:
                    idx = i
                    while True:
                        idx = text.find(open_tag, idx)
                        if idx == -1:
                            break
                        # Skip if this is actually part of a closing tag
                        # (e.g. "<think>" inside "</think>")
                        after = idx + len(open_tag)
                        if after < len(text) and text[after] == "/":
                            idx = after
                            continue
                        if earliest_idx == -1 or idx < earliest_idx:
                            earliest_idx = idx
                            matched_open = open_tag
                        break
                if earliest_idx != -1:
                    content_parts.append(text[i:earliest_idx])
                    self._in_thought = True
                    self._active_close = self._CLOSE_MAP[matched_open]
                    i = earliest_idx + len(matched_open)
                else:
                    # No complete opening tag found.  Check if the tail might
                    # contain a partial opening tag (e.g. "<thou" at the end).
                    last_lt = text.rfind("<", i)
                    if last_lt >= i:
                        content_parts.append(text[i:last_lt])
                        self._buf = text[last_lt:]
                    else:
                        content_parts.append(text[i:])
                    break
        return "".join(content_parts), "".join(thinking_parts)

    def flush(self) -> tuple[str, str]:
        """Drain the buffer after the stream ends.

        Any remaining buffered text is classified according to the current
        state: still inside a thinking block → thinking; otherwise → content.
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
