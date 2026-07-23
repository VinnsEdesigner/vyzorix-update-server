// Package middleware provides HTTP middleware.
package middleware

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// getEnv returns an environment variable with a default value.
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return defaultVal
}

// getEnvInt returns an integer from an environment variable.
func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}

	return defaultVal
}

// IPIntelligenceConfig holds configuration for IP intelligence.
type IPIntelligenceConfig struct {
	AbuseIPDBAPIKey              string
	Whitelist                    string
	AbuseIPDBConfidenceThreshold int
	MaxFailedAttempts            int
	BlockDuration                time.Duration
	WindowDuration               time.Duration
	Enabled                      bool
	AbuseIPDBEnabled             bool
}

// DefaultIPIntelligenceConfig returns sensible defaults.
func DefaultIPIntelligenceConfig() IPIntelligenceConfig {
	return IPIntelligenceConfig{
		Enabled:                      true,
		MaxFailedAttempts:            10,
		BlockDuration:                15 * time.Minute,
		WindowDuration:               1 * time.Hour,
		Whitelist:                    "",
		AbuseIPDBEnabled:             false,
		AbuseIPDBConfidenceThreshold: 50,
	}
}

// LoadIPIntelligenceConfig loads from environment variables.
func LoadIPIntelligenceConfig() IPIntelligenceConfig {
	cfg := DefaultIPIntelligenceConfig()
	cfg.Enabled = getEnvBool("IP_INTELLIGENCE_ENABLED", true)
	cfg.AbuseIPDBEnabled = getEnvBool("ABUSEIPDB_ENABLED", false)
	cfg.AbuseIPDBAPIKey = getEnv("ABUSEIPDB_API_KEY", "")
	cfg.AbuseIPDBConfidenceThreshold = getEnvInt("ABUSEIPDB_CONFIDENCE_THRESHOLD", 50)
	cfg.MaxFailedAttempts = getEnvInt("IP_MAX_FAILED_ATTEMPTS", 10)
	cfg.BlockDuration = time.Duration(getEnvInt("IP_BLOCK_DURATION_MINUTES", 15)) * time.Minute
	cfg.WindowDuration = time.Duration(getEnvInt("IP_WINDOW_DURATION_MINUTES", 60)) * time.Minute
	cfg.Whitelist = getEnv("IP_WHITELIST", "")

	return cfg
}

// IPIntelligence provides IP-based threat detection.
//
type IPIntelligence struct {
	mu          sync.RWMutex
	config      IPIntelligenceConfig
	failures    map[string]*ipFailureRecord
	blocked     map[string]time.Time
	whitelist   map[string]bool
	stopCleanup chan struct{}
	cleanupWg   sync.WaitGroup
}

type ipFailureRecord struct {
	FirstSeen time.Time
	LastSeen  time.Time
	Count     int
}

// NewIPIntelligence creates a new IP intelligence service.
func NewIPIntelligence(config IPIntelligenceConfig) *IPIntelligence {
	ii := &IPIntelligence{
		config:   config,
		failures: make(map[string]*ipFailureRecord),
		blocked:  make(map[string]time.Time),
	}

	// Parse whitelist.
	ii.whitelist = make(map[string]bool)

	if config.Whitelist != "" {
		for _, ip := range strings.Split(config.Whitelist, ",") {
			ii.whitelist[strings.TrimSpace(ip)] = true
		}
	}

	return ii
}

// GetClientIP extracts the real client IP from the request.
// Handles X-Forwarded-For, X-Real-IP, and direct connection.
func (ii *IPIntelligence) GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (may contain multiple IPs).
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (original client).
		if idx := strings.Index(xff, ","); idx != -1 {
			xff = xff[:idx]
		}

		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header.
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr.
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// IsWhitelisted checks if an IP is on the whitelist.
func (ii *IPIntelligence) IsWhitelisted(ip string) bool {
	ii.mu.RLock()
	defer ii.mu.RUnlock()

	return ii.whitelist[ip]
}

// IsBlocked checks if an IP is currently blocked.
func (ii *IPIntelligence) IsBlocked(ip string) bool {
	ii.mu.RLock()
	defer ii.mu.RUnlock()

	blockedUntil, exists := ii.blocked[ip]
	if !exists {
		return false
	}

	// Check if block has expired.
	if time.Now().After(blockedUntil) {
		return false
	}

	return true
}

// RecordFailedAttempt records a failed authentication attempt from an IP.
func (ii *IPIntelligence) RecordFailedAttempt(ip string) {
	ii.mu.Lock()
	defer ii.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-ii.config.WindowDuration)

	record, exists := ii.failures[ip]
	if !exists {
		ii.failures[ip] = &ipFailureRecord{
			Count:     1,
			FirstSeen: now,
			LastSeen:  now,
		}

		return
	}

	// Reset if outside window.
	if record.FirstSeen.Before(cutoff) {
		ii.failures[ip] = &ipFailureRecord{
			Count:     1,
			FirstSeen: now,
			LastSeen:  now,
		}

		return
	}

	// Increment.
	record.Count++
	record.LastSeen = now

	// Check if we should block.
	if record.Count >= ii.config.MaxFailedAttempts {
		ii.blocked[ip] = now.Add(ii.config.BlockDuration)
	}
}

