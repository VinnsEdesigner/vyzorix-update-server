package vyzorix

// LoginRequest — generated from openapi/schemas.go.
#LoginRequest: {
        email: string
        password: string
        device_fingerprint?: string
        remember: bool
}

// LoginResult — generated from openapi/schemas.go.
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
}

// LoginWithTokensResult — generated from openapi/schemas.go.
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
}

// RegisterRequest — generated from openapi/schemas.go.
#RegisterRequest: {
        email: string
        password: string
        name: string
        role?: string
}

// RegisterResult — generated from openapi/schemas.go.
#RegisterResult: {
        operator_id: string
        email: string
        name: string
}

// RefreshTokenRequest — generated from openapi/schemas.go.
#RefreshTokenRequest: {
        refresh_token: string
}

// RefreshTokenResult — generated from openapi/schemas.go.
#RefreshTokenResult: {
        access_token: string
        refresh_token: string
        session_id: string
        expires_at: int64
}
