package vyzorix

// UpdateAPIKeyRequest — generated from openapi/schemas.go.
#UpdateAPIKeyRequest: {
        name?: string
        scope?: string
}

// UpdateAdminClientRequest — generated from openapi/schemas.go.
#UpdateAdminClientRequest: {
        name?: string
        rate_limit?: int
        is_active?: bool
        allowed_origins?: [...string]
        allowed_paths?: [...string]
}

// UpdateOperatorRequest — generated from openapi/schemas.go.
#UpdateOperatorRequest: {
        email?: string
        name?: string
        role?: string
        mfa_enabled?: bool
        email_verified?: bool
}

// UpdateSettingsRequest — generated from openapi/schemas.go.
#UpdateSettingsRequest: {
        client?: #ClientSettings
        name?: string
        reset?: bool
}

// ThresholdUpdateRequest — generated from openapi/schemas.go.
#ThresholdUpdateRequest: {
        riskWarn?: int
        riskCrit?: int
        thermalWarn?: int
        thermalCrit?: int
        bufferWarn?: int
        bufferCrit?: int
}

// NotificationUpdateRequest — generated from openapi/schemas.go.
#NotificationUpdateRequest: {
        enabled?: bool
        channels?: ('[...string]', True)
        email?: #EmailNotifications
        push?: #PushNotifications
        webhook?: #WebhookNotifications
}

// UpdateNameRequest — generated from openapi/schemas.go.
#UpdateNameRequest: {
        name: string
}

// UpdateInboxEntryRequest — generated from openapi/schemas.go.
#UpdateInboxEntryRequest: {
        notes?: string
}

// UpdateVersionManifest — generated from openapi/schemas.go.
#UpdateVersionManifest: {
        version: string
        apk_filename: string
        sha256: string
        release_type: string
        release_notes: string
        apk_size: int64
        released_at: int64
        is_latest: bool
}

// UpdateChangelogEntry — generated from openapi/schemas.go.
#UpdateChangelogEntry: {
        version: string
        date: string
        type: string
        notes: string
}

// UpdateCheckRequest — generated from openapi/schemas.go.
#UpdateCheckRequest: {
        current_version: string
}

// UpdateCheckResult — generated from openapi/schemas.go.
#UpdateCheckResult: {
        latest_version?: string
        current_version: string
        release_notes?: string
        download_url?: string
        sha256?: string
        apk_size?: int64
        update_available: bool
}

// UpdaterVersionManifestResult — generated from openapi/schemas.go.
#UpdaterVersionManifestResult: {
        version: string
        apk_filename: string
        apk_sha256: string
        release_notes: string
        version_code: int
        apk_size_bytes: int64
}

// UpdaterCheckResult — generated from openapi/schemas.go.
#UpdaterCheckResult: {
        version: string
        apk_filename: string
        apk_sha256: string
        release_notes: string
        version_code: int
        apk_size_bytes: int64
        update_available: bool
}

// UpdateVersionResponse — generated from openapi/schemas.go.
#UpdateVersionResponse: {
        version: string
        releaseType: string
        status: string
        apkFilename: string
        sha256: string
        releaseNotes?: string
        apkSize: int64
        releasedAt: int64
        isLatest: bool
}

// UpdateVersionListResult — generated from openapi/schemas.go.
#UpdateVersionListResult: {
        versions: [...#UpdateVersionResponse]
        pagination: #Pagination
}

// UpdateChangelogEntryResult — generated from openapi/schemas.go.
#UpdateChangelogEntryResult: {
        version: string
        date: string
        type: string
        notes: string
}

// UpdateChangelogResult — generated from openapi/schemas.go.
#UpdateChangelogResult: {
        changelog: [...#UpdateChangelogEntryResult]
}

// UpdateSyncStatusInfo — generated from openapi/schemas.go.
#UpdateSyncStatusInfo: {
        status: string
        error?: string
        lastSyncAt?: int64
        nextSyncAt?: int64
        versionsFound?: int
}

// UpdateLatestVersionInfo — generated from openapi/schemas.go.
#UpdateLatestVersionInfo: {
        version: string
        releaseType: string
        apkFilename: string
        sha256: string
        releasedAt: int64
        apkSize: int64
}

// UpdateStatusResult — generated from openapi/schemas.go.
#UpdateStatusResult: {
        latest?: #UpdateLatestVersionInfo
        device?: #UpdateDeviceStatusInfo
        sync: #UpdateSyncStatusInfo
}

// UpdateSyncResponse — generated from openapi/schemas.go.
#UpdateSyncResponse: {
        status: string
        message?: string
        startedAt: int64
        versionsFound?: int
}

// UpdateExportResult — generated from openapi/schemas.go.
#UpdateExportResult: {
        format: string
        versions: [...#UpdateVersionResponse]
        changelog: [...#UpdateChangelogEntryResult]
        exportedAt: int64
}

// UpdatePushResult — generated from openapi/schemas.go.
#UpdatePushResult: {
        pushId: string
        version: string
        installType: string
        scheduledAt?: int64
        initiatedBy: string
        status: string
        deviceIds: [...string]
        failedDevices?: [...#UpdateFailedDevice]
        devices: #UpdatePushDeviceCounts
        initiatedAt: int64
}

// UpdatePushHistoryEntry — generated from openapi/schemas.go.
#UpdatePushHistoryEntry: {
        completedAt?: int64
        cancelledAt?: int64
        scheduledAt?: int64
        id: string
        version: string
        installType: string
        status: string
        initiatedBy: string
        devices: #UpdatePushHistoryDeviceCounts
        deviceCount: int
        initiatedAt: int64
}

// UpdatePushHistoryListResult — generated from openapi/schemas.go.
#UpdatePushHistoryListResult: {
        pushes: [...#UpdatePushHistoryEntry]
        pagination: #Pagination
}

// UpdatePushDetailResult — generated from openapi/schemas.go.
#UpdatePushDetailResult: {
        scheduledAt?: int64
        completedAt?: int64
        cancelledAt?: int64
        id: string
        version: string
        installType: string
        status: string
        initiatedBy: string
        devices: [...#UpdatePushDetailDevice]
        initiatedAt: int64
}

// UpdateCancelPushResult — generated from openapi/schemas.go.
#UpdateCancelPushResult: {
        id: string
        status: string
        cancelledBy: string
        cancelledAt: int64
}

// UpdatePushRequest — generated from openapi/schemas.go.
#UpdatePushRequest: {
        scheduledAt?: int64
        version: string
        installType: string
        deviceIds: [...string]
}

// UpdateSyncStatusResult — generated from openapi/schemas.go.
#UpdateSyncStatusResult: {
        status: string
        error?: string
        lastSyncAt?: int64
        nextSyncAt?: int64
        versionsFound?: int
}

// UpdateCheckerResult — generated from openapi/schemas.go.
#UpdateCheckerResult: {
        usage_stats?: #UsageStatsSnapshot
        latest_version?: string
        current_version: string
        release_name?: string
        release_url?: string
        update_available: bool
}
