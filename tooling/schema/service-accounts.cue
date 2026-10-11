package vyzorix

// CreateServiceAccountRequest.
#CreateServiceAccountRequest: {
        name: string
}

// CreateServiceAccountTokenRequest.
#CreateServiceAccountTokenRequest: {
        expires_at?: string
        service_id: string
        name: string
        scopes: [...string]
}

// RotateServiceAccountTokenRequest.
#RotateServiceAccountTokenRequest: {
        expires_at?: string
        name: string
        scopes: [...string]
}

// ServiceAccount.
#ServiceAccount: {
        created_at: int64
        updated_at: int64
        id: string
        org_id: string
        name: string
        enabled: bool
}

// ServiceAccountToken.
#ServiceAccountToken: {
        created_at: int64
        expires_at?: int64
        revoked_at?: int64
        id: string
        service_id: string
        name: string
        key_prefix: string
        scopes: [...string]
        valid: bool
}

// ServiceAccountListResult.
#ServiceAccountListResult: {
        service_accounts: [...#ServiceAccount]
}

// ServiceAccountTokenListResult.
#ServiceAccountTokenListResult: {
        tokens: [...#ServiceAccountToken]
}

// ServiceAccountTokenCreated.
#ServiceAccountTokenCreated: {
        secret: string
}

// ServiceAccountTokenRotated.
#ServiceAccountTokenRotated: {
        secret: string
}
