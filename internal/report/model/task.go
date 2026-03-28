package model

const (
	TaskPending = "pending"
	TaskRunning = "running"
	TaskDone    = "done"
	TaskFailed  = "failed"
)

type AddReportTask struct {
	ID       string  `json:"id"`
	Status   string  `json:"status"`
	Progress float32 `json:"progress"` // 0.0 to 1.0
	Failed   []int   `json:"failed"`   // List of failed report indexes
}
