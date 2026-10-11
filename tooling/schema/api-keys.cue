package vyzorix

// CreateAPIKeyRequest.
#CreateAPIKeyRequest: {
        expires_in_days?: int
        name: string
        scope: string
}

// APIKey.
#APIKey: {
        created_at: int64
        updated_at: int64
        expires_at?: int64
        last_request_at?: int64
        revoked_at?: int64
        id: string
        operator_id?: string
        name: string
        key_prefix: string
        scope: string
        request_count: int64
        is_active: bool
}

// APIKeyWithSecret.
#APIKeyWithSecret: {
        updated_at: int64
        created_at: int64
        expires_at?: int64
        revoked_at?: int64
        last_request_at?: int64
        key_prefix: string
        scope: string
        id: string
        name: string
        operator_id?: string
        api_key: string
        request_count: int64
        is_active: bool
}

// APIKeyListResult.
#APIKeyListResult: {
        keys: [...#APIKey]
        pagination: #Pagination
        monthly_limit: int
        keys_created_this_month: int
}

// GlobalAPIKeyStatsResult.
#GlobalAPIKeyStatsResult: {
        requests_by_scope: {[string]: int64}
        top_operators: [...#TopOperatorStat]
        total_keys: int
        active_keys: int
        revoked_keys: int
        total_operators: int
        max_per_month: int
        total_requests: int64
}

// AdminAPIKey.
// Flattened from the embedded APIKeyResponse so swag can introspect it.
#AdminAPIKey: {
        expires_at?: int64
        last_request_at?: int64
        revoked_at?: int64
        operator_id?: string
        operator_name?: string
        id: string
        name: string
        key_prefix: string
        scope: string
        created_at: int64
        updated_at: int64
        request_count: int64
        is_active: bool
}

// AdminAPIKeyListResult.
#AdminAPIKeyListResult: {
        keys: [...#AdminAPIKey]
        pagination: #Pagination
        monthly_limit?: int
        keys_created_this_month?: int
}
