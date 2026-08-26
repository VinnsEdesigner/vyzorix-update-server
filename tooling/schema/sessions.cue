package vyzorix

// SessionInfo — generated from openapi/schemas.go.
#SessionInfo: {
        id: string
        ip_address: string
        user_agent: string
        created_at: string
        expires_at: string
        selected_organization_id?: string
        is_current: bool
}

// SessionListResult — generated from openapi/schemas.go.
#SessionListResult: {
        sessions: [...#SessionInfo]
        total: int
}

// ConcurrentSessionsResult — generated from openapi/schemas.go.
#ConcurrentSessionsResult: {
        sessions: [...#SessionInfo]
        count: int
        has_concurrent: bool
}
