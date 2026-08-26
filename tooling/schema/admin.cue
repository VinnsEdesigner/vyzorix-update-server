package vyzorix

// TopOperatorStat — generated from openapi/schemas.go.
#TopOperatorStat: {
        operator_id: string
        operator_name: string
        total_requests: int64
        active_key_count: int
}

// OperatorAPIKeyStatsResult — generated from openapi/schemas.go.
#OperatorAPIKeyStatsResult: {
        operator_id: string
        total_keys: int64
        active_keys: int64
        revoked_keys: int64
        keys_created_this_month: int
        monthly_limit: int
}

// AdminOperator — generated from openapi/schemas.go.
#AdminOperator: {
        id: string
        email: string
        name: string
        role?: string
        created_at: int64
        updated_at: int64
        mfa_enabled: bool
        email_verified: bool
}

// AdminOperatorListResult — generated from openapi/schemas.go.
#AdminOperatorListResult: {
        operators: [...#AdminOperator]
        total: int
}

// CreateOperatorRequest — generated from openapi/schemas.go.
#CreateOperatorRequest: {
        email: string
        password: string
        name: string
        role?: string
}

// OperatorThresholds — generated from openapi/schemas.go.
#OperatorThresholds: {
        riskWarn?: int
        riskCrit?: int
        thermalWarn?: int
        thermalCrit?: int
        bufferWarn?: int
        bufferCrit?: int
}

// OperatorSettingsResult — generated from openapi/schemas.go.
#OperatorSettingsResult: {
        notifications?: #NotificationSettings
        client?: #ClientSettings
        thresholds?: #OperatorThresholds
}

// OperatorSettingsResultLegacy — generated from openapi/schemas.go.
#OperatorSettingsResultLegacy: {
        client?: #ClientSettings
        id: string
        email: string
        name: string
        role?: string
        mfa_enabled: bool
        email_verified: bool
}
