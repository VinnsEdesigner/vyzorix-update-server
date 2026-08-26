package vyzorix

// CreateServiceAccountRequest — generated from openapi/schemas.go.
#CreateServiceAccountRequest: {
        name: string
}

// CreateServiceAccountTokenRequest — generated from openapi/schemas.go.
#CreateServiceAccountTokenRequest: {
        expires_at?: string
        service_id: string
        name: string
        scopes: [...string]
}

// RotateServiceAccountTokenRequest — generated from openapi/schemas.go.
#RotateServiceAccountTokenRequest: {
        expires_at?: string
        name: string
        scopes: [...string]
}

// ServiceAccount — generated from openapi/schemas.go.
#ServiceAccount: {
        created_at: string
        updated_at: string
        id: string
        org_id: string
        name: string
        enabled: bool
}

// ServiceAccountToken — generated from openapi/schemas.go.
#ServiceAccountToken: {
        created_at: string
        expires_at?: string
        revoked_at?: string
        id: string
        service_id: string
        name: string
        key_prefix: string
        scopes: [...string]
        valid: bool
}

// ServiceAccountListResult — generated from openapi/schemas.go.
#ServiceAccountListResult: {
        service_accounts: [...#ServiceAccount]
}

// ServiceAccountTokenListResult — generated from openapi/schemas.go.
#ServiceAccountTokenListResult: {
        tokens: [...#ServiceAccountToken]
}

// ServiceAccountTokenCreated — generated from openapi/schemas.go.
#ServiceAccountTokenCreated: {
        secret: string
}

// ServiceAccountTokenRotated — generated from openapi/schemas.go.
#ServiceAccountTokenRotated: {
        secret: string
}
