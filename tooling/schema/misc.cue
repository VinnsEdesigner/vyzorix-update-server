package vyzorix

// ErrorResponse — generated from openapi/schemas.go.
#ErrorResponse: {
        fields?: {[string]: _}
        error: string
        message: string
        trace_id?: string
        docs?: string
}

// Pagination — generated from openapi/schemas.go.
#Pagination: {
        page: int
        limit: int
        total: int64
        total_pages: int
}

// DeletedResult — generated from openapi/schemas.go.
#DeletedResult: {
        deleted: bool
}

// RevokedResult — generated from openapi/schemas.go.
#RevokedResult: {
        revoked: bool
}

// AdminClient — generated from openapi/schemas.go.
#AdminClient: {
        id: string
        last_request_at?: int64
        name: string
        operator_id: string
        platform: string
        allowed_origins: [...string]
        allowed_paths: [...string]
        created_at: int64
        rate_limit: int
        request_count: int64
        updated_at: int64
        is_active: bool
}

// AdminClientListResult — generated from openapi/schemas.go.
#AdminClientListResult: {
        clients: [...#AdminClient]
        total: int
}

// AdminClientResult — generated from openapi/schemas.go.
#AdminClientResult: {
        client: #AdminClient
}

// SupportBundleResult — generated from openapi/schemas.go.
#SupportBundleResult: {
        generated_at: string
        hostname: string
        go_version: string
        goroutines: int
        go_max_procs: int
        go_num_cpu: int
        schema_version?: int
        device_count?: int
        org_count?: int
        operator_count?: int
}

// LogoutRequest — generated from openapi/schemas.go.
#LogoutRequest: {
        all_devices: bool
}

// SuccessResult — generated from openapi/schemas.go.
#SuccessResult: {
        message?: string
        success: bool
}

// LockoutStatusResult — generated from openapi/schemas.go.
#LockoutStatusResult: {
        unlock_at?: int64
        reason?: string
        attempts: int
        max_attempts?: int
        locked: bool
}

// RevokeResult — generated from openapi/schemas.go.
#RevokeResult: {
        message?: string
        revoked_count?: int
        success: bool
}

// PushNotifications — generated from openapi/schemas.go.
#PushNotifications: {
        thresholdBreach: bool
        deviceOffline: bool
        deviceOnline: bool
        updateAvailable: bool
        commandFailed: bool
        registrationRequest: bool
}

// WebhookNotifications — generated from openapi/schemas.go.
#WebhookNotifications: {
        url: string
        secret?: string
        types: [...string]
        enabled: bool
}

// NotificationSettings — generated from openapi/schemas.go.
#NotificationSettings: {
        channels: [...string]
        webhook: #WebhookNotifications
        email: #EmailNotifications
        push: #PushNotifications
        enabled: bool
}

// ClientSettings — generated from openapi/schemas.go.
#ClientSettings: {
        serverUrl: string
        deviceId: string
        requestTimeoutMs: int
        logBufferLimit: int
        signalHistoryLimit: int
        autoReconnect: bool
        strictHmac: bool
        notificationsEnabled: bool
}

// SettingsResponseResult — generated from openapi/schemas.go.
#SettingsResponseResult: {
        notifications?: #NotificationSettings
        client?: #ClientSettings
        preferences?: {[string]: _}
}

// ThresholdsResult — generated from openapi/schemas.go.
#ThresholdsResult: {
        thresholds?: #OperatorThresholds
}

// PreferencesResult — generated from openapi/schemas.go.
#PreferencesResult: {
        preferences: {[string]: _}
}

// WebhookTestResult — generated from openapi/schemas.go.
#WebhookTestResult: {
        error?: string
        message?: string
        statusCode?: int
        responseTime?: int64
        success: bool
}

// WebhookSecretResult — generated from openapi/schemas.go.
#WebhookSecretResult: {
        secret: string
}

// DownloadProgressRequest — generated from openapi/schemas.go.
#DownloadProgressRequest: {
        device_id: string
        version: string
        progress: int
}

// DownloadProgressResult — generated from openapi/schemas.go.
#DownloadProgressResult: {
        recorded: bool
}

// WebhookTestRequest — generated from openapi/schemas.go.
#WebhookTestRequest: {
        url: string
}

// CursorPaginationResult — generated from openapi/schemas.go.
#CursorPaginationResult: {
        nextCursor?: string
        limit: int
        hasMore: bool
}
