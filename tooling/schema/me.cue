package vyzorix

// MetricStatsDTO — generated from openapi/schemas.go.
#MetricStatsDTO: {
        current: float64
        avg: float64
        min: float64
        max: float64
}

// MessageResult — generated from openapi/schemas.go.
#MessageResult: {
        message: string
}

// MeResult — generated from openapi/schemas.go.
#MeResult: {
        selected_organization?: #OrganizationInfo
        id: string
        email: string
        name: string
        last_organization_id?: string
        organizations: [...#OrganizationInfo]
        mfa_enabled: bool
        email_verified: bool
        needs_organization: bool
}

// TimelineEventResult — generated from openapi/schemas.go.
#TimelineEventResult: {
        data?: {[string]: _}
        id: string
        deviceId: string
        type: string
        timestamp: string
}

// TimelineResult — generated from openapi/schemas.go.
#TimelineResult: {
        nextCursor?: string
        events: [...#TimelineEventResult]
        hasMore: bool
}

// MetricAggregateResult — generated from openapi/schemas.go.
#MetricAggregateResult: {
        avg: float64
        min: float64
        max: float64
}
