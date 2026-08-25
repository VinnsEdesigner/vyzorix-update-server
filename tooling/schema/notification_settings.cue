package vyzorix

// EmailNotifications controls which events trigger email alerts.
#EmailNotifications: {
        thresholdBreach:     bool | *true
        deviceOffline:       bool | *true
        deviceOnline:        bool | *false
        updateAvailable:     bool | *false
        commandFailed:       bool | *true
        registrationRequest: bool | *false
}

// PushNotifications mirrors EmailNotifications for push channels.
#PushNotifications: {
        thresholdBreach:     bool | *true
        deviceOffline:       bool | *true
        deviceOnline:        bool | *false
        updateAvailable:     bool | *false
        commandFailed:       bool | *false
        registrationRequest: bool | *false
}

// WebhookNotifications configures a webhook delivery endpoint.
#WebhookNotifications: {
        url:     string
        secret?: string
        types:   [...string] | *[]
        enabled: bool | *false
}

// NotificationSettings is the full notification configuration for an operator.
#NotificationSettings: {
        channels: [...string] | *["email"]
        webhook:  #WebhookNotifications
        email:    #EmailNotifications
        push:     #PushNotifications
        enabled:  bool | *true
}
