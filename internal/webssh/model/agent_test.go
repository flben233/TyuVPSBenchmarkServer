package model

import "testing"

func TestAgentStatusAndSafetyLevelConstants(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{name: "AgentStatusRunning", got: AgentStatusRunning, want: "running"},
		{name: "AgentStatusAwaitingApproval", got: AgentStatusAwaitingApproval, want: "awaiting_approval"},
		{name: "AgentStatusCompleted", got: AgentStatusCompleted, want: "completed"},
		{name: "AgentStatusFailed", got: AgentStatusFailed, want: "failed"},
		{name: "SafetyLevelSafe", got: SafetyLevelSafe, want: "safe"},
		{name: "SafetyLevelWarning", got: SafetyLevelWarning, want: "warning"},
		{name: "SafetyLevelDangerous", got: SafetyLevelDangerous, want: "dangerous"},
		{name: "SafetyLevelBlocked", got: SafetyLevelBlocked, want: "blocked"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("expected %s to be %q, got %q", tc.name, tc.want, tc.got)
			}
		})
	}
}
