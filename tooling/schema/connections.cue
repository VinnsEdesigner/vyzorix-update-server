package vyzorix

// ConnectionStatusResult.
#ConnectionStatusResult: {
        device_id: string
        status?: string
        online: bool
}

// ConnectionListResult.
#ConnectionListResult: {
        connections: [...#ConnectionStatusResult]
}

// ConnectionMetricsResult.
#ConnectionMetricsResult: {
        total_connections: int
        online_connections: int
        timestamp: int64
        total_connected: int
        total_queued: int
        aggregate_metrics?: {[string]: int64}
        queue_metrics?: _
}

// DiagnosticsConnectionInfo.
#DiagnosticsConnectionInfo: {
        webSocketStatus: string
        fcmStatus: string
        clientIp?: string
        protocol: string
        connectedAt?: int64
        lastSeen?: int64
}
