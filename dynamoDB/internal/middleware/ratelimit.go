package middleware

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
	appErrors "github.com/ramon/goals-tasks-api/internal/errors"
	"github.com/ramon/goals-tasks-api/internal/shared"
)

type ipLimiter struct {
	ips map[string]*rate.Limiter
	mu  sync.RWMutex
	r   rate.Limit
	b   int
}

func newIPLimiter(r rate.Limit, b int) *ipLimiter {
	return &ipLimiter{
		ips: make(map[string]*rate.Limiter),
		r:   r,
		b:   b,
	}
}

func (i *ipLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if exists {
		return limiter
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists = i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

func RateLimit(r rate.Limit, b int) func(http.Handler) http.Handler {
	limiter := newIPLimiter(r, b)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			lim := limiter.getLimiter(ip)
			if !lim.Allow() {
				traceID := shared.GetCorrelationID(r.Context())
				shared.SendError(w, appErrors.ErrRateLimitExceeded, traceID)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
