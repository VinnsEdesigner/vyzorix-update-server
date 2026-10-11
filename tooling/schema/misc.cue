package vyzorix

// ErrorResponse.
#ErrorResponse: {
        fields?: {[string]: string}
        error: string
        message: string
        trace_id?: string
        docs?: string
}

// Pagination.
#Pagination: {
        page: int
        limit: int
        total: int64
        total_pages: int
}

// DeletedResult.
#DeletedResult: {
        deleted: bool
}

// RevokedResult.
#RevokedResult: {
        revoked: bool
}

// AdminClient.
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

// AdminClientListResult.
#AdminClientListResult: {
        clients: [...#AdminClient]
        total: int
}

// AdminClientResult.
#AdminClientResult: {
        client: #AdminClient
}

// SupportBundleResult.
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

// LogoutRequest.
#LogoutRequest: {
        all_devices: bool
}

// SuccessResult.
#SuccessResult: {
        message?: string
        success: bool
}

// LockoutStatusResult.
#LockoutStatusResult: {
        unlock_at?: int64
        reason?: string
        attempts: int
        max_attempts?: int
        locked: bool
        retry_after?: int64
}

// RevokeResult.
#RevokeResult: {
        message?: string
        revoked_count?: int
        success: bool
}

// PushNotifications.
#PushNotifications: {
        thresholdBreach: bool
        deviceOffline: bool
        deviceOnline: bool
        updateAvailable: bool
        commandFailed: bool
        registrationRequest: bool
}

// WebhookNotifications.
#WebhookNotifications: {
        url: string
        secret?: string
        types: [...string]
        enabled: bool
}

// NotificationSettings.
#NotificationSettings: {
        channels: [...string]
        webhook: #WebhookNotifications
        email: #EmailNotifications
        push: #PushNotifications
        enabled: bool
}

// ClientSettings.
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

// SettingsResponseResult.
#SettingsResponseResult: {
        notifications?: #NotificationSettings
        client?: #ClientSettings
        preferences?: {[string]: _}
}

// ThresholdsResult.
#ThresholdsResult: {
        thresholds?: #OperatorThresholds
}

// PreferencesResult.
#PreferencesResult: {
        preferences: {[string]: _}
}

// WebhookTestResult.
#WebhookTestResult: {
        error?: string
        message?: string
        statusCode?: int
        responseTime?: int64
        success: bool
}

// WebhookSecretResult.
#WebhookSecretResult: {
        secret: string
}

// DownloadProgressRequest.
#DownloadProgressRequest: {
        device_id: string
        version: string
        progress: int
}

// DownloadProgressResult.
#DownloadProgressResult: {
        recorded: bool
}

// WebhookTestRequest.
#WebhookTestRequest: {
        url: string
}

// CursorPaginationResult.
#CursorPaginationResult: {
        nextCursor?: string
        limit: int
        hasMore: bool
}
