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
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/config"
	"golang.org/x/time/rate"
)

// rateLimitEntry holds a token bucket limiter and the last time it was used, for cleanup purposes.
// The embedded rate.Limiter is concurrency-safe; lastSeen is always accessed under the store mutex.
type rateLimitEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// rateLimiterStore manages per-key token bucket limiters with periodic cleanup.
type rateLimiterStore struct {
	mu       sync.Mutex
	entries  map[string]*rateLimitEntry
	rps      rate.Limit // tokens per second
	burst    int        // max burst size (= per-minute limit)
	ttl      time.Duration
	stopOnce sync.Once
	stopCh   chan struct{}
}

// newRateLimiterStore creates a new rate limiter store with a background cleanup goroutine.
func newRateLimiterStore(requestsPerMinute int, ttl time.Duration) *rateLimiterStore {

	// Convert requests-per-minute to a per-second rate. Use infinite rate when limiting is disabled.
	rps := rate.Inf
	if requestsPerMinute > 0 {
		rps = rate.Limit(float64(requestsPerMinute) / 60.0)
	}

	// Initialize the store and allocate the stop channel used by the cleanup goroutine.
	s := &rateLimiterStore{
		entries: make(map[string]*rateLimitEntry),
		rps:     rps,
		burst:   requestsPerMinute,
		ttl:     ttl,
		stopCh:  make(chan struct{}),
	}

	// Start the background goroutine that periodically evicts stale entries.
	go s.cleanup()

	// Return store
	return s
}

// get returns the rate limit entry for the given key, creating one if it does not exist.
func (s *rateLimiterStore) get(key string) *rateLimitEntry {

	// Serialize all access to the shared entries map.
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a new limiter entry for keys that have not been seen before.
	entry, ok := s.entries[key]
	if !ok {
		entry = &rateLimitEntry{
			limiter: rate.NewLimiter(s.rps, s.burst),
		}
		s.entries[key] = entry
	}

	// Refresh the timestamp to prevent premature eviction by the cleanup goroutine.
	entry.lastSeen = time.Now()

	// Return entry
	return entry
}

// cleanup runs as a goroutine and periodically removes entries not accessed within the TTL.
func (s *rateLimiterStore) cleanup() {

	// Tick at the TTL interval so every entry is eligible for eviction after one full period.
	ticker := time.NewTicker(s.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			// Remove all entries that have not been accessed within the TTL window.
			cutoff := time.Now().Add(-s.ttl)
			s.mu.Lock()
			for key, entry := range s.entries {
				if entry.lastSeen.Before(cutoff) {
					delete(s.entries, key)
				}
			}
			s.mu.Unlock()

		case <-s.stopCh:
			return
		}
	}
}

// stop signals the cleanup goroutine to exit.
func (s *rateLimiterStore) stop() {

	// Close the stop channel exactly once to signal the cleanup goroutine to exit.
	s.stopOnce.Do(func() { close(s.stopCh) })
}

// isApiPath checks whether the URI is an API URI.
func isApiPath(path string) bool {
	return strings.HasPrefix(path, "/api/")
}

// rateLimitHeaders writes standard X-RateLimit-* headers to a response.
// Remaining and limit are in requests-per-minute terms.
func rateLimitHeaders(ctx *gin.Context, limitPerMin int, tokens float64) {
	remaining := int(tokens)
	if remaining < 0 {
		remaining = 0
	}
	ctx.Header("X-RateLimit-Limit", strconv.Itoa(limitPerMin))
	ctx.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	ctx.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10)) // Reset is the time when the bucket will be full again; approximate as 1 minute from now.
}

// MwRateLimit returns a Gin middleware that applies rate limiting after JWT authentication.
// - Routes in apiEndpointsNoAuth (e.g. /backend/authenticator) are limited per source IP.
// - All other API routes are limited per authenticated user ID.
// Limits and behavior are driven by config.RateLimit; zero values disable the respective limit.
// The middleware must be registered after MwJwtAuthentication so the current user is available.
func MwRateLimit() gin.HandlerFunc {

	// Get config
	conf := config.GetConfig()

	// Read config values
	var authLimit = conf.RateLimit.AuthRequestsPerMinute
	var apiLimit = conf.RateLimit.ApiRequestsPerMinute

	// Prepare memory
	var authStore = newRateLimiterStore(authLimit, 5*time.Minute)
	var apiStore = newRateLimiterStore(apiLimit, 5*time.Minute)

	// Return request handling function
	return func(ctx *gin.Context) {

		// Only rate-limit API paths and pass through others such as static files.
		if !isApiPath(ctx.Request.URL.Path) {
			ctx.Next()
			return
		}

		// Get logger for current request context
		logger := GetContextLogger(ctx)

		// Check if route requires authentication
		isNoAuth := scanUtils.StrContained(ctx.Request.URL.Path, apiEndpointsNoAuth)

		// Check rate limits
		if isNoAuth {

			// Get limiter by IP
			clientIp := ctx.ClientIP()
			entry := authStore.get(clientIp)

			// rate.Limiter is concurrency-safe, no additional locking needed.
			reservation := entry.limiter.Reserve()
			tokens := entry.limiter.Tokens()

			// Write rate limit headers
			rateLimitHeaders(ctx, authLimit, tokens)

			// Check rate limits
			if !reservation.OK() || reservation.Delay() > 0 {

				// Cancel rate limit reservation
				reservation.Cancel()

				// Log exceeded rate limit
				logger.Debugf("Rate limit exceeded for IP '%s'.", clientIp)

				// Add supportive response header
				ctx.Header("Retry-After", strconv.Itoa(int(reservation.Delay().Seconds()+1)))

				// Send rate limit error response
				RespondRateLimit(ctx) // Return rate limit error. Situation already logged!
				return
			}
		} else {

			// Get limiter by user
			user := GetContextUser(ctx)

			// Defense in depth: if auth middleware failed to set a user, reject with 401 rather than panicking.
			if user == nil {
				RespondAuthError(ctx)
				return
			}

			// Get the per-user token bucket for the authenticated API rate limit.
			entry := apiStore.get(fmt.Sprintf("%d", user.Id))

			// rate.Limiter is concurrency-safe, no additional locking needed.
			reservation := entry.limiter.Reserve()
			tokens := entry.limiter.Tokens()

			// Write rate limit headers
			rateLimitHeaders(ctx, apiLimit, tokens)

			// Check rate limits
			if !reservation.OK() || reservation.Delay() > 0 {

				// Cancel rate limit reservation
				reservation.Cancel()

				// Log exceeded rate limit
				logger.Debugf("Rate limit exceeded for user '%d'.", user.Id)

				// Add supportive response header
				ctx.Header("Retry-After", strconv.Itoa(int(reservation.Delay().Seconds()+1)))

				// Send rate limit error response
				RespondRateLimit(ctx) // Return rate limit error. Situation already logged!
				return
			}
		}

		// Success, step to next middleware plugin, if existing
		ctx.Next() // Should only be called in middleware handlers!
	}
}
