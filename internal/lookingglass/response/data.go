package response

type LookingGlassResponse struct {
	Id         int64  `json:"id"`
	ServerName string `json:"server_name"`
	TestURL    string `json:"test_url"`
}

type LookingGlassIDResponse struct {
	Id int64 `json:"id"`
}
