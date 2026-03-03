package request

type AddReportRequest struct {
	HTML      string `json:"html" binding:"required"`
	MonitorID *int64 `json:"monitor_id"`
	OtherInfo string `json:"other_info"`
}

type DeleteReportRequest struct {
	ID string `json:"id" binding:"required"`
}

type UpdateReportRequest struct {
	ID        string `json:"id" binding:"required"`
	MonitorID int64  `json:"monitor_id" binding:"required"`
	OtherInfo string `json:"other_info"`
}
