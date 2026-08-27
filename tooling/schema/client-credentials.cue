package vyzorix

// ClientCredential — generated from openapi/schemas.go.
#ClientCredential: {
        id:               string
        operator_id:      string
        name:             string
        platform:         string
        allowed_origins:  [...string]
        allowed_paths:    [...string]
        rate_limit:       int
        request_count:    int64
        created_at:       int64
        updated_at:       int64
        is_active:        bool
        last_request_at?: int64
        secret?:          string
}

// ClientCredentialListResult — generated from openapi/schemas.go.
#ClientCredentialListResult: {
        clients: [...#ClientCredential]
}

// CreateClientCredentialRequest — generated from openapi/schemas.go.
#CreateClientCredentialRequest: {
        name: string
}

// UpdateClientCredentialRequest — generated from openapi/schemas.go.
#UpdateClientCredentialRequest: {
        name?: string
}
