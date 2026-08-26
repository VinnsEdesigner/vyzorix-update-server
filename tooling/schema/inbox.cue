package vyzorix

// InboxRequest — generated from openapi/schemas.go.
#InboxRequest: {
        imei: string
        deviceName: string
        deviceClass: string
        model: string
        manufacturer: string
        osVersion: string
        appVersion: string
        fcmToken: string
        firebaseInstallId: string
        idempotencyKey?: string
}

// InboxAckRequest — generated from openapi/schemas.go.
#InboxAckRequest: {
        action: string
        notes?: string
}

// InboxEntryResponse — generated from openapi/schemas.go.
#InboxEntryResponse: {
        acknowledgedAt?: int64
        approvingAt?: int64
        approvedAt?: int64
        rejectedAt?: int64
        model: string
        firebaseInstallId: string
        id: string
        manufacturer: string
        osVersion: string
        appVersion: string
        fcmToken: string
        deviceClass?: string
        status: string
        operatorId?: string
        deviceName?: string
        imei: string
        notes?: string
        createdAt: int64
}

// InboxListResult — generated from openapi/schemas.go.
#InboxListResult: {
        requests: [...#InboxEntryResponse]
        pagination: #Pagination
}

// InboxAckResult — generated from openapi/schemas.go.
#InboxAckResult: {
        acknowledgedAt?: int64
        approvingAt?: int64
        approvedAt?: int64
        rejectedAt?: int64
        id: string
        imei: string
        status: string
        commandSecret?: string
        notes?: string
        fcmPushSent: bool
}

// InboxResendResult — generated from openapi/schemas.go.
#InboxResendResult: {
        id: string
        imei: string
        status: string
        fcmPushSent: bool
}
