package vyzorix

// ConnectionStatusResult — generated from openapi/schemas.go.
#ConnectionStatusResult: {
        device_id: string
        status?: string
        online: bool
}

// ConnectionListResult — generated from openapi/schemas.go.
#ConnectionListResult: {
        connections: [...#ConnectionStatusResult]
}

// ConnectionMetricsResult — generated from openapi/schemas.go.
#ConnectionMetricsResult: {
        total_connections: int
        online_connections: int
}

// DiagnosticsConnectionInfo — generated from openapi/schemas.go.
#DiagnosticsConnectionInfo: {
        webSocketStatus: string
        fcmStatus: string
        clientIp?: string
        protocol: string
        connectedAt?: int64
        lastSeen?: int64
}
