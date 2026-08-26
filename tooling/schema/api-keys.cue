package vyzorix

// CreateAPIKeyRequest — generated from openapi/schemas.go.
#CreateAPIKeyRequest: {
        expires_in_days?: int
        name: string
        scope: string
}

// APIKey — generated from openapi/schemas.go.
#APIKey: {
        created_at: string
        updated_at: string
        expires_at?: string
        last_request_at?: string
        revoked_at?: string
        id: string
        operator_id?: string
        name: string
        key_prefix: string
        scope: string
        request_count: int64
        is_active: bool
}

// APIKeyWithSecret — generated from openapi/schemas.go.
#APIKeyWithSecret: {
        updated_at: string
        created_at: string
        expires_at?: string
        revoked_at?: string
        last_request_at?: string
        key_prefix: string
        scope: string
        id: string
        name: string
        operator_id?: string
        api_key: string
        request_count: int64
        is_active: bool
}

// APIKeyListResult — generated from openapi/schemas.go.
#APIKeyListResult: {
        keys: [...#APIKey]
        pagination: #Pagination
        monthly_limit: int
        keys_created_this_month: int
}

// GlobalAPIKeyStatsResult — generated from openapi/schemas.go.
#GlobalAPIKeyStatsResult: {
        requests_by_scope: {[string]: _}
        top_operators: [...#TopOperatorStat]
        total_keys: int
        active_keys: int
        revoked_keys: int
        total_operators: int
        max_per_month: int
        total_requests: int64
}

// AdminAPIKey — generated from openapi/schemas.go.
#AdminAPIKey: {
        operator_id?: string
        operator_name?: string
}

// AdminAPIKeyListResult — generated from openapi/schemas.go.
#AdminAPIKeyListResult: {
        keys: [...#AdminAPIKey]
        pagination: #Pagination
}
