package service

import (
	"VPSBenchmarkBackend/internal/auth/model"
	"VPSBenchmarkBackend/internal/auth/store"
	"VPSBenchmarkBackend/internal/cache"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GithubTokenResponse represents the response from GitHub OAuth token endpoint
type GithubTokenResponse struct {
	AccessToken string `json:"access_token"` // Ignore the refresh token because we only use the token to fetch user info once
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

type AuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// GithubLogin exchanges a GitHub OAuth code for a JWT token
func GithubLogin(code string) (*AuthToken, error) {
	cfg := config.Get()

	// Exchange code for access token
	accessToken, err := exchangeCodeForToken(code, cfg)
	if err != nil {
		return nil, err
	}

	// Get user info from GitHub
	userInfo, err := getGithubUserInfo(accessToken)
	if err != nil {
		return nil, err
	}

	userRecord, err := store.GetUserByID(userInfo.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	name := userInfo.Name
	if name == "" {
		name = userInfo.Login
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		gid := store.DefaultUserGroupId
		if cfg.AdminID == userInfo.ID {
			gid = store.DefaultAdminGroupId
		}
		user := model.User{
			ID:      userInfo.ID,
			Name:    name,
			Login:   userInfo.Login,
			GroupID: gid, // Default group ID, can be updated later by admin
		}
		// Create user in database if not exists
		if err = store.CreateUser(&user); err != nil {
			return nil, err
		}
	} else {
		userRecord = model.User{
			ID:      userRecord.ID,
			Name:    name,
			Login:   userInfo.Login,
			GroupID: userRecord.GroupID, // Keep existing group ID
		}
		// Update user info in database if exists
		if _, err = store.UpdateUser(userRecord); err != nil {
			return nil, err
		}
	}

	// Generate token
	return generateToken(userInfo, cfg)
}

func RefreshToken(userID int64, refreshToken string) (*AuthToken, error) {
	key := fmt.Sprintf("token:%d:%s", userID, refreshToken)
	client := cache.GetClient()
	result, err := client.Get(context.Background(), key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			cursor := uint64(0)
			var keys []string
			pattern := fmt.Sprintf("token:%d:*", userID)
			for {
				keys, cursor, err = client.Scan(context.Background(), 0, pattern, 100).Result()
				if err != nil {
					return nil, err
				}
				if cursor == 0 || len(keys) == 0 {
					break
				}
				if len(keys) > 0 {
					client.Unlink(context.Background(), keys...)
				}
			}
			return nil, &common.InvalidParamError{Message: "Invalid refresh token"}
		} else {
			return nil, err
		}
	}
	// Delete the old refresh token
	if err = client.Del(context.Background(), key).Err(); err != nil {
		return nil, err
	}

	// Generate new token
	userInfo := &GithubUserInfo{}
	if err = json.Unmarshal([]byte(result), userInfo); err != nil {
		return nil, err
	}
	return generateToken(userInfo, config.Get())
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
	if config.Get().GithubHttpProxy == "" {
		return &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	proxyURL, _ := url.Parse(config.Get().GithubHttpProxy)
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
}

// generateToken generates a token for the user
func generateToken(userInfo *GithubUserInfo, cfg *config.Config) (*AuthToken, error) {
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
		"exp":        time.Now().Add(time.Second * time.Duration(cfg.AccessTokenExp)).Unix(), // Token expires in configured Seconds
	}
	claims.GetExpirationTime()

	accessJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	randID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	randIDStr := randID.String()

	accessToken, err := accessJWT.SignedString([]byte(cfg.JwtSecret))
	if err != nil {
		return nil, err
	}
	refreshExp := time.Duration(cfg.RefreshTokenExp) * time.Second
	refreshClaims := jwt.MapClaims{
		"rand_id":   randIDStr,
		"github_id": userInfo.ID,
		"iat":       time.Now().Unix(),
		"exp":       time.Now().Add(refreshExp - 120*time.Second).Unix(), // Refresh token expires in configured Seconds, but we set the JWT exp to be 2 minutes earlier to allow some buffer for token refresh before it actually expires
	}
	refreshJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshJWT.SignedString([]byte(cfg.JwtSecret))
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("token:%d:%s", userInfo.ID, randIDStr)
	infoJson, err := json.Marshal(userInfo)
	if err != nil {
		return nil, err
	}
	err = cache.GetClient().Set(context.Background(), key, string(infoJson), refreshExp).Err()
	if err != nil {
		return nil, err
	}

	return &AuthToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
