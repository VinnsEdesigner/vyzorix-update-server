package vyzorix

// DeviceStatus — generated from openapi/schemas.go.
#DeviceStatus: {
        device_id: string
        app_version: string
        device_class: string
        last_seen: int64
        online: bool
}

// DeviceListItem — generated from openapi/schemas.go.
#DeviceListItem: {
        registered_at?: int64
        id: string
        imei?: string
        device_name?: string
        model?: string
        manufacturer?: string
        app_version?: string
        status: string
        last_seen?: int64
        online?: bool
}

// DeviceListResult — generated from openapi/schemas.go.
#DeviceListResult: {
        nextCursor?: int
        devices: [...#DeviceListItem]
        total: int64
}

// DeviceDetailResult — generated from openapi/schemas.go.
#DeviceDetailResult: {
        registered_at?: int64
        id: string
        imei?: string
        device_name?: string
        model?: string
        manufacturer?: string
        app_version?: string
        status: string
        last_seen?: int64
}

// DeviceCountResult — generated from openapi/schemas.go.
#DeviceCountResult: {
        count: int
        serverTime?: int64
}

// DeviceTagsResult — generated from openapi/schemas.go.
#DeviceTagsResult: {
        tags: [...string]
}

// SetDeviceTagsRequest — generated from openapi/schemas.go.
#SetDeviceTagsRequest: {
        tags: [...string]
}

// DeviceConfirmResult — generated from openapi/schemas.go.
#DeviceConfirmResult: {
        registered_at?: int64
        device_id: string
        imei: string
        server_time: int64
        confirmed: bool
        online: bool
}

// DeviceConfirmRequest — generated from openapi/schemas.go.
#DeviceConfirmRequest: {
        imei: string
        commandSecret: string
}

// DeviceSettingsResult — generated from openapi/schemas.go.
#DeviceSettingsResult: {
        id: string
        deviceImei: string
        customName?: string
        location?: string
        metadata?: {[string]: _}
        thresholds?: #OperatorThresholds
        createdAt: string
        updatedAt: string
}

// UpdateDeviceSettingsRequest — generated from openapi/schemas.go.
#UpdateDeviceSettingsRequest: {
        customName?: string
        location?: string
        metadata?: {[string]: _}
        thresholds?: #OperatorThresholds
}

// DeviceTagAddedResult — generated from openapi/schemas.go.
#DeviceTagAddedResult: {
        added: string
}

// DeviceTagRemovedResult — generated from openapi/schemas.go.
#DeviceTagRemovedResult: {
        removed: string
}

// DeviceTransferRequest — generated from openapi/schemas.go.
#DeviceTransferRequest: {
        target_organization_id: string
}

// DeviceTransferResult — generated from openapi/schemas.go.
#DeviceTransferResult: {
        message?: string
        device_id?: string
        from_org_id?: string
        to_org_id?: string
        success: bool
}

// DeviceFCMTokenRequest — generated from openapi/schemas.go.
#DeviceFCMTokenRequest: {
        fcmToken: string
}

// DeviceDisconnectResult — generated from openapi/schemas.go.
#DeviceDisconnectResult: {
        deviceId: string
        operatorId?: string
        disconnected: bool
}

// DeviceEvent — generated from openapi/schemas.go.
#DeviceEvent: {
        created_at: string
        data: {[string]: _}
        id: string
        device_id: string
        type: string
}

// DeviceEventListResult — generated from openapi/schemas.go.
#DeviceEventListResult: {
        events: [...#DeviceEvent]
}

// DeviceLog — generated from openapi/schemas.go.
#DeviceLog: {
        timestamp: string
        level: string
        message: string
        source?: string
}

// DeviceLogListResult — generated from openapi/schemas.go.
#DeviceLogListResult: {
        logs: [...#DeviceLog]
        pagination: #Pagination
}

// UpdateDeviceStatusInfo — generated from openapi/schemas.go.
#UpdateDeviceStatusInfo: {
        currentVersion?: string
        needsUpdate: bool
}

// UpdatePushDeviceCounts — generated from openapi/schemas.go.
#UpdatePushDeviceCounts: {
        total: int
        pending: int
        sent: int
        acknowledged: int
        failed: int
}

// UpdateFailedDevice — generated from openapi/schemas.go.
#UpdateFailedDevice: {
        deviceId: string
        reason: string
}

// UpdatePushHistoryDeviceCounts — generated from openapi/schemas.go.
#UpdatePushHistoryDeviceCounts: {
        pending?: int
        sent?: int
        acknowledged?: int
        failed?: int
}

// UpdatePushDetailDevice — generated from openapi/schemas.go.
#UpdatePushDetailDevice: {
        id: string
        deviceId: string
        deviceName?: string
        status: string
        sentAt?: int64
        acknowledgedAt?: int64
        error?: string
}

// DeviceUpdateStatusRequest — generated from openapi/schemas.go.
#DeviceUpdateStatusRequest: {
        dispatchId: string
        deviceId: string
        status: string
        error?: string
}

// DeviceUpdateStatusResponse — generated from openapi/schemas.go.
#DeviceUpdateStatusResponse: {
        message: string
        acknowledged: bool
}

// DeviceInspection — generated from openapi/schemas.go.
#DeviceInspection: {
        battery?: float64
        device_id: string
        app_version: string
        last_seen: int64
        online: bool
        fcm_token_valid: bool
}

// DeviceInspectionResult — generated from openapi/schemas.go.
#DeviceInspectionResult: {
        connection: #DiagnosticsConnectionInfo
        identity: #DiagnosticsIdentityInfo
        software: #DiagnosticsSoftwareInfo
        registration: #DiagnosticsRegistrationInfo
        telemetry: #DiagnosticsTelemetryInfo
}

// DeviceLogEvent — generated from openapi/schemas.go.
#DeviceLogEvent: {
        data?: {[string]: _}
        id: string
        type: string
        timestamp: int64
}

// DeviceLogEventListResult — generated from openapi/schemas.go.
#DeviceLogEventListResult: {
        events: [...#DeviceLogEvent]
        pagination: #CursorPaginationResult
}
