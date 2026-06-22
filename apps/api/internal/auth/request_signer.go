// Package auth provides authentication utilities for clients.
// Deprecated: Use github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security instead.
package auth

import (
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/request_signer"
)

// Re-export from new location for backward compatibility.
type RequestSigner = request_signer.Signer

func NewRequestSigner(clientID, clientSecret string) (*request_signer.Signer, error) {
return request_signer.New(clientID, clientSecret)
}
