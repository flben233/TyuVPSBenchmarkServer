package service

import (
	"fmt"
	"sync"
)

const MaxSessionsPerUser = 3

type RateLimiter struct {
	mu       sync.Mutex
	sessions map[int64]int
}

var limiter = &RateLimiter{
	sessions: make(map[int64]int),
}

func AcquireSession(userID int64) error {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	count := limiter.sessions[userID]
	if count >= MaxSessionsPerUser {
		return fmt.Errorf("too many concurrent sessions (max %d)", MaxSessionsPerUser)
	}
	limiter.sessions[userID] = count + 1
	return nil
}

func ReleaseSession(userID int64) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	count := limiter.sessions[userID]
	if count > 0 {
		limiter.sessions[userID] = count - 1
	}
	if limiter.sessions[userID] == 0 {
		delete(limiter.sessions, userID)
	}
}
