package middleware

import (
	"net/http"
	"sync"
	"time"

	"mowsy-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int
	window   time.Duration
}

type Visitor struct {
	requests []time.Time
	mu       sync.RWMutex
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		window:   window,
	}

	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for key, visitor := range rl.visitors {
				visitor.mu.Lock()
				cutoff := time.Now().Add(-rl.window)
				validRequests := []time.Time{}
				
				for _, t := range visitor.requests {
					if t.After(cutoff) {
						validRequests = append(validRequests, t)
					}
				}
				
				if len(validRequests) == 0 {
					delete(rl.visitors, key)
				} else {
					visitor.requests = validRequests
				}
				visitor.mu.Unlock()
			}
			rl.mu.Unlock()
		}
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.RLock()
	visitor, exists := rl.visitors[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		visitor = &Visitor{
			requests: []time.Time{},
		}
		rl.visitors[key] = visitor
		rl.mu.Unlock()
	}

	visitor.mu.Lock()
	defer visitor.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)
	
	validRequests := []time.Time{}
	for _, t := range visitor.requests {
		if t.After(cutoff) {
			validRequests = append(validRequests, t)
		}
	}
	
	if len(validRequests) >= rl.rate {
		return false
	}
	
	visitor.requests = append(validRequests, now)
	return true
}

func RateLimitMiddleware(rate int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, window)
	
	return func(c *gin.Context) {
		key := c.ClientIP()
		
		if userID, exists := c.Get("user_id"); exists {
			key = "user_" + string(rune(userID.(uint)))
		}
		
		if !limiter.Allow(key) {
			utils.ErrorResponse(c, http.StatusTooManyRequests, "Rate limit exceeded")
			c.Abort()
			return
		}
		
		c.Next()
	}
}