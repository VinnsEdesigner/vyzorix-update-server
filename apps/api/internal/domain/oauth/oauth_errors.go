package oauth

import "errors"

var (
	ErrOAuthNotSupported = errors.New("oauth provider not supported")
	ErrOAuthFailed       = errors.New("oauth authentication failed")
	ErrOAuthStateInvalid = errors.New("invalid oauth state")
	ErrOAuthStateExpired = errors.New("oauth state expired")
)