// ClearFailedAttempts clears failure records for an IP (after successful auth).
func (ii *IPIntelligence) ClearFailedAttempts(ip string) {
	ii.mu.Lock()
	defer ii.mu.Unlock()
	delete(ii.failures, ip)
}

// GetFailureCount returns the current failure count for an IP.
func (ii *IPIntelligence) GetFailureCount(ip string) int {
	ii.mu.RLock()
	defer ii.mu.RUnlock()

	record, exists := ii.failures[ip]
	if !exists {
		return 0
	}

	cutoff := time.Now().Add(-ii.config.WindowDuration)
	if record.FirstSeen.Before(cutoff) {
		return 0
	}

	return record.Count
}

// Cleanup removes expired entries to prevent memory bloat.
func (ii *IPIntelligence) Cleanup() {
	ii.mu.Lock()
	defer ii.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-ii.config.WindowDuration)
	failureCutoff := now.Add(-ii.config.BlockDuration * 2)

	// Clean expired blocks.
	for ip, blockedUntil := range ii.blocked {
		if blockedUntil.Before(failureCutoff) {
			delete(ii.blocked, ip)
		}
	}

	// Clean old failure records.
	for ip, record := range ii.failures {
		if record.FirstSeen.Before(cutoff) {
			delete(ii.failures, ip)
		}
	}
}

// StartCleanupRoutine starts a background goroutine to periodically cleanup.
func (ii *IPIntelligence) StartCleanupRoutine(ctx context.Context, interval time.Duration) {
	ii.stopCleanup = make(chan struct{})
	ii.cleanupWg.Add(1)

	go func() {
		defer ii.cleanupWg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ii.stopCleanup:
				return
			case <-ticker.C:
				ii.Cleanup()
			}
		}
	}()
}

// Stop stops the cleanup goroutine.
func (ii *IPIntelligence) Stop() {
	if ii.stopCleanup != nil {
		close(ii.stopCleanup)
		ii.cleanupWg.Wait()
	}
}

// Middleware returns a Gin middleware that blocks known malicious IPs.
func (ii *IPIntelligence) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ii.config.Enabled {
			c.Next()
			return
		}

		ip := ii.GetClientIP(c.Request)

		// Skip whitelisted IPs.
		if ii.IsWhitelisted(ip) {
			c.Next()
			return
		}

		// Check if blocked.
		if ii.IsBlocked(ip) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "ip_blocked",
				"message": "Too many failed attempts, please try again later",
			})

			return
		}

		c.Next()
	}
}

// RecordAuthFailure is a helper to record authentication failure.
// Call this from login handlers when authentication fails.
func (ii *IPIntelligence) RecordAuthFailure(c *gin.Context) {
	if !ii.config.Enabled {
		return
	}

	ip := ii.GetClientIP(c.Request)
	if ii.IsWhitelisted(ip) {
		return
	}

	ii.RecordFailedAttempt(ip)
}

// RecordAuthSuccess is a helper to record successful authentication.
// Call this from login handlers when authentication succeeds.
func (ii *IPIntelligence) RecordAuthSuccess(c *gin.Context) {
	if !ii.config.Enabled {
		return
	}

	ip := ii.GetClientIP(c.Request)
	ii.ClearFailedAttempts(ip)
}

// CheckAbuseIPDB checks an IP against AbuseIPDB API.
// This is optional and can be called during authentication.
func (ii *IPIntelligence) CheckAbuseIPDB(ctx context.Context, ip string) (bool, int, error) {
	if !ii.config.AbuseIPDBEnabled || ii.config.AbuseIPDBAPIKey == "" {
		return false, 0, nil
	}

	// Skip private IPs.
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil || parsedIP.IsPrivate() || parsedIP.IsLoopback() {
		return false, 0, nil
	}

	// Call AbuseIPDB API.
	url := "https://api.abuseipdb.com/api/v2/check"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, 0, err
	}

	req.Header.Set("Key", ii.config.AbuseIPDBAPIKey)
	req.Header.Set("Accept", "application/json")
	q := req.URL.Query()
	q.Add("ipAddress", ip)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return false, 0, err
	}

	defer func() { _ = resp.Body.Close() }()

	// Parse response.
	type abuseIPDBResponse struct {
		Data struct {
			IPAddress            string `json:"ipAddress"`
			AbuseConfidenceScore int    `json:"abuseConfidenceScore"`
		} `json:"data"`
	}

	var result abuseIPDBResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, 0, err
	}

	blocked := result.Data.AbuseConfidenceScore >= ii.config.AbuseIPDBConfidenceThreshold

	return blocked, result.Data.AbuseConfidenceScore, nil
}
