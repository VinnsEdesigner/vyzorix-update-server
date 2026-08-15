package oauth

// GoogleOAuthURLRequest asks the server for the Google OAuth authorization URL.
type GoogleOAuthURLRequest struct{}

// GoogleOAuthURLResponse returns the URL to redirect the browser to.
type GoogleOAuthURLResponse struct {
	URL string `json:"url"`
}

// GoogleOAuthCallbackRequest is the internal payload after Google validates the callback.
type GoogleOAuthCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// OAuthState represents the state stored during OAuth flow.
type OAuthState struct {
	RedirectURL string `json:"redirectUrl"`
	Nonce       string `json:"nonce"`
}

// OAuthProvider represents a supported OAuth provider.
type OAuthProvider string

const (
	ProviderGoogle OAuthProvider = "google"
	ProviderGitHub OAuthProvider = "github"
)

// OAuthLinkRequest is the payload for linking an OAuth provider to an existing account.
type OAuthLinkRequest struct {
	Provider   OAuthProvider `json:"provider"`
	OAuthToken string        `json:"oauthToken"`
}
