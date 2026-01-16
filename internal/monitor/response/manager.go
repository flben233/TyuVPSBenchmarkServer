package response

type HostResponse struct {
	Target       string `json:"target"`
	Name         string `json:"name"`
	Id           int64  `json:"id"`
	UploaderName string `json:"uploader_name,omitempty"`
	ReviewStatus int    `json:"review_status,omitempty"`
}

type HostIDResponse struct {
	Id int64 `json:"id"`
}
