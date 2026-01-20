package response

// ReportInfoResponse represents report info in list responses
type ReportInfoResponse struct {
	Name string `json:"name"`
	Id   string `json:"id"`
	Date string `json:"date"`
}

type ReportIDResponse struct {
	Id string `json:"id"`
}
