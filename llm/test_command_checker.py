import pytest

from command_checker import CommandChecker, DEFAULT_ALLOWED_COMMANDS


# ===================================================================
# split_commands
# ===================================================================

class TestSplitCommands:

    def test_single_command(self):
        assert CommandChecker.split_commands("ls -la") == ["ls -la"]

    def test_semicolon(self):
        assert CommandChecker.split_commands("ls;pwd") == ["ls", "pwd"]

    def test_semicolon_with_spaces(self):
        assert CommandChecker.split_commands("ls -la ; pwd") == ["ls -la", "pwd"]

    def test_double_amp(self):
        assert CommandChecker.split_commands("mkdir foo && cd foo") == ["mkdir foo", "cd foo"]

    def test_double_pipe(self):
        assert CommandChecker.split_commands("test -d foo || echo missing") == ["test -d foo", "echo missing"]

    def test_pipe(self):
        assert CommandChecker.split_commands("cat file | grep x") == ["cat file", "grep x"]

    def test_single_amp_background(self):
        assert CommandChecker.split_commands("sleep 1 &") == ["sleep 1"]

    def test_mixed_separators(self):
        result = CommandChecker.split_commands("ls;pwd&&cat file||echo no|grep x")
        assert result == ["ls", "pwd", "cat file", "echo no", "grep x"]

    def test_empty_string(self):
        assert CommandChecker.split_commands("") == []

    def test_only_separators(self):
        assert CommandChecker.split_commands(";;;") == []

    def test_only_spaces(self):
        assert CommandChecker.split_commands("   ") == []

    def test_trailing_semicolon(self):
        assert CommandChecker.split_commands("ls;") == ["ls"]

    def test_leading_semicolon(self):
        assert CommandChecker.split_commands(";ls") == ["ls"]

    def test_consecutive_separators(self):
        assert CommandChecker.split_commands("ls;;pwd") == ["ls", "pwd"]

    def test_single_quotes_ignore_separators(self):
        assert CommandChecker.split_commands("echo 'a;b|c&&d||e&f'") == ["echo 'a;b|c&&d||e&f'"]

    def test_double_quotes_ignore_separators(self):
        assert CommandChecker.split_commands('echo "a;b|c&&d||e"') == ['echo "a;b|c&&d||e"']

    def test_backticks_ignore_separators(self):
        assert CommandChecker.split_commands("echo `a;b`") == ["echo `a;b`"]

    def test_mixed_quotes_semicolon_outside(self):
        result = CommandChecker.split_commands("echo 'hello';ls")
        assert result == ["echo 'hello'", "ls"]

    def test_double_quotes_inside_single(self):
        result = CommandChecker.split_commands("echo 'he said \"hello\"; bye'")
        assert result == ["echo 'he said \"hello\"; bye'"]

    def test_single_quotes_inside_double(self):
        result = CommandChecker.split_commands("echo \"it's; fine\";ls")
        assert result == ["echo \"it's; fine\"", "ls"]

    def test_escaped_semicolon(self):
        result = CommandChecker.split_commands("echo a\\;b")
        assert result == ["echo a\\;b"]

    def test_escaped_pipe(self):
        result = CommandChecker.split_commands("echo a\\|b")
        assert result == ["echo a\\|b"]

    def test_complex_pipeline(self):
        result = CommandChecker.split_commands("cat /var/log/syslog | grep error | sort | uniq -c | sort -rn | head -20")
        assert len(result) == 6
        assert result[0] == "cat /var/log/syslog"
        assert result[-1] == "head -20"

    def test_ampersand_in_middle(self):
        result = CommandChecker.split_commands("ls & pwd")
        assert result == ["ls", "pwd"]

    def test_chained_with_all_ops(self):
        result = CommandChecker.split_commands("ls && cat file || echo no; pwd | grep x &")
        assert result == ["ls", "cat file", "echo no", "pwd", "grep x"]


# ===================================================================
# _extract_base_command
# ===================================================================

