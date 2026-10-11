package vyzorix

// SessionInfo.
#SessionInfo: {
        id: string
        ip_address: string
        user_agent: string
        created_at: int64
        expires_at: int64
        selected_organization_id?: string
        is_current: bool
}

// SessionListResult.
#SessionListResult: {
        sessions: [...#SessionInfo]
        total: int
}

// ConcurrentSessionsResult.
#ConcurrentSessionsResult: {
        sessions: [...#SessionInfo]
        count: int
        has_concurrent: bool
}
