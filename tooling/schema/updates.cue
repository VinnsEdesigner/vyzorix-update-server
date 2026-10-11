package vyzorix

// UpdateAPIKeyRequest.
#UpdateAPIKeyRequest: {
        name?: string
        scope?: string
}

// UpdateAdminClientRequest.
#UpdateAdminClientRequest: {
        name?: string
        rate_limit?: int
        is_active?: bool
        allowed_origins?: [...string]
        allowed_paths?: [...string]
}

// UpdateOperatorRequest.
#UpdateOperatorRequest: {
        email?: string
        name?: string
        role?: string
        mfa_enabled?: bool
        email_verified?: bool
}

// UpdateSettingsRequest.
#UpdateSettingsRequest: {
        client?: #ClientSettings
        name?: string
        reset?: bool
}

// ThresholdUpdateRequest.
#ThresholdUpdateRequest: {
        riskWarn?: int
        riskCrit?: int
        thermalWarn?: int
        thermalCrit?: int
        bufferWarn?: int
        bufferCrit?: int
}

// NotificationUpdateRequest.
#NotificationUpdateRequest: {
        enabled?: bool
        channels?: [...string]
        email?: #EmailNotifications
        push?: #PushNotifications
        webhook?: #WebhookNotifications
}

// UpdateNameRequest.
#UpdateNameRequest: {
        name: string
}

// UpdateInboxEntryRequest.
#UpdateInboxEntryRequest: {
        notes?: string
        status?: string
}

// UpdateVersionManifest.
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

// UpdateChangelogEntry.
#UpdateChangelogEntry: {
        version: string
        date: int64
        type: string
        notes: string
}

// UpdateCheckRequest.
#UpdateCheckRequest: {
        current_version: string
}

// UpdateCheckResult.
#UpdateCheckResult: {
        latest_version?: string
        current_version: string
        release_notes?: string
        download_url?: string
        sha256?: string
        apk_size?: int64
        update_available: bool
}

// UpdaterVersionManifestResult.
#UpdaterVersionManifestResult: {
        version: string
        apk_filename: string
        apk_sha256: string
        release_notes: string
        version_code: int
        apk_size_bytes: int64
}

// UpdaterCheckResult.
#UpdaterCheckResult: {
        version: string
        apk_filename: string
        apk_sha256: string
        release_notes: string
        version_code: int
        apk_size_bytes: int64
        update_available: bool
}

// UpdateVersionResponse.
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

// UpdateVersionListResult.
#UpdateVersionListResult: {
        versions: [...#UpdateVersionResponse]
        pagination: #Pagination
}

// UpdateChangelogEntryResult.
#UpdateChangelogEntryResult: {
        version: string
        date: string
        type: string
        notes: string
}

// UpdateChangelogResult.
#UpdateChangelogResult: {
        changelog: [...#UpdateChangelogEntryResult]
}

// UpdateSyncStatusInfo.
#UpdateSyncStatusInfo: {
        status: string
        error?: string
        lastSyncAt?: int64
        nextSyncAt?: int64
        versionsFound?: int
}

// UpdateLatestVersionInfo.
#UpdateLatestVersionInfo: {
        version: string
        releaseType: string
        apkFilename: string
        sha256: string
        releasedAt: int64
        apkSize: int64
}

// UpdateStatusResult.
#UpdateStatusResult: {
        latest?: #UpdateLatestVersionInfo
        device?: #UpdateDeviceStatusInfo
        sync: #UpdateSyncStatusInfo
}

// UpdateSyncResponse.
#UpdateSyncResponse: {
        status: string
        message?: string
        startedAt: int64
        versionsFound?: int
}

// UpdateExportResult.
#UpdateExportResult: {
        format: string
        versions: [...#UpdateVersionResponse]
        changelog: [...#UpdateChangelogEntryResult]
        exportedAt: int64
}

// UpdatePushResult.
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

// UpdatePushHistoryEntry.
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

// UpdatePushHistoryListResult.
#UpdatePushHistoryListResult: {
        pushes: [...#UpdatePushHistoryEntry]
        pagination: #Pagination
}

// UpdatePushDetailResult.
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

// UpdateCancelPushResult.
#UpdateCancelPushResult: {
        id: string
        status: string
        cancelledBy: string
        cancelledAt: int64
}

// UpdatePushRequest.
#UpdatePushRequest: {
        scheduledAt?: int64
        version: string
        installType: string
        deviceIds: [...string]
}

// UpdateSyncStatusResult.
#UpdateSyncStatusResult: {
        status: string
        error?: string
        lastSyncAt?: int64
        nextSyncAt?: int64
        versionsFound?: int
}

// UpdateCheckerResult.
#UpdateCheckerResult: {
        usage_stats?: #UsageStatsSnapshot
        latest_version?: string
        current_version: string
        release_name?: string
        release_url?: string
        update_available: bool
        error?: string
}

// WebhookInfoResult.
#WebhookInfoResult: {
        webhook_enabled: bool
        supported_events: [...string]
}
