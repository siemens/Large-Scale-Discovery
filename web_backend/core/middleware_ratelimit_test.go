/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/config"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
)

// mwInjectContext sets up the context storage expected by MwRateLimit. Call this before MwRateLimit.
func mwInjectContext(user *database.T_user) gin.HandlerFunc {

	// Return a middleware handler that populates the request context storage.
	return func(ctx *gin.Context) {

		// Attach a test logger and the configured user to the shared context storage.
		SetContextStorage(ctx, &ContextStorage{
			Logger:      scanUtils.NewTestLogger(),
			CurrentUser: user,
		})
		ctx.Next()
	}
}

// newTestRouter builds a minimal Gin router for rate limit testing.
// noAuthPaths is registered in apiEndpointsNoAuth so the middleware treats them accordingly.
// user is attached to the context (nil = unauthenticated).
func newTestRouter(rateLimitConf config.RateLimit, noAuthPaths []string, user *database.T_user) *gin.Engine {

	// Switch gin to test mode to suppress debug output.
	gin.SetMode(gin.TestMode)

	// Register no-auth paths in the package-level slice the middleware checks.
	apiEndpointsNoAuth = noAuthPaths

	// Install the rate limit config.
	config.Set(&config.WebConfig{
		RateLimit: rateLimitConf,
	})

	// Build a minimal router with context injection and rate limiting.
	r := gin.New()
	r.Use(mwInjectContext(user))
	r.Use(MwRateLimit())
	r.GET("/api/v1/noauth", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/api/v1/auth", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Return router
	return r
}

// do sends a GET request to the router and returns the recorded response.
func do(r *gin.Engine, path, clientIP string) *httptest.ResponseRecorder {

	// Build a GET request and set RemoteAddr so ctx.ClientIP returns the expected address.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	if clientIP != "" {
		req.RemoteAddr = clientIP + ":0"
	}

	// Serve the request and return the recorded response.
	r.ServeHTTP(w, req)
	return w
}

// TestRateLimiterStore_GetCreatesEntry verifies that get creates a new entry for an unseen key.
func TestRateLimiterStore_GetCreatesEntry(t *testing.T) {

	// Initialize the store and release it at end of test.
	s := newRateLimiterStore(60, time.Minute)
	defer s.stop()

	// Verify that get creates an entry for a new key.
	e := s.get("key1")
	if e == nil {
		t.Fatal("expected entry, got nil")
	}
}

// TestRateLimiterStore_GetReusesEntry verifies that get returns the same entry for repeated calls.
func TestRateLimiterStore_GetReusesEntry(t *testing.T) {

	// Initialize the store and release it at end of test.
	s := newRateLimiterStore(60, time.Minute)
	defer s.stop()

	// Verify that get returns the same pointer for repeated calls with the same key.
	e1 := s.get("key1")
	e2 := s.get("key1")
	if e1 != e2 {
		t.Error("expected same entry to be returned on second call")
	}
}

// TestRateLimiterStore_SeparateKeys verifies that different keys produce independent limiter entries.
func TestRateLimiterStore_SeparateKeys(t *testing.T) {

	// Initialize the store and release it at end of test.
	s := newRateLimiterStore(60, time.Minute)
	defer s.stop()

	// Verify that different keys produce independent entries.
	e1 := s.get("key1")
	e2 := s.get("key2")
	if e1 == e2 {
		t.Error("expected distinct entries for different keys")
	}
}

// TestRateLimiterStore_Cleanup verifies that stale entries are evicted after the TTL elapses.
func TestRateLimiterStore_Cleanup(t *testing.T) {

	// Use a short TTL so cleanup runs within the test duration.
	ttl := 50 * time.Millisecond
	s := newRateLimiterStore(60, ttl)
	defer s.stop()

	// Populate the store with one entry, then wait long enough for cleanup to evict it.
	s.get("key1")
	time.Sleep(3 * ttl)

	// Verify the stale entry was removed.
	s.mu.Lock()
	_, still := s.entries["key1"]
	s.mu.Unlock()

	if still {
		t.Error("expected stale entry to be cleaned up")
	}
}

