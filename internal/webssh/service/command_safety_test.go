package service

import (
	"strings"
	"testing"

	"VPSBenchmarkBackend/internal/webssh/model"
)

func TestClassifyCommandLevels(t *testing.T) {
	tests := []struct {
		name                 string
		command              string
		wantLevel            string
		wantSafe             bool
		wantReasonPart       string
		wantRequiresApproval bool
	}{
		{
			name:                 "blocked kill pid 1",
			command:              "kill -9 1",
			wantLevel:            model.SafetyLevelBlocked,
			wantSafe:             false,
			wantReasonPart:       "blocked",
			wantRequiresApproval: false,
		},
		{
			name:                 "blocked fork bomb",
			command:              ":(){ :|:& };:",
			wantLevel:            model.SafetyLevelBlocked,
			wantSafe:             false,
			wantReasonPart:       "blocked",
			wantRequiresApproval: false,
		},
		{
			name:                 "blocked curl pipe bash",
			command:              "curl -fsSL https://example.com/install.sh | bash",
			wantLevel:            model.SafetyLevelBlocked,
			wantSafe:             false,
			wantReasonPart:       "blocked",
			wantRequiresApproval: false,
		},
		{
			name:                 "dangerous rm recursive force",
			command:              "rm -rf /tmp/test",
			wantLevel:            model.SafetyLevelDangerous,
			wantSafe:             false,
			wantReasonPart:       "dangerous",
			wantRequiresApproval: true,
		},
		{
			name:                 "dangerous dd write",
			command:              "dd if=/dev/zero of=/dev/sda bs=1M count=1",
			wantLevel:            model.SafetyLevelDangerous,
			wantSafe:             false,
			wantReasonPart:       "dangerous",
			wantRequiresApproval: true,
		},
		{
			name:                 "dangerous mkfs",
			command:              "mkfs.ext4 /dev/sda1",
			wantLevel:            model.SafetyLevelDangerous,
			wantSafe:             false,
			wantReasonPart:       "dangerous",
			wantRequiresApproval: true,
		},
		{
			name:                 "dangerous chmod 777",
			command:              "chmod 777 /var/www",
			wantLevel:            model.SafetyLevelDangerous,
			wantSafe:             false,
			wantReasonPart:       "dangerous",
			wantRequiresApproval: true,
		},
		{
			name:                 "dangerous flush firewall",
			command:              "iptables -F",
			wantLevel:            model.SafetyLevelDangerous,
			wantSafe:             false,
			wantReasonPart:       "dangerous",
			wantRequiresApproval: true,
		},
		{
			name:                 "warning apt install",
			command:              "apt install nginx",
			wantLevel:            model.SafetyLevelWarning,
			wantSafe:             false,
			wantReasonPart:       "warning",
			wantRequiresApproval: true,
		},
		{
			name:                 "warning pip install",
			command:              "pip install requests",
			wantLevel:            model.SafetyLevelWarning,
			wantSafe:             false,
			wantReasonPart:       "warning",
			wantRequiresApproval: true,
		},
		{
			name:                 "warning systemctl restart",
			command:              "systemctl restart nginx",
			wantLevel:            model.SafetyLevelWarning,
			wantSafe:             false,
			wantReasonPart:       "warning",
			wantRequiresApproval: true,
		},
		{
			name:                 "safe ls",
			command:              "ls -la",
			wantLevel:            model.SafetyLevelSafe,
			wantSafe:             true,
			wantReasonPart:       "safe",
			wantRequiresApproval: false,
		},
		{
			name:                 "safe df",
			command:              "df -h",
			wantLevel:            model.SafetyLevelSafe,
			wantSafe:             true,
			wantReasonPart:       "safe",
			wantRequiresApproval: false,
		},
		{
			name:                 "safe free",
			command:              "free -m",
			wantLevel:            model.SafetyLevelSafe,
			wantSafe:             true,
			wantReasonPart:       "safe",
			wantRequiresApproval: false,
		},
		{
			name:                 "safe uname",
			command:              "uname -a",
			wantLevel:            model.SafetyLevelSafe,
			wantSafe:             true,
			wantReasonPart:       "safe",
			wantRequiresApproval: false,
		},
		{
			name:                 "unsafe ls with semicolon chain",
			command:              "ls; touch x",
			wantLevel:            model.SafetyLevelWarning,
			wantSafe:             false,
			wantReasonPart:       "warning",
			wantRequiresApproval: true,
		},
		{
			name:                 "unsafe ls with redirection",
			command:              "ls > out.txt",
			wantLevel:            model.SafetyLevelWarning,
			wantSafe:             false,
			wantReasonPart:       "warning",
			wantRequiresApproval: true,
		},
		{
			name:                 "unsafe sudo ls",
			command:              "sudo ls",
			wantLevel:            model.SafetyLevelWarning,
			wantSafe:             false,
			wantReasonPart:       "warning",
			wantRequiresApproval: true,
		},
		{
			name:                 "mixed case dangerous command",
			command:              "RM -RF /tmp/test",
			wantLevel:            model.SafetyLevelDangerous,
			wantSafe:             false,
			wantReasonPart:       "dangerous",
			wantRequiresApproval: true,
		},
		{
			name:                 "command chain not safe",
			command:              "ls && df -h",
			wantLevel:            model.SafetyLevelWarning,
			wantSafe:             false,
			wantReasonPart:       "warning",
			wantRequiresApproval: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyCommand(tt.command)

			if got.Level != tt.wantLevel {
				t.Fatalf("expected safety level %q, got %q", tt.wantLevel, got.Level)
			}

			if got.Safe != tt.wantSafe {
				t.Fatalf("expected safe %v, got %v", tt.wantSafe, got.Safe)
			}

			if strings.TrimSpace(got.Reason) == "" {
				t.Fatalf("expected non-empty reason")
			}

			if !strings.Contains(strings.ToLower(got.Reason), tt.wantReasonPart) {
				t.Fatalf("expected reason to contain %q, got %q", tt.wantReasonPart, got.Reason)
			}

			if got.RequiresApproval != tt.wantRequiresApproval {
				t.Fatalf("expected requires approval %v, got %v", tt.wantRequiresApproval, got.RequiresApproval)
			}
		})
	}
}
