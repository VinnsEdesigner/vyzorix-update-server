package vyzorix

// MetricStatsDTO.
#MetricStatsDTO: {
        current: float64
        avg: float64
        min: float64
        max: float64
}

// MessageResult.
#MessageResult: {
        message: string
}

// MeResult.
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

// TimelineEventResult — matches the runtime diagnostics EventDTO wire shape.
#TimelineEventResult: {
        data?: {[string]: _}
        id: string
        type: string
        timestamp: int64
}

// TimelineResult — matches the runtime diagnostics TimelineResponse wire shape
// ({ events, pagination }), NOT the previous { nextCursor, events, hasMore }.
#TimelineResult: {
        events: [...#TimelineEventResult]
        pagination: #CursorPaginationResult
}

// MetricAggregateResult.
#MetricAggregateResult: {
        avg: float64
        min: float64
        max: float64
}