class TestExtractBaseCommand:

    def test_simple(self):
        assert CommandChecker._extract_base_command("ls -la") == "ls"

    def test_no_args(self):
        assert CommandChecker._extract_base_command("pwd") == "pwd"

    def test_path_prefix(self):
        assert CommandChecker._extract_base_command("/usr/bin/rm -rf /") == "rm"

    def test_sudo_prefix(self):
        assert CommandChecker._extract_base_command("sudo rm -rf /") == "rm"

    def test_sudo_with_path(self):
        assert CommandChecker._extract_base_command("sudo /sbin/reboot") == "reboot"

    def test_double_prefix(self):
        assert CommandChecker._extract_base_command("sudo nohup cat file") == "cat"

    def test_triple_prefix(self):
        # timeout has its own args between it and the real command,
        # so only sudo/nohup are stripped; timeout itself becomes the base
        assert CommandChecker._extract_base_command("sudo timeout 10 rm -rf /") == "timeout"

    def test_env_assignment(self):
        assert CommandChecker._extract_base_command("FOO=bar ls -la") == "ls"

    def test_multiple_env_assignments(self):
        assert CommandChecker._extract_base_command("A=1 B=2 ls") == "ls"

    def test_env_and_sudo(self):
        assert CommandChecker._extract_base_command("FOO=bar sudo rm -rf /") == "rm"

    def test_empty_string(self):
        assert CommandChecker._extract_base_command("") == ""

    def test_only_spaces(self):
        assert CommandChecker._extract_base_command("   ") == ""

    def test_only_env_assignment(self):
        assert CommandChecker._extract_base_command("FOO=bar") == ""

    def test_case_insensitive(self):
        assert CommandChecker._extract_base_command("LS -la") == "ls"

    def test_strace_prefix(self):
        # strace has its own args (-f), so it is NOT stripped
        assert CommandChecker._extract_base_command("strace -f cat file") == "strace"

    def test_nice_prefix(self):
        # nice has its own args (-n 19), so it is NOT stripped
        assert CommandChecker._extract_base_command("nice -n 19 rm -rf /tmp") == "nice"

    def test_quoted_command(self):
        assert CommandChecker._extract_base_command("'ls' -la") == "'ls'"

    def test_timeout_prefix(self):
        # timeout has its own arg (5), so it is NOT stripped
        assert CommandChecker._extract_base_command("timeout 5 curl https://example.com") == "timeout"


# ===================================================================
# check (integration)
# ===================================================================

