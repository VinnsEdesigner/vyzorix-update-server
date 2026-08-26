package vyzorix

// DiagnosticsIdentityInfo — generated from openapi/schemas.go.
#DiagnosticsIdentityInfo: {
        imei: string
        deviceName?: string
        model?: string
        manufacturer?: string
}

// DiagnosticsSoftwareInfo — generated from openapi/schemas.go.
#DiagnosticsSoftwareInfo: {
        osVersion: string
        appVersion: string
        securityPatch?: string
        buildId?: string
}

// DiagnosticsRegistrationInfo — generated from openapi/schemas.go.
#DiagnosticsRegistrationInfo: {
        status: string
        registeredAt?: int64
        fcmTokenRefreshedAt?: int64
        fcmTokenValid: bool
        commandSecretSet: bool
}

// DiagnosticsTelemetryInfo — generated from openapi/schemas.go.
#DiagnosticsTelemetryInfo: {
        lastTimestamp: int64
        framesToday: int
        avgLatencyMs?: int
        totalBytesToday: int64
        sessionsToday: int
}
