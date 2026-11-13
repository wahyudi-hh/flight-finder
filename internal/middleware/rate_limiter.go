package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type clientLimiter struct {
	limiter *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu				sync.Mutex
	clients			map[string]*clientLimiter
	rate			rate.Limit
	burst			int
	lastCleanup		time.Time
	cleanupInterval	time.Duration
	idleTimeout		time.Duration
}

func NewIPRateLimiter(requestsPerSecond float64, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		clients: make(map[string]*clientLimiter),
		rate: rate.Limit(requestsPerSecond),
		burst: burst,
		cleanupInterval: time.Minute,
		idleTimeout: 10 * time.Minute,
	}
}

func (rl *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	rl.cleanup(now)

	client, exists := rl.clients[ip]
    if !exists {
        client = &clientLimiter{
            limiter: rate.NewLimiter(rl.rate, rl.burst),
        }
        rl.clients[ip] = client
    }

    client.lastSeen = now

    return client.limiter
}

func (rl *IPRateLimiter) cleanup(now time.Time) {
	if now.Sub(rl.lastCleanup) >= rl.cleanupInterval {
		for clientIP, client := range rl.clients {
			if now.Sub(client.lastSeen) > rl.idleTimeout {
				delete(rl.clients, clientIP)
			}
		}
		rl.lastCleanup = now
	}
}

func (rl *IPRateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip, _, err := net.SplitHostPort(r.RemoteAddr)
        if err != nil {
            ip = r.RemoteAddr
        }

        limiter := rl.getLimiter(ip)

        if !limiter.Allow() {
            http.Error(
                w,
                "rate limit exceeded",
                http.StatusTooManyRequests,
            )
            return
        }

        next.ServeHTTP(w, r)
    })
}