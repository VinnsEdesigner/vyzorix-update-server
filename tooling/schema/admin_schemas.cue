package vyzorix

// CreateAPIKeyRequest is the body for POST /v1/auth/api-keys.
#CreateAPIKeyRequest: {
        name:            string
        scope:           string
        expires_in_days?: int
}

// UpdateAPIKeyRequest is the body for PATCH /v1/auth/api-keys/:keyId.
#UpdateAPIKeyRequest: {
        name?:  string
        scope?: string
}

// AdminClient is a registered OAuth client managed by super_admins.
// Mirrors application/dto.ClientResponse (the wire shape handlers emit).
#AdminClient: {
        id:              string
        operator_id:     string
        name:            string
        platform:        string
        allowed_origins: [...string]
        allowed_paths:   [...string]
        rate_limit:      int
        request_count:   int64
        created_at:      int64
        updated_at:      int64
        is_active:       bool
        last_request_at?: int64
}

// AdminClientListResult wraps a list of admin clients.
#AdminClientListResult: {
        clients: [...#AdminClient]
        total:   int
}

// AdminClientResult wraps a single admin client (GET/PATCH/rotate-key responses).
#AdminClientResult: {
        client: #AdminClient
}

// UpdateAdminClientRequest is the body for PATCH /v1/admin/clients/:clientId.
#UpdateAdminClientRequest: {
        name?:            string
        rate_limit?:      int
        is_active?:       bool
        allowed_origins?: [...string]
        allowed_paths?:   [...string]
}
