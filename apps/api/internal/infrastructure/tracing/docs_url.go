// Package tracing provides request tracing capabilities.
package tracing

import (
	"os"
	"strings"
	"sync"
)

// DefaultDocsBaseURL is used when no base URL is configured via SetDocsBaseURL
// or the VYZORIX_DOCS_BASE_URL environment variable. It carries no trailing slash.
const DefaultDocsBaseURL = "https://docs.vyzorix.com/errors"

// DocsURLBuilder builds error documentation URLs dynamically.
type DocsURLBuilder struct {
	baseURL string
	mu      sync.RWMutex
}

// DefaultDocsURLBuilder is the global docs URL builder, seeded from the
// VYZORIX_DOCS_BASE_URL environment variable if present.
var DefaultDocsURLBuilder = func() *DocsURLBuilder {
	b := &DocsURLBuilder{}
	if env := strings.TrimRight(os.Getenv("VYZORIX_DOCS_BASE_URL"), "/"); env != "" {
		b.baseURL = env
	} else {
		b.baseURL = DefaultDocsBaseURL
	}
	return b
}()

// SetBaseURL sets the base URL for documentation links.
func (b *DocsURLBuilder) SetBaseURL(url string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.baseURL = url
}

// GetBaseURL returns the current base URL.
func (b *DocsURLBuilder) GetBaseURL() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.baseURL
}

// BuildURL builds a documentation URL for an error code.
func (b *DocsURLBuilder) BuildURL(errorCode string) string {
	b.mu.RLock()
	base := b.baseURL
	b.mu.RUnlock()

	if base == "" {
		base = DefaultDocsBaseURL
	}
	return base + "/" + errorCode
}

// Global helpers using DefaultDocsURLBuilder.

// SetDocsBaseURL sets the global docs base URL.
func SetDocsBaseURL(url string) {
	DefaultDocsURLBuilder.SetBaseURL(url)
}

// BuildErrorDocsURL builds a docs URL for an error code.
func BuildErrorDocsURL(errorCode string) string {
	return DefaultDocsURLBuilder.BuildURL(errorCode)
}
