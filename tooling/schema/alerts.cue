package vyzorix

// AlertRuleRequest — generated from openapi/schemas.go.
#AlertRuleRequest: {
        name: string
        metric: string
        condition: string
        webhook_url: string
        on_no_data: string
        on_error: string
        threshold: float64
        for_seconds: int
        notify_interval_seconds: int
        enabled: bool
}

// AlertInstance — generated from openapi/schemas.go.
#AlertInstance: {
        evaluated_at: string
        labels: {[string]: _}
        state: string
        value: float64
}

// AlertRule — generated from openapi/schemas.go.
#AlertRule: {
        created_at: string
        updated_at: string
        webhook_url: string
        metric: string
        condition: string
        id: string
        name: string
        org_id: string
        on_no_data: string
        on_error: string
        instances: [...#AlertInstance]
        threshold: float64
        for_seconds: int
        notify_interval_seconds: int
        enabled: bool
}

// AlertRuleListResult — generated from openapi/schemas.go.
#AlertRuleListResult: {
        rules: [...#AlertRule]
}

// AlertHistoryEvent — generated from openapi/schemas.go.
#AlertHistoryEvent: {
        created_at: string
        id: string
        rule_id: string
        from_state: string
        to_state: string
        value: float64
}

// AlertHistoryResult — generated from openapi/schemas.go.
#AlertHistoryResult: {
        events: [...#AlertHistoryEvent]
}

// AlertEvaluateResult — generated from openapi/schemas.go.
#AlertEvaluateResult: {
        rule_id: string
        transitioned: int
}
