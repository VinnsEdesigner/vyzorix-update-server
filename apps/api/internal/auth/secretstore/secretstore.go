// Package secretstore provides encrypted secret storage for devices.
// Deprecated: Use github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security instead.
package secretstore

import (
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/secretstore"
)

// Re-export from new location for backward compatibility.
type SecretStore = secretstore.Store

var ErrNotFound = secretstore.ErrNotFound

func NewSecretStore(baseDir, masterKeyBase64 string) (*secretstore.Store, error) {
return secretstore.New(baseDir, masterKeyBase64)
}
