package vyzorix

// InboxRequest.
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

// InboxAckRequest.
#InboxAckRequest: {
        action: string
        notes?: string
}

// InboxEntryResponse.
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

// InboxListResult.
#InboxListResult: {
        requests: [...#InboxEntryResponse]
        pagination: #Pagination
}

// InboxAckResult.
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

// InboxResendResult.
#InboxResendResult: {
        id: string
        imei: string
        status: string
        fcmPushSent: bool
}