class TestCheck:

    def test_allowed_default_ls(self):
        r = CommandChecker().check("ls -la /home")
        assert r.all_allowed is True
        assert r.disallowed == []

    def test_allowed_default_grep_pipe(self):
        r = CommandChecker().check("cat file | grep error")
        assert r.all_allowed is True

    def test_disallowed_rm(self):
        r = CommandChecker().check("rm -rf /tmp/old")
        assert r.all_allowed is False
        assert "rm" in r.disallowed

    def test_disallowed_mixed(self):
        r = CommandChecker().check("ls; rm -rf /")
        assert r.all_allowed is False
        assert "rm" in r.disallowed
        assert "ls" not in r.disallowed

    def test_mixed_pipeline_allowed_and_disallowed(self):
        r = CommandChecker().check("ls | rm -rf /")
        assert r.all_allowed is False
        assert "rm" in r.disallowed

    def test_multiple_disallowed_deduplicated(self):
        r = CommandChecker().check("rm -rf /a; rm -rf /b")
        assert r.disallowed.count("rm") == 1

    def test_two_different_disallowed(self):
        r = CommandChecker().check("rm -rf /; mkfs /dev/sda1")
        assert r.all_allowed is False
        assert set(r.disallowed) == {"rm", "mkfs"}

    def test_empty_command(self):
        r = CommandChecker().check("")
        assert r.all_allowed is True

    def test_whitespace_only(self):
        r = CommandChecker().check("   ")
        assert r.all_allowed is True

    def test_user_allowed_extends(self):
        r = CommandChecker(user_allowed=["rm"]).check("rm -rf /tmp")
        assert r.all_allowed is True

    def test_session_allowed_extends(self):
        r = CommandChecker(session_allowed=["rm"]).check("rm -rf /tmp")
        assert r.all_allowed is True

    def test_user_and_session_combined(self):
        r = CommandChecker(
            user_allowed=["rm"],
            session_allowed=["shutdown"],
        ).check("rm -rf /; shutdown now")
        assert r.all_allowed is True

    def test_user_allowed_case_insensitive(self):
        r = CommandChecker(user_allowed=["RM"]).check("rm -rf /")
        assert r.all_allowed is True

    def test_user_allowed_stripped(self):
        r = CommandChecker(user_allowed=[" rm ", "  "]).check("rm -rf /")
        assert r.all_allowed is True

    def test_sudo_disallowed(self):
        r = CommandChecker().check("sudo rm -rf /")
        assert r.all_allowed is False
        assert "rm" in r.disallowed

    def test_sudo_allowed_command(self):
        r = CommandChecker().check("sudo ls /root")
        assert r.all_allowed is True

    def test_path_prefix_disallowed(self):
        r = CommandChecker().check("/usr/bin/rm -rf /")
        assert r.all_allowed is False
        assert "rm" in r.disallowed

    def test_path_prefix_allowed(self):
        r = CommandChecker().check("/usr/bin/ls")
        assert r.all_allowed is True

    def test_env_prefix_disallowed(self):
        r = CommandChecker().check("FOO=bar rm -rf /")
        assert r.all_allowed is False
        assert "rm" in r.disallowed

    def test_env_prefix_allowed(self):
        r = CommandChecker().check("LANG=C sort file")
        assert r.all_allowed is True

    def test_separators_in_quotes_ignored(self):
        r = CommandChecker().check("echo 'rm -rf /'")
        assert r.all_allowed is True

    def test_pipe_each_segment_checked(self):
        r = CommandChecker().check("ls | rm -rf / | cat file")
        assert r.all_allowed is False
        assert "rm" in r.disallowed

    def test_double_pipe_each_segment_checked(self):
        r = CommandChecker().check("ls || rm -rf /")
        assert r.all_allowed is False

    def test_double_amp_each_segment_checked(self):
        r = CommandChecker().check("ls && rm -rf /")
        assert r.all_allowed is False

    def test_background_checked(self):
        r = CommandChecker().check("rm -rf / &")
        assert r.all_allowed is False

    def test_all_default_commands_individually_allowed(self):
        for cmd in DEFAULT_ALLOWED_COMMANDS:
            r = CommandChecker().check(cmd)
            assert r.all_allowed is True, f"default command '{cmd}' should be allowed"

    def test_complex_allowed_chain(self):
        r = CommandChecker().check(
            "find /var/log -name '*.log' | grep -l error | sort | uniq -c | sort -rn | head -10"
        )
        assert r.all_allowed is True

    def test_complex_disallowed_chain(self):
        r = CommandChecker().check(
            "find /tmp -name '*.log' | xargs rm -rf; shutdown now"
        )
        assert r.all_allowed is False
        assert "xargs" in r.disallowed
        assert "shutdown" in r.disallowed

    def test_default_whitelist_not_mutable(self):
        original_size = len(DEFAULT_ALLOWED_COMMANDS)
        CommandChecker(user_allowed=["rm"])
        assert len(DEFAULT_ALLOWED_COMMANDS) == original_size

    def test_reject_substring_not_prefix(self):
        r = CommandChecker().check("lsof -i :80")
        assert r.all_allowed is False
        assert "lsof" in r.disallowed

    def test_reject_similar_name(self):
        r = CommandChecker().check("lsblk_something")
        assert r.all_allowed is False

    def test_docker_not_in_default(self):
        r = CommandChecker().check("docker run -d nginx")
        assert r.all_allowed is False
        assert "docker" in r.disallowed

    def test_reboot_not_in_default(self):
        r = CommandChecker().check("reboot")
        assert r.all_allowed is False

    def test_mkdir_not_in_default(self):
        r = CommandChecker().check("mkdir /tmp/test")
        assert r.all_allowed is False

    def test_cp_not_in_default(self):
        r = CommandChecker().check("cp file1 file2")
        assert r.all_allowed is False

    def test_mv_not_in_default(self):
        r = CommandChecker().check("mv file1 file2")
        assert r.all_allowed is False

    def test_chmod_not_in_default(self):
        r = CommandChecker().check("chmod 755 /tmp")
        assert r.all_allowed is False

    def test_systemctl_subcommand_allowed(self):
        r = CommandChecker().check("systemctl status nginx")
        assert r.all_allowed is True

    def test_awk_allowed(self):
        r = CommandChecker().check("awk '{print $1}' file")
        assert r.all_allowed is True

    def test_curl_allowed(self):
        r = CommandChecker().check("curl -s https://example.com")
        assert r.all_allowed is True

    def test_wget_allowed(self):
        r = CommandChecker().check("wget https://example.com/file.tar.gz")
        assert r.all_allowed is True

    def test_dd_not_in_default(self):
        r = CommandChecker().check("dd if=/dev/zero of=/dev/sda")
        assert r.all_allowed is False

    def test_nohup_disallowed_inner(self):
        r = CommandChecker().check("nohup rm -rf /")
        assert r.all_allowed is False
        assert "rm" in r.disallowed

    def test_nohup_allowed_inner(self):
        r = CommandChecker().check("nohup cat file &")
        assert r.all_allowed is True

    def test_empty_user_and_session_lists(self):
        r = CommandChecker(user_allowed=[], session_allowed=[]).check("ls")
        assert r.all_allowed is True

    def test_none_user_and_session(self):
        r = CommandChecker(user_allowed=None, session_allowed=None).check("rm")
        assert r.all_allowed is False

    def test_xargs_not_in_default(self):
        r = CommandChecker().check("xargs rm -rf")
        assert r.all_allowed is False
        assert "xargs" in r.disallowed

    def test_tee_not_in_default(self):
        r = CommandChecker().check("echo hi | tee /tmp/file")
        assert r.all_allowed is False
        assert "tee" in r.disallowed

    def test_nice_wrapping_disallowed(self):
        r = CommandChecker().check("nice -n 19 rm -rf /")
        assert r.all_allowed is False
        assert "nice" in r.disallowed

    def test_timeout_wrapping_disallowed(self):
        r = CommandChecker().check("timeout 10 rm -rf /")
        assert r.all_allowed is False
        assert "timeout" in r.disallowed
