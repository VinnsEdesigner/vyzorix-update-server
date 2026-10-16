// Package middleware provides HTTP middleware.
package middleware

import (
	"crypto/subtle"
	"net"
	"net/http"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/gin-gonic/gin"
)

// MetricsAccessConfig controls who may scrape /metrics.
type MetricsAccessConfig struct {
	// Token, when non-empty, must be presented as `Authorization: Bearer <token>`.
	Token string
	// AllowedCIDRs, when non-empty, restricts scrapers to these networks.
	AllowedCIDRs []string
	// AllowLoopback always permits 127.0.0.1/::1 (a local scraper needs no token).
	AllowLoopback bool
	// AllowPrivate permits RFC1918/link-local peers when no CIDRs are configured,
	// for deployments scraped from inside a private network.
	AllowPrivate bool
}

// MetricsAccess returns middleware enforcing the metrics access policy.
//
// It fails closed: a request that satisfies none of the configured rules is
// rejected with 403 rather than served. Authentication failures return 401 when a
// token was presented but was wrong, so a misconfigured scraper is diagnosable.
func MetricsAccess(cfg MetricsAccessConfig) gin.HandlerFunc {
	prefixes := parseCIDRs(cfg.AllowedCIDRs)

	return func(c *gin.Context) {
		// Transport peer, not c.ClientIP(): ClientIP honours X-Forwarded-For, which
		// any caller can set, and an allowlist must not be steerable by a header.
		peer := peerIP(c.Request.RemoteAddr)

		// Loopback scrapers are always allowed - a same-host agent needs no secret.
		if cfg.AllowLoopback && peer != nil && peer.IsLoopback() {
			c.Next()
			return
		}

		// Network allowlist, when configured, is required in addition to any token.
		cidrOK := len(prefixes) == 0 || ipInPrefixes(peer, prefixes)
		if !cidrOK {
			denyMetrics(c, http.StatusForbidden, "scrape source not in metrics allowlist")
			return
		}

		// Scrape token, when configured, is mandatory for non-loopback scrapers.
		if cfg.Token != "" {
			presented := bearerToken(c.GetHeader("Authorization"))
			if presented == "" {
				denyMetrics(c, http.StatusUnauthorized, "metrics requires a scrape token")
				return
			}
			if subtle.ConstantTimeCompare([]byte(presented), []byte(cfg.Token)) != 1 {
				denyMetrics(c, http.StatusUnauthorized, "invalid metrics scrape token")
				return
			}
			c.Next()
			return
		}

		// No token configured: the network allowlist (or the private-network
		// allowance) is the control, and it already passed above.
		if len(prefixes) > 0 || (cfg.AllowPrivate && peer != nil && (peer.IsPrivate() || peer.IsLinkLocalUnicast())) {
			c.Next()
			return
		}

		// No token, no CIDR, not private: deny.
		denyMetrics(c, http.StatusForbidden, "metrics is not reachable from this address")
	}
}

// denyMetrics aborts with a structured error and no metric data.
func denyMetrics(c *gin.Context, status int, msg string) {
	c.Header("WWW-Authenticate", `Bearer realm="metrics"`)
	responses.RespondStructuredAbort(c, status, msg)
}

// bearerToken extracts the token from an Authorization header, tolerating the
// scheme case-insensitively (RFC 7235).
func bearerToken(header string) string {
	const scheme = "bearer "
	if len(header) < len(scheme) || !strings.EqualFold(header[:len(scheme)], scheme) {
		return ""
	}
	return strings.TrimSpace(header[len(scheme):])
}

// parseCIDRs parses the configured prefixes, ignoring unparseable entries.
func parseCIDRs(raw []string) []*net.IPNet {
	out := make([]*net.IPNet, 0, len(raw))
	for _, entry := range raw {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if _, ipnet, err := net.ParseCIDR(entry); err == nil {
			out = append(out, ipnet)
			continue
		}
		// A bare address is treated as a single-host prefix.
		if ip := net.ParseIP(entry); ip != nil {
			bits := 32
			if ip.To4() == nil {
				bits = 128
			}
			out = append(out, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
		}
	}
	return out
}

// peerIP parses the transport peer address ("host:port") from RemoteAddr. This is
// the actual TCP peer and cannot be influenced by request headers.
func peerIP(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// No port (unusual); try the value as-is.
		host = remoteAddr
	}
	return net.ParseIP(strings.TrimSpace(host))
}

func ipInPrefixes(ip net.IP, prefixes []*net.IPNet) bool {
	if ip == nil {
		return false
	}
	for _, p := range prefixes {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}
