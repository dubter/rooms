package rate_limiter

import (
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// visitor holds limiter and lastSeen for specific user.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// rateLimiter used to rate limit an incoming requests.
type rateLimiter struct {
	sync.RWMutex

	visitors map[string]*visitor
	limit    rate.Limit
	burst    int
	ttl      time.Duration
}

// newRateLimiter creates an instance of the rateLimiter.
func newRateLimiter(rps, burst int, ttl time.Duration) *rateLimiter {
	return &rateLimiter{
		visitors: make(map[string]*visitor),
		limit:    rate.Limit(rps),
		burst:    burst,
		ttl:      ttl,
	}
}

// getVisitor returns limiter for the specific visitor by its IP,
// looking up within the visitors map.
func (l *rateLimiter) getVisitor(ip string) *rate.Limiter {
	l.RLock()
	v, exists := l.visitors[ip]
	l.RUnlock()

	if !exists {
		limiter := rate.NewLimiter(l.limit, l.burst)
		l.Lock()
		l.visitors[ip] = &visitor{limiter, time.Now()}
		l.Unlock()

		return limiter
	}

	v.lastSeen = time.Now()

	return v.limiter
}

// cleanupVisitors removes old entries from the visitors map.
func (l *rateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		l.Lock()
		for ip, v := range l.visitors {
			if time.Since(v.lastSeen) > l.ttl {
				delete(l.visitors, ip)
			}
		}
		l.Unlock()
	}
}

// Limit creates a new rate limiter middleware handler.
func Limit(rps int, burst int, ttl time.Duration, logger *slog.Logger) func(http.Handler) http.Handler {
	l := newRateLimiter(rps, burst, ttl)

	// run a background worker to clean up old entries
	go l.cleanupVisitors()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				logger.Error("Rate limiter", slog.String("error", err.Error()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if !l.getVisitor(ip).Allow() {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
