package request

type LookingGlassRequest struct {
	ServerName string `json:"server_name" binding:"required"`
	TestURL    string `json:"test_url" binding:"required"`
}
