package response

type LookingGlassResponse struct {
	Id           int64  `json:"id"`
	ServerName   string `json:"server_name"`
	TestURL      string `json:"test_url"`
	UploaderName string `json:"uploader_name,omitempty"`
	ReviewStatus int    `json:"review_status,omitempty"`
}

type LookingGlassIDResponse struct {
	Id int64 `json:"id"`
}
