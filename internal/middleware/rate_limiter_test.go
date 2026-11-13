package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIPRateLimiter(t *testing.T) {
	limiter := NewIPRateLimiter(2.5, 4)

	require.NotNil(t, limiter)
	assert.Equal(t, 2.5, float64(limiter.rate))
	assert.Equal(t, 4, limiter.burst)
	assert.NotNil(t, limiter.clients)
	assert.Equal(t, time.Minute, limiter.cleanupInterval)
	assert.Equal(t, 10*time.Minute, limiter.idleTimeout)
}

func TestIPRateLimiterMiddleware_AllowsRequestsWithinBurst(t *testing.T) {
	limiter := NewIPRateLimiter(0, 2)
	nextCalls := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalls++
		w.WriteHeader(http.StatusNoContent)
	})
	handler := limiter.Middleware(next)

	for i := 0; i < 2; i++ {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = "192.0.2.1:12345"

		handler.ServeHTTP(response, request)

		assert.Equal(t, http.StatusNoContent, response.Code)
	}

	assert.Equal(t, 2, nextCalls)
}

func TestIPRateLimiterMiddleware_RejectsRequestsOverLimit(t *testing.T) {
	limiter := NewIPRateLimiter(0, 1)
	nextCalls := 0
	handler := limiter.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalls++
	}))

	firstResponse := httptest.NewRecorder()
	firstRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	firstRequest.RemoteAddr = "192.0.2.10:12345"
	handler.ServeHTTP(firstResponse, firstRequest)
	require.Equal(t, http.StatusOK, firstResponse.Code)

	secondResponse := httptest.NewRecorder()
	secondRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	secondRequest.RemoteAddr = "192.0.2.10:54321"
	handler.ServeHTTP(secondResponse, secondRequest)

	assert.Equal(t, http.StatusTooManyRequests, secondResponse.Code)
	assert.Equal(t, "rate limit exceeded\n", secondResponse.Body.String())
	assert.Equal(t, 1, nextCalls)
}

func TestIPRateLimiterMiddleware_LimitsClientsIndependently(t *testing.T) {
	limiter := NewIPRateLimiter(0, 1)
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, ip := range []string{"192.0.2.1:1234", "192.0.2.2:1234"} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = ip
		handler.ServeHTTP(response, request)
		assert.Equal(t, http.StatusNoContent, response.Code)
	}
}

func TestIPRateLimiterMiddleware_UsesRemoteAddressWithoutPort(t *testing.T) {
	limiter := NewIPRateLimiter(0, 1)
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := 0; i < 2; i++ {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = "192.0.2.3"
		handler.ServeHTTP(response, request)
		if i == 0 {
			assert.Equal(t, http.StatusNoContent, response.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, response.Code)
		}
	}
}

func TestIPRateLimiterGetLimiter_RefreshesLastSeen(t *testing.T) {
	limiter := NewIPRateLimiter(1, 1)

	first := limiter.getLimiter("192.0.2.1")
	firstSeen := limiter.clients["192.0.2.1"].lastSeen
	time.Sleep(time.Millisecond)
	second := limiter.getLimiter("192.0.2.1")

	assert.Same(t, first, second)
	assert.True(t, limiter.clients["192.0.2.1"].lastSeen.After(firstSeen))
}

func TestIPRateLimiterCleanup_RemovesIdleClients(t *testing.T) {
	limiter := NewIPRateLimiter(1, 1)
	now := time.Now()
	limiter.cleanupInterval = time.Minute
	limiter.idleTimeout = 10 * time.Minute
	limiter.lastCleanup = now.Add(-time.Minute)
	limiter.clients["idle"] = &clientLimiter{lastSeen: now.Add(-11 * time.Minute)}
	limiter.clients["active"] = &clientLimiter{lastSeen: now.Add(-9 * time.Minute)}

	limiter.cleanup(now)

	_, idleExists := limiter.clients["idle"]
	_, activeExists := limiter.clients["active"]
	assert.False(t, idleExists)
	assert.True(t, activeExists)
	assert.Equal(t, now, limiter.lastCleanup)
}

func TestIPRateLimiterCleanup_RespectsCleanupInterval(t *testing.T) {
	limiter := NewIPRateLimiter(1, 1)
	now := time.Now()
	limiter.lastCleanup = now.Add(-time.Second)
	limiter.clients["idle"] = &clientLimiter{lastSeen: now.Add(-time.Hour)}

	limiter.cleanup(now)

	_, exists := limiter.clients["idle"]
	assert.True(t, exists)
}
