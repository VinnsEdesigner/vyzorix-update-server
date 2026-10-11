package vyzorix

// TopOperatorStat.
#TopOperatorStat: {
        operator_id: string
        operator_name: string
        total_requests: int64
        active_key_count: int
}

// OperatorAPIKeyStatsResult.
#OperatorAPIKeyStatsResult: {
        operator_id: string
        total_keys: int64
        active_keys: int64
        revoked_keys: int64
        keys_created_this_month: int
        monthly_limit: int
}

// AdminOperator.
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

// AdminOperatorListResult.
#AdminOperatorListResult: {
        operators: [...#AdminOperator]
        total: int
}

// CreateOperatorRequest.
#CreateOperatorRequest: {
        email: string
        password: string
        name: string
        role?: string
}

// OperatorThresholds.
#OperatorThresholds: {
        riskWarn?: int
        riskCrit?: int
        thermalWarn?: int
        thermalCrit?: int
        bufferWarn?: int
        bufferCrit?: int
}

// OperatorSettingsResult.
#OperatorSettingsResult: {
        notifications?: #NotificationSettings
        client?: #ClientSettings
        thresholds?: #OperatorThresholds
}

// OperatorSettingsResultLegacy.
#OperatorSettingsResultLegacy: {
        client?: #ClientSettings
        id: string
        email: string
        name: string
        role?: string
        mfa_enabled: bool
        email_verified: bool
}

// AdminClientActionResult.
#AdminClientActionResult: {
        success: bool
        client_id?: string
}
