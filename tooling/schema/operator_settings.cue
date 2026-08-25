package vyzorix

// ClientSettings controls dashboard behavior for an operator's client.
#ClientSettings: {
        serverUrl:            string
        deviceId:             string
        requestTimeoutMs:     int | *30000
        logBufferLimit:       int | *100
        signalHistoryLimit:   int | *100
        autoReconnect:        bool | *true
        strictHmac:           bool | *false
        notificationsEnabled: bool | *true
}

// OperatorSettingsResultLegacy is the flat response shape for GET /me/settings.
#OperatorSettingsResultLegacy: {
        id:             string
        email:          string
        name:           string
        role?:          string
        mfa_enabled:    bool
        email_verified: bool
        client?:        #ClientSettings
}
