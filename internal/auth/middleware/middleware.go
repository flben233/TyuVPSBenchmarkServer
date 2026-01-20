package middleware

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth is a middleware that validates JWT tokens
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "Authorization header is required"))
			c.Abort()
			return
		}

		// Check for Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "Invalid authorization header format, expected 'Bearer <token>'"))
			c.Abort()
			return
		}

		tokenString := parts[1]
		cfg := config.Get()

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "Invalid token: "+err.Error()))
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "Invalid token"))
			c.Abort()
			return
		}

		// Extract claims and set them in the context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if name, ok := claims["name"].(string); ok {
				c.Set("user_name", name)
			}
			if login, ok := claims["login"].(string); ok {
				c.Set("user_login", login)
			}
			if avatarURL, ok := claims["avatar_url"].(string); ok {
				c.Set("user_avatar_url", avatarURL)
			}
		}

		c.Next()
	}
}

func CheckAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userLogin, exists := c.Get("user_login")
		if !exists {
			c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
			c.Abort()
			return
		}
		cfg := config.Get()
		if userLogin.(string) != cfg.AdminID {
			c.JSON(http.StatusForbidden, common.Error(common.ForbiddenCode, "User is not an admin"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// OptionalJWTAuth is a middleware that validates JWT tokens if present, but doesn't require them
func OptionalJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Check for Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		cfg := config.Get()

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.Next()
			return
		}

		// Extract claims and set them in the context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if name, ok := claims["name"].(string); ok {
				c.Set("user_name", name)
			}
			if avatarURL, ok := claims["avatar_url"].(string); ok {
				c.Set("user_avatar_url", avatarURL)
			}
		}

		c.Next()
	}
}

func AllowCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", config.Get().FrontendURL)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, Accept, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
