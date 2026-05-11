package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(redisURL string) (*RateLimiter, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	return &RateLimiter{client: redis.NewClient(opts)}, nil
}

func (rl *RateLimiter) Limit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			key := fmt.Sprintf("ratelimit:%s:%s", r.URL.Path, ip)

			ctx := context.Background()
			count, err := rl.client.Incr(ctx, key).Result()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				rl.client.Expire(ctx, key, window)
			}

			if count > int64(maxRequests) {
				http.Error(w, `{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Too many requests"}}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
