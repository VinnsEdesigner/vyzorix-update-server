package vyzorix

// UsageStatsSnapshot — generated from openapi/schemas.go.
#UsageStatsSnapshot: {
        collected_at: string
        toggles: {[string]: _}
        counts: #UsageStatsCounts
}

// UsageStatsCounts — generated from openapi/schemas.go.
#UsageStatsCounts: {
        devices: int
        operators: int
        organizations: int
        service_accounts: int
        alert_rules: int
        contact_points: int
        annotations: int
}
