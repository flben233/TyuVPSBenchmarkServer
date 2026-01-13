package request

type AddReportRequest struct {
	HTML string `json:"html" binding:"required"`
}

type DeleteReportRequest struct {
	ID string `json:"id" binding:"required"`
}
