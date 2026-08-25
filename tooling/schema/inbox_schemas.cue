package vyzorix

// InboxRequest is the device registration request body.
#InboxRequest: {
        imei:              string
        deviceName?:       string
        deviceClass?:      string
        model?:            string
        manufacturer?:     string
        osVersion?:        string
        appVersion?:       string
        fcmToken:          string
        firebaseInstallId: string
        idempotencyKey?:   string
}

// DeviceConfirmRequest is the body for POST /v1/device/confirm.
#DeviceConfirmRequest: {
        imei:          string
        commandSecret: string
}

// InboxAckRequest is the body for acknowledging an inbox entry.
#InboxAckRequest: {
        action: string
        notes?: string
}

// UpdateInboxEntryRequest is the body for PATCH inbox entry.
#UpdateInboxEntryRequest: {
        notes?: string
}
