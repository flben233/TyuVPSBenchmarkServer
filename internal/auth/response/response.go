package response

// LoginResponse is the response structure for login
type LoginResponse struct {
	Token string `json:"token"`
}

// UserInfo contains the user information from JWT
type UserInfo struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}
