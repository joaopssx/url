package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	sync.Mutex
	count       int
	windowStart time.Time
}

type RateLimiter struct {
	ips   sync.Map
	limit int
}

func NewRateLimiter(limit int) *RateLimiter {
	rl := &RateLimiter{
		limit: limit,
	}

	go func() {
		for {
			time.Sleep(5 * time.Minute)
			now := time.Now()
			rl.ips.Range(func(key, value interface{}) bool {
				entry := value.(*rateLimitEntry)
				entry.Lock()
				if now.Sub(entry.windowStart) > 1*time.Minute {
					entry.Unlock()
					rl.ips.Delete(key)
				} else {
					entry.Unlock()
				}
				return true
			})
		}
	}()

	return rl
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)

		value, _ := rl.ips.LoadOrStore(ip, &rateLimitEntry{
			count:       0,
			windowStart: time.Now(),
		})
		entry := value.(*rateLimitEntry)

		entry.Lock()
		defer entry.Unlock()

		now := time.Now()
		if now.Sub(entry.windowStart) > 1*time.Minute {
			entry.count = 0
			entry.windowStart = now
		}

		entry.count++

		if entry.count > rl.limit {
			retryAfter := int(time.Minute.Seconds() - now.Sub(entry.windowStart).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}

		c.Next()
	}
}

func getClientIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}
