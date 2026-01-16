package service

import (
	"VPSBenchmarkBackend/internal/config"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GithubTokenResponse represents the response from GitHub OAuth token endpoint
type GithubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

// GithubUserInfo represents the user info from GitHub API
type GithubUserInfo struct {
	Login     string `json:"login"`
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

// GithubLogin exchanges a GitHub OAuth code for a JWT token
func GithubLogin(code string) (string, error) {
	cfg := config.Get()

	// Exchange code for access token
	accessToken, err := exchangeCodeForToken(code, cfg)
	if err != nil {
		return "", err
	}

	// Get user info from GitHub
	userInfo, err := getGithubUserInfo(accessToken)
	if err != nil {
		return "", err
	}

	// Generate JWT token
	return generateJWTToken(userInfo, cfg)
}

// exchangeCodeForToken exchanges the OAuth code for an access token
func exchangeCodeForToken(code string, cfg *config.Config) (string, error) {
	params := url.Values{}
	params.Add("client_id", cfg.ClientID)
	params.Add("client_secret", cfg.ClientSecret)
	params.Add("code", code)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", nil)
	if err != nil {
		return "", err
	}
	req.URL.RawQuery = params.Encode()
	req.Header.Set("Accept", "application/json")

	client := getHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp GithubTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	if tokenResp.Error != "" {
		return "", errors.New(tokenResp.ErrorDesc)
	}

	if tokenResp.AccessToken == "" {
		return "", errors.New("failed to get access token from GitHub")
	}

	return tokenResp.AccessToken, nil
}

// getGithubUserInfo fetches the user info from GitHub API
func getGithubUserInfo(accessToken string) (*GithubUserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := getHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get user info from GitHub")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GithubUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func getHttpClient() *http.Client {
	proxyURL, _ := url.Parse(config.Get().GithubHttpProxy)
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
}

// generateJWTToken generates a JWT token for the user
func generateJWTToken(userInfo *GithubUserInfo, cfg *config.Config) (string, error) {
	// Use login name if display name is empty
	name := userInfo.Name
	if name == "" {
		name = userInfo.Login
	}

	claims := jwt.MapClaims{
		"name":       name,
		"avatar_url": userInfo.AvatarURL,
		"github_id":  userInfo.ID,
		"login":      userInfo.Login,
		"iat":        time.Now().Unix(),
		"exp":        time.Now().Add(time.Second * time.Duration(cfg.JwtExpiry)).Unix(), // Token expires in configured Seconds
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JwtSecret))
}
