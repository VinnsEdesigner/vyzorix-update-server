package vyzorix

// ClientCredential.
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
        client_id?: string
        client_secret?: string
}

// ClientCredentialListResult.
#ClientCredentialListResult: {
        clients: [...#ClientCredential]
}

// CreateClientCredentialRequest.
#CreateClientCredentialRequest: {
        name: string
}

// UpdateClientCredentialRequest.
#UpdateClientCredentialRequest: {
        name?: string
}

// ClientCredentialUpdateResult.
#ClientCredentialUpdateResult: {
        client?: #ClientCredential
        message?: string
}

// ClientSecretRotateResult.
#ClientSecretRotateResult: {
        client_id?: string
        client_secret?: string
        message?: string
}
