package vyzorix

// UsageStatsSnapshot.
#UsageStatsSnapshot: {
        collected_at: int64
        toggles: {[string]: bool}
        counts: #UsageStatsCounts
}

// UsageStatsCounts.
#UsageStatsCounts: {
        devices: int
        operators: int
        organizations: int
        service_accounts: int
        alert_rules: int
        contact_points: int
        annotations: int
}
