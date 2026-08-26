package vyzorix

// MessageResult is a simple {message: string} response.
#MessageResult: {
	message: string
}

// SuccessResult is a {success: bool, message?: string} response.
#SuccessResult: {
	success: bool
	message?: string
}

// LockoutStatusResult is the lockout status response.
#LockoutStatusResult: {
	locked:       bool
	reason?:      string
	retryAfter?:  int
	attempts?:    int
	unlockAt?:    int
}

// AdminOperator is a single operator in admin responses.
#AdminOperator: {
	id:            string
	email:         string
	name:          string
	role?:         string
	mfaEnabled:    bool
	emailVerified: bool
	createdAt:    int64
	updatedAt:    int64
}

// AdminOperatorListResult wraps a list of operators.
#AdminOperatorListResult: {
	operators: [...#AdminOperator]
	total:     int
}

// APIKey is a single API key in list/get responses.
#APIKey: {
	id:           string
	operatorID?:  string
	name:         string
	keyPrefix:    string
	scope:        string
	expiresAt?:   string
	isActive:     bool
	requestCount: int
	lastRequestAt?: string
	createdAt:    int64
	updatedAt:    string
	revokedAt?:   string
}

// APIKeyWithSecret is the create/rotate response with the full key.
#APIKeyWithSecret: {
	...#APIKey
	apiKey: string
}

// APIKeyListResult wraps paginated API keys.
#APIKeyListResult: {
	keys:                 [...#APIKey]
	pagination:           #Pagination
	monthlyLimit:         int
	keysCreatedThisMonth: int
}

// Pagination is the shared paging envelope.
#Pagination: {
	page:       int
	limit:      int
	total:      int
	totalPages: int
}

// CommandDispatchResult is the POST /commands/{imei}/execute response.
#CommandDispatchResult: {
	dispatchID:   string
	commandID?:   string
	deviceID?:    string
	status?:      string
	deviceOnline?: bool
	serverTime?:  int
}

// ServiceAccount is a service account entity.
#ServiceAccount: {
	id:        string
	orgID:     string
	name:      string
	enabled:   bool
	createdAt: string
	updatedAt: int64
}

// ServiceAccountListResult wraps a list of service accounts.
#ServiceAccountListResult: {
	serviceAccounts: [...#ServiceAccount]
}

// ServiceAccountToken is a token for a service account.
#ServiceAccountToken: {
	id:        string
	serviceID: string
	name:      string
	keyPrefix: string
	scopes:    [...string]
	valid:     bool
	expiresAt?: string
	createdAt: string
	revokedAt?: string
}

// ServiceAccountTokenListResult wraps a list of tokens.
#ServiceAccountTokenListResult: {
	tokens: [...#ServiceAccountToken]
}