// TestMwRateLimit_AuthEndpoint_AllowsUpToLimit verifies that requests within the burst limit all succeed.
func TestMwRateLimit_AuthEndpoint_AllowsUpToLimit(t *testing.T) {

	// Build a router limited to 3 auth requests per minute.
	r := newTestRouter(config.RateLimit{
		AuthRequestsPerMinute: 3,
		ApiRequestsPerMinute:  120,
	}, []string{"/api/v1/noauth"}, nil)

	// All requests within the burst limit must succeed.
	for i := 0; i < 3; i++ {
		w := do(r, "/api/v1/noauth", "10.0.0.1")
		if w.Code != http.StatusOK {
			t.Errorf("request %d: want 200, got %d", i+1, w.Code)
		}
	}
}

// TestMwRateLimit_AuthEndpoint_Returns429WhenExceeded verifies that exceeding the auth limit returns 429.
func TestMwRateLimit_AuthEndpoint_Returns429WhenExceeded(t *testing.T) {

	// Build a router limited to 2 auth requests per minute.
	r := newTestRouter(config.RateLimit{
		AuthRequestsPerMinute: 2,
		ApiRequestsPerMinute:  120,
	}, []string{"/api/v1/noauth"}, nil)

	// Exhaust the 2-request burst.
	do(r, "/api/v1/noauth", "10.1.0.1")
	do(r, "/api/v1/noauth", "10.1.0.1")

	// Third request must be rate-limited.
	w := do(r, "/api/v1/noauth", "10.1.0.1")
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("want 429, got %d", w.Code)
	}
}

// TestMwRateLimit_AuthEndpoint_SeparatesIPs verifies that each client IP has its own independent bucket.
func TestMwRateLimit_AuthEndpoint_SeparatesIPs(t *testing.T) {

	// Build a router limited to 1 auth request per minute per IP.
	r := newTestRouter(config.RateLimit{
		AuthRequestsPerMinute: 1,
		ApiRequestsPerMinute:  120,
	}, []string{"/api/v1/noauth"}, nil)

	// IP A exhausts its bucket.
	do(r, "/api/v1/noauth", "10.2.0.1")
	wA := do(r, "/api/v1/noauth", "10.2.0.1")

	// IP B should still succeed since it has its own independent bucket.
	wB := do(r, "/api/v1/noauth", "10.2.0.2")

	if wA.Code != http.StatusTooManyRequests {
		t.Errorf("IP A: want 429, got %d", wA.Code)
	}
	if wB.Code != http.StatusOK {
		t.Errorf("IP B: want 200, got %d", wB.Code)
	}
}

// TestMwRateLimit_ApiEndpoint_Returns429WhenExceeded verifies that exceeding the per-user API limit returns 429.
func TestMwRateLimit_ApiEndpoint_Returns429WhenExceeded(t *testing.T) {

	// Build a router limited to 2 authenticated API requests per minute.
	user := &database.T_user{Id: 42}
	r := newTestRouter(config.RateLimit{
		AuthRequestsPerMinute: 10,
		ApiRequestsPerMinute:  2,
	}, []string{"/api/v1/noauth"}, user)

	// Exhaust the 2-request burst.
	do(r, "/api/v1/auth", "")
	do(r, "/api/v1/auth", "")

	// Third request must be rate-limited.
	w := do(r, "/api/v1/auth", "")
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("want 429, got %d", w.Code)
	}
}

// TestMwRateLimit_RateLimitHeaders_Present verifies that all three X-RateLimit-* headers are present.
func TestMwRateLimit_RateLimitHeaders_Present(t *testing.T) {

	// Build a router and send one authenticated request.
	r := newTestRouter(config.RateLimit{
		AuthRequestsPerMinute: 10,
		ApiRequestsPerMinute:  120,
	}, []string{"/api/v1/noauth"}, &database.T_user{Id: 1})

	w := do(r, "/api/v1/auth", "")

	// Verify all three rate limit headers are present.
	if w.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("X-RateLimit-Limit header missing")
	}
	if w.Header().Get("X-RateLimit-Remaining") == "" {
		t.Error("X-RateLimit-Remaining header missing")
	}
	if w.Header().Get("X-RateLimit-Reset") == "" {
		t.Error("X-RateLimit-Reset header missing")
	}
}

