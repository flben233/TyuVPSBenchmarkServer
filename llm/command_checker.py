from __future__ import annotations

from dataclasses import dataclass, field

DEFAULT_ALLOWED_COMMANDS: frozenset[str] = frozenset({
    "ls", "cat", "head", "tail", "grep", "find", "wc", "sort", "uniq",
    "diff", "file", "stat", "readlink", "md5sum", "sha256sum",
    "ps", "df", "du", "free", "uptime", "uname", "whoami", "id",
    "hostname", "pwd", "date", "env", "printenv", "which", "whereis",
    "ip", "ifconfig", "netstat", "ss", "ping", "traceroute",
    "nslookup", "dig", "host", "curl", "wget",
    "systemctl", "journalctl",
    "lsblk", "lscpu", "vmstat", "iostat", "mpstat",
    "awk", "sed", "cut", "tr", "jq",
    "echo", "printf", "test", "true", "false",
    "man", "history", "less", "more", "tree",
})


@dataclass
class CommandCheckResult:
    all_allowed: bool
    disallowed: list[str] = field(default_factory=list)


class CommandChecker:
    """Whitelist-based command safety checker.

    The checker splits a compound shell command into individual sub-commands
    (respecting quoting) and verifies that every sub-command's *base command*
    (first token) appears in the merged whitelist.

    The merged whitelist is the union of:
      1. A built-in ``DEFAULT_ALLOWED_COMMANDS`` (idempotent / read-only commands).
      2. ``user_allowed`` – the per-user persistent whitelist stored in the Go DB.
      3. ``session_allowed`` – ephemeral additions for the current conversation.
    """

    def __init__(
        self,
        user_allowed: list[str] | None = None,
        session_allowed: list[str] | None = None,
    ) -> None:
        self._whitelist = DEFAULT_ALLOWED_COMMANDS | {
            c.strip().lower()
            for c in (user_allowed or [])
            if c.strip()
        } | {
            c.strip().lower()
            for c in (session_allowed or [])
            if c.strip()
        }

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------

    def check(self, command: str) -> CommandCheckResult:
        """Check *command* against the whitelist.

        Returns a ``CommandCheckResult`` whose ``all_allowed`` is ``True``
        when every sub-command's base is in the whitelist, and ``False``
        otherwise with ``disallowed`` listing the offending base commands.
        """
        if not command or not command.strip():
            return CommandCheckResult(all_allowed=True)

        sub_commands = self.split_commands(command)
        disallowed: list[str] = []
        seen: set[str] = set()

        for sub in sub_commands:
            base = self._extract_base_command(sub)
            if not base:
                continue
            if base not in self._whitelist and base not in seen:
                disallowed.append(base)
                seen.add(base)

        return CommandCheckResult(
            all_allowed=len(disallowed) == 0,
            disallowed=disallowed,
        )

    # ------------------------------------------------------------------
    # Command splitting – state machine
    # ------------------------------------------------------------------

    @staticmethod
    def split_commands(command: str) -> list[str]:
        """Split a compound shell command into individual sub-commands.

        Handles ``;``, ``&&``, ``||``, ``|``, and trailing ``&``.
        Respects single quotes, double quotes, and backticks so that
        separators inside quoted strings are ignored.
        """
        result: list[str] = []
        current: list[str] = []
        i = 0
        n = len(command)
        in_single = False
        in_double = False
        in_backtick = False

        while i < n:
            ch = command[i]

            # --- quote handling ---
            if ch == "'" and not in_double and not in_backtick:
                in_single = not in_single
                current.append(ch)
                i += 1
                continue
            if ch == '"' and not in_single and not in_backtick:
                in_double = not in_double
                current.append(ch)
                i += 1
                continue
            if ch == '`' and not in_single and not in_double:
                in_backtick = not in_backtick
                current.append(ch)
                i += 1
                continue

            if in_single or in_double or in_backtick:
                current.append(ch)
                i += 1
                continue

            # --- separator handling (outside quotes) ---
            if ch == ';':
                _flush(current, result)
                current = []
                i += 1
                continue

            if ch == '&':
                if i + 1 < n and command[i + 1] == '&':
                    _flush(current, result)
                    current = []
                    i += 2
                    continue
                # single & (background) – flush
                _flush(current, result)
                current = []
                i += 1
                continue

            if ch == '|':
                if i + 1 < n and command[i + 1] == '|':
                    _flush(current, result)
                    current = []
                    i += 2
                    continue
                # single | (pipe) – flush
                _flush(current, result)
                current = []
                i += 1
                continue

            # --- escape ---
            if ch == '\\' and i + 1 < n:
                current.append(ch)
                current.append(command[i + 1])
                i += 2
                continue

            current.append(ch)
            i += 1

        _flush(current, result)
        return result

    # ------------------------------------------------------------------
    # Base command extraction
    # ------------------------------------------------------------------

    @staticmethod
    def _extract_base_command(sub: str) -> str:
        """Extract the base command (first word) from a sub-command string.

        Handles leading environment variable assignments (``VAR=val cmd``)
        and the ``nice`` / ``nohup`` / ``sudo`` / ``timeout`` prefixes so
        that ``sudo rm -rf /`` yields ``rm``.
        """
        tokens = _tokenise(sub)
        # skip env assignments (KEY=VAL)
        for t in tokens:
            if '=' in t and not t.startswith('='):
                continue
            break
        else:
            return ""

        cmd = t
        # Unwrap sudo / nohup – these have a simple "prefix real_command" syntax
        # with no intervening arguments.  Other wrappers like nice/timeout/strace
        # have their own flags between the prefix and the real command, so trying
        # to strip them would be fragile and is skipped on purpose.
        prefixes = {"sudo", "nohup"}
        while cmd in prefixes and tokens:
            tokens = tokens[1:]
            # skip more env assignments after prefix
            while tokens:
                t2 = tokens[0]
                if '=' in t2 and not t2.startswith('='):
                    tokens = tokens[1:]
                    continue
                break
            if not tokens:
                return ""
            cmd = tokens[0]

        # strip path prefix: /usr/bin/rm -> rm
        if '/' in cmd:
            cmd = cmd.rsplit('/', 1)[-1]

        return cmd.lower()


# ------------------------------------------------------------------
# Helpers
# ------------------------------------------------------------------

def _flush(current: list[str], result: list[str]) -> None:
    s = ''.join(current).strip()
    if s:
        result.append(s)


def _tokenise(sub: str) -> list[str]:
    """Simple shell-style tokeniser that respects quoting."""
    tokens: list[str] = []
    current: list[str] = []
    i = 0
    n = len(sub)
    while i < n:
        ch = sub[i]
        if ch in (' ', '\t'):
            if current:
                tokens.append(''.join(current))
                current = []
            i += 1
            continue
        if ch in ("'", '"', '`'):
            quote = ch
            current.append(ch)
            i += 1
            while i < n and sub[i] != quote:
                if sub[i] == '\\' and i + 1 < n:
                    current.append(sub[i])
                    current.append(sub[i + 1])
                    i += 2
                else:
                    current.append(sub[i])
                    i += 1
            if i < n:
                current.append(sub[i])
                i += 1
            continue
        if ch == '\\' and i + 1 < n:
            current.append(sub[i])
            current.append(sub[i + 1])
            i += 2
            continue
        current.append(ch)
        i += 1
    if current:
        tokens.append(''.join(current))
    return tokens
