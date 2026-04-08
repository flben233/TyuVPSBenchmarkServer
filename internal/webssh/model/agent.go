package model

const (
	AgentStatusRunning          = "running"
	AgentStatusAwaitingApproval = "awaiting_approval"
	AgentStatusCompleted        = "completed"
	AgentStatusFailed           = "failed"
)

const (
	SafetyLevelSafe      = "safe"
	SafetyLevelWarning   = "warning"
	SafetyLevelDangerous = "dangerous"
	SafetyLevelBlocked   = "blocked"
)
