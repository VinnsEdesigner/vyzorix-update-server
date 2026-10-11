package vyzorix

// AlertRuleRequest.
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

// AlertInstance.
#AlertInstance: {
        evaluated_at: int64
        labels: {[string]: string}
        state: string
        value: float64
}

// AlertRule.
#AlertRule: {
        created_at: int64
        updated_at: int64
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

// AlertRuleListResult.
#AlertRuleListResult: {
        rules: [...#AlertRule]
}

// AlertHistoryEvent.
#AlertHistoryEvent: {
        created_at: int64
        id: string
        rule_id: string
        from_state: string
        to_state: string
        value: float64
}

// AlertHistoryResult.
#AlertHistoryResult: {
        events: [...#AlertHistoryEvent]
}

// AlertEvaluateResult.
#AlertEvaluateResult: {
        rule_id: string
        transitioned: bool
}
