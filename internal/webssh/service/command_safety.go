package service

import (
	"regexp"
	"strings"

	"VPSBenchmarkBackend/internal/webssh/model"
)

type CommandSafetyClassification struct {
	SafetyLevel      string
	RequiresApproval bool
}

type SafetyResult struct {
	Level            string `json:"risk_level"`
	Safe             bool   `json:"safe"`
	Reason           string `json:"reason"`
	RequiresApproval bool   `json:"requires_approval"`
}

var (
	killInitPattern      = regexp.MustCompile(`\bkill\s+-9\s+1\b`)
	forkBombPattern      = regexp.MustCompile(`:\s*\(\)\s*\{\s*:\s*\|\s*:\s*&\s*\}\s*;\s*:`)
	curlPipeShellPattern = regexp.MustCompile(`\bcurl\b.*\|\s*(bash|sh)\b`)

	rmRfPattern      = regexp.MustCompile(`\brm\s+-[a-z]*r[a-z]*f[a-z]*\b|\brm\s+-[a-z]*f[a-z]*r[a-z]*\b`)
	ddPattern        = regexp.MustCompile(`\bdd\b`)
	mkfsPattern      = regexp.MustCompile(`\bmkfs(\.[a-z0-9]+)?\b`)
	chmod777Pattern  = regexp.MustCompile(`\bchmod\s+(-r\s+)?777\b`)
	iptablesFPattern = regexp.MustCompile(`\biptables\s+-f\b`)

	aptInstallPattern       = regexp.MustCompile(`\bapt(-get)?\s+install\b`)
	pipInstallPattern       = regexp.MustCompile(`\bpip(3)?\s+install\b`)
	systemctlRestartPattern = regexp.MustCompile(`\bsystemctl\s+restart\b`)
	readOnlySafePattern     = regexp.MustCompile(`^(ls|df|free|uname)(\s+[-_./:=a-z0-9]+)*\s*$`)
	shellMetaPattern        = regexp.MustCompile("[;&|><`]|\\$\\(|\\n|\\r")
)

func ClassifyCommand(command string) SafetyResult {
	normalized := strings.ToLower(strings.TrimSpace(command))

	if isBlockedCommand(normalized) {
		return SafetyResult{
			Level:            model.SafetyLevelBlocked,
			Safe:             false,
			Reason:           "blocked command pattern detected",
			RequiresApproval: false,
		}
	}

	if isDangerousCommand(normalized) {
		return SafetyResult{
			Level:            model.SafetyLevelDangerous,
			Safe:             false,
			Reason:           "dangerous command pattern detected",
			RequiresApproval: true,
		}
	}

	if isWarningCommand(normalized) {
		return SafetyResult{
			Level:            model.SafetyLevelWarning,
			Safe:             false,
			Reason:           "warning command pattern detected",
			RequiresApproval: true,
		}
	}

	if isReadOnlyDiagnosticCommand(normalized) {
		return SafetyResult{
			Level:            model.SafetyLevelSafe,
			Safe:             true,
			Reason:           "safe read-only diagnostic command",
			RequiresApproval: false,
		}
	}

	return SafetyResult{
		Level:            model.SafetyLevelWarning,
		Safe:             false,
		Reason:           "warning by default for unknown command",
		RequiresApproval: true,
	}
}

func ClassifyCommandSafety(command string) CommandSafetyClassification {
	result := ClassifyCommand(command)
	return CommandSafetyClassification{
		SafetyLevel:      result.Level,
		RequiresApproval: result.RequiresApproval,
	}
}

func isBlockedCommand(command string) bool {
	return killInitPattern.MatchString(command) ||
		forkBombPattern.MatchString(command) ||
		curlPipeShellPattern.MatchString(command)
}

func isDangerousCommand(command string) bool {
	return rmRfPattern.MatchString(command) ||
		ddPattern.MatchString(command) ||
		mkfsPattern.MatchString(command) ||
		chmod777Pattern.MatchString(command) ||
		iptablesFPattern.MatchString(command)
}

func isWarningCommand(command string) bool {
	return aptInstallPattern.MatchString(command) ||
		pipInstallPattern.MatchString(command) ||
		systemctlRestartPattern.MatchString(command)
}

func isReadOnlyDiagnosticCommand(command string) bool {
	if shellMetaPattern.MatchString(command) {
		return false
	}

	return readOnlySafePattern.MatchString(command)
}
