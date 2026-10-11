package vyzorix

// DeviceStatus.
#DeviceStatus: {
        device_id: string
        app_version: string
        device_class: string
        last_seen: int64
        online: bool
}

// DeviceListItem.
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

// DeviceListResult.
#DeviceListResult: {
        nextCursor?: int
        devices: [...#DeviceListItem]
        total: int64
}

// DeviceDetailResult.
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

// DeviceCountResult.
#DeviceCountResult: {
        count: int
        serverTime?: int64
}

// DeviceTagsResult.
#DeviceTagsResult: {
        tags: [...string]
}

// SetDeviceTagsRequest.
#SetDeviceTagsRequest: {
        tags: [...string]
}

// DeviceConfirmResult.
#DeviceConfirmResult: {
        registered_at?: int64
        device_id: string
        imei: string
        server_time: int64
        confirmed: bool
        online: bool
}

// DeviceConfirmRequest.
#DeviceConfirmRequest: {
        imei: string
        commandSecret: string
}

// DeviceSettingsResult.
#DeviceSettingsResult: {
        id: string
        deviceImei: string
        customName?: string
        location?: string
        metadata?: {[string]: string}
        thresholds?: #OperatorThresholds
        createdAt: string
        updatedAt: string
}

// UpdateDeviceSettingsRequest.
#UpdateDeviceSettingsRequest: {
        customName?: string
        location?: string
        metadata?: {[string]: string}
        thresholds?: #OperatorThresholds
}

// DeviceTagAddedResult.
#DeviceTagAddedResult: {
        added: string
}

// DeviceTagRemovedResult.
#DeviceTagRemovedResult: {
        removed: string
}

// DeviceTransferRequest.
#DeviceTransferRequest: {
        target_organization_id: string
}

// DeviceTransferResult.
#DeviceTransferResult: {
        message?: string
        device_id?: string
        from_org_id?: string
        to_org_id?: string
        success: bool
        device_imei?: string
        source_org?: string
        target_org?: string
}

// DeviceFCMTokenRequest.
#DeviceFCMTokenRequest: {
        fcmToken: string
}

// DeviceDisconnectResult.
#DeviceDisconnectResult: {
        deviceId: string
        operatorId?: string
        disconnected: bool
}

// DeviceEvent.
#DeviceEvent: {
        created_at: int64
        data: {[string]: _}
        id: string
        device_id: string
        type: string
        severity?: string
        source?: string
        timestamp: int64
}

// DeviceEventListResult.
#DeviceEventListResult: {
        events: [...#DeviceEvent]
        has_more?: bool
        count?: int
}

// DeviceLog.
#DeviceLog: {
        timestamp: int64
        level: string
        message: string
        source?: string
}

// DeviceLogListResult.
#DeviceLogListResult: {
        logs: [...#DeviceLog]
        pagination: #Pagination
}

// UpdateDeviceStatusInfo.
#UpdateDeviceStatusInfo: {
        currentVersion?: string
        needsUpdate: bool
}

// UpdatePushDeviceCounts.
#UpdatePushDeviceCounts: {
        total: int
        pending: int
        sent: int
        acknowledged: int
        failed: int
}

// UpdateFailedDevice.
#UpdateFailedDevice: {
        deviceId: string
        reason: string
}

// UpdatePushHistoryDeviceCounts.
#UpdatePushHistoryDeviceCounts: {
        pending?: int
        sent?: int
        acknowledged?: int
        failed?: int
}

// UpdatePushDetailDevice.
#UpdatePushDetailDevice: {
        id: string
        deviceId: string
        deviceName?: string
        status: string
        sentAt?: int64
        acknowledgedAt?: int64
        error?: string
}

// DeviceUpdateStatusRequest.
#DeviceUpdateStatusRequest: {
        dispatchId: string
        deviceId: string
        status: string
        error?: string
}

// DeviceUpdateStatusResponse.
#DeviceUpdateStatusResponse: {
        message: string
        acknowledged: bool
}

// DeviceInspection.
#DeviceInspection: {
        battery?: float64
        device_id: string
        app_version: string
        last_seen: int64
        online: bool
        fcm_token_valid: bool
}

// DeviceInspectionResult.
#DeviceInspectionResult: {
        connection: #DiagnosticsConnectionInfo
        identity: #DiagnosticsIdentityInfo
        software: #DiagnosticsSoftwareInfo
        registration: #DiagnosticsRegistrationInfo
        telemetry: #DiagnosticsTelemetryInfo
}

// DeviceLogEvent.
#DeviceLogEvent: {
        data?: {[string]: _}
        id: string
        type: string
        timestamp: int64
}

// DeviceLogEventListResult.
#DeviceLogEventListResult: {
        events: [...#DeviceLogEvent]
        pagination: #CursorPaginationResult
}