// TestMwRateLimit_ZeroLimit_Disabled verifies that a zero limit disables rate limiting entirely.
func TestMwRateLimit_ZeroLimit_Disabled(t *testing.T) {

	// Build a router with both limits set to zero, which disables rate limiting.
	r := newTestRouter(config.RateLimit{
		AuthRequestsPerMinute: 0,
		ApiRequestsPerMinute:  0,
	}, []string{"/api/v1/noauth"}, nil)

	// Many requests should all succeed when limiting is disabled.
	for i := 0; i < 20; i++ {
		w := do(r, "/api/v1/noauth", "10.3.0.1")
		if w.Code != http.StatusOK {
			t.Errorf("request %d: want 200 (limit disabled), got %d", i+1, w.Code)
		}
	}
}

// TestMwRateLimit_NonApiPath_PassesThrough verifies that non-API paths bypass rate limiting entirely.
func TestMwRateLimit_NonApiPath_PassesThrough(t *testing.T) {

	// Build a router with aggressive limits and register a static-file route.
	r := newTestRouter(config.RateLimit{
		AuthRequestsPerMinute: 1,
		ApiRequestsPerMinute:  1,
	}, []string{}, nil)
	r.GET("/static/app.js", func(c *gin.Context) { c.Status(http.StatusOK) })

	// All static-file requests must pass through regardless of the limit.
	for i := 0; i < 5; i++ {
		w := do(r, "/static/app.js", "10.4.0.1")
		if w.Code != http.StatusOK {
			t.Errorf("request %d: want 200 for static path, got %d", i+1, w.Code)
		}
	}
}

// TestMwRateLimit_429Response_HasRetryAfterHeader verifies that 429 responses include a Retry-After header.
func TestMwRateLimit_429Response_HasRetryAfterHeader(t *testing.T) {

	// Build a router limited to 1 auth request per minute.
	r := newTestRouter(config.RateLimit{
		AuthRequestsPerMinute: 1,
		ApiRequestsPerMinute:  120,
	}, []string{"/api/v1/noauth"}, nil)

	// Consume the single available token, then trigger a 429.
	do(r, "/api/v1/noauth", "10.5.0.1")
	w := do(r, "/api/v1/noauth", "10.5.0.1")

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429, got %d", w.Code)
	}

	// Verify the Retry-After header is present on the 429 response.
	if w.Header().Get("Retry-After") == "" {
		t.Error("Retry-After header missing on 429 response")
	}
}

// TestMwRateLimit_LooksLikeApiButNotApi_PassesThrough verifies that paths starting with "/api" but not "/api/" bypass rate limiting entirely and do not panic with a nil user.
func TestMwRateLimit_LooksLikeApiButNotApi_PassesThrough(t *testing.T) {

	// Build a router with aggressive limits so any accidentally rate-limited request surfaces as non-200.
	r := newTestRouter(config.RateLimit{
		AuthRequestsPerMinute: 1,
		ApiRequestsPerMinute:  1,
	}, []string{}, nil)
	r.GET("/api-docs/", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/apidocs/", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/api-versioning", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Prepare and run test cases
	tests := []struct {
		name string
		path string
	}{
		{"api-docs", "/api-docs/"},
		{"apidocs", "/apidocs/"},
		{"api-versioning", "/api-versioning"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			// Verify the path passes through without triggering rate limiting or panicking.
			w := do(r, tc.path, "10.6.0.1")
			if w.Code != http.StatusOK {
				t.Errorf("MwRateLimit() code = '%v', want = '200'", w.Code)
			}
		})
	}
}

// TestMwRateLimit_AuthRouteWithoutUser_DoesNotPanic verifies that an authenticated route reached with a nil user returns 401 rather than panicking.
func TestMwRateLimit_AuthRouteWithoutUser_DoesNotPanic(t *testing.T) {

	// Build a router with nil user to simulate a request that bypassed JWT authentication.
	r := newTestRouter(config.RateLimit{
		AuthRequestsPerMinute: 10,
		ApiRequestsPerMinute:  10,
	}, []string{}, nil)

	// Verify the middleware responds with 401 instead of panicking.
	w := do(r, "/api/v1/auth", "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("MwRateLimit() code = '%v', want = '401'", w.Code)
	}
}
