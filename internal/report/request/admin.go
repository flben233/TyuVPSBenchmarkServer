package request

type AddReportRequest struct {
	HTML      string `json:"html" binding:"required"`
	MonitorID *int64 `json:"monitor_id"`
}

type DeleteReportRequest struct {
	ID string `json:"id" binding:"required"`
}
