package model

type AddReportTask struct {
	Failed []int `json:"failed"` // List of failed report indexes
}
