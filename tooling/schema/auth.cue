package vyzorix

// LoginRequest.
#LoginRequest: {
        email: string
        password: string
        device_fingerprint?: string
        remember: bool
}

// LoginResult.
#LoginResult: {
        selected_organization?: #OrganizationInfo
        operator_id: string
        email: string
        name: string
        last_organization_id?: string
        signing_key: string
        mfa_session?: string
        device_fingerprint?: string
        organizations?: [...#OrganizationInfo]
        mfa_enabled: bool
        needs_organization: bool
        requires_mfa?: bool
        error?: string
        message?: string
        retry_after?: int64
        locked_until?: int64
}

// LoginWithTokensResult.
#LoginWithTokensResult: {
        selected_organization?: #OrganizationInfo
        mfa_session?: string
        email: string
        name: string
        last_organization_id?: string
        access_token: string
        refresh_token: string
        session_id: string
        signing_key: string
        operator_id: string
        device_fingerprint?: string
        organizations?: [...#OrganizationInfo]
        expires_at: int64
        needs_organization: bool
        requires_mfa?: bool
        mfa_enabled: bool
        error?: string
        message?: string
        retry_after?: int64
        locked_until?: int64
}

// RegisterRequest.
#RegisterRequest: {
        email: string
        password: string
        name: string
        role?: string
}

// RegisterResult.
#RegisterResult: {
        operator_id: string
        email: string
        name: string
}

// RefreshTokenRequest.
#RefreshTokenRequest: {
        refresh_token: string
}

// RefreshTokenResult.
#RefreshTokenResult: {
        access_token: string
        refresh_token: string
        session_id: string
        expires_at: int64
}
