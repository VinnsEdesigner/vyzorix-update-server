package vyzorix

// DiagnosticsIdentityInfo.
#DiagnosticsIdentityInfo: {
        imei: string
        deviceName?: string
        model?: string
        manufacturer?: string
}

// DiagnosticsSoftwareInfo.
#DiagnosticsSoftwareInfo: {
        osVersion: string
        appVersion: string
        securityPatch?: string
        buildId?: string
}

// DiagnosticsRegistrationInfo.
#DiagnosticsRegistrationInfo: {
        status: string
        registeredAt?: int64
        fcmTokenRefreshedAt?: int64
        fcmTokenValid: bool
        commandSecretSet: bool
}

// DiagnosticsTelemetryInfo.
#DiagnosticsTelemetryInfo: {
        lastTimestamp: int64
        framesToday: int
        avgLatencyMs?: int
        totalBytesToday: int64
        sessionsToday: int
}
