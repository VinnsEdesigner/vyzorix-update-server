package vyzorix

// TelemetryFrameDTO — generated from openapi/schemas.go.
#TelemetryFrameDTO: {
        timestamp: int64
        uptime: int64
        riskScore: float64
        thermalTemp: float64
        bufferLevel: float64
}

// TelemetryStatsDTO — generated from openapi/schemas.go.
#TelemetryStatsDTO: {
        riskScore: #MetricStatsDTO
        thermalTemp: #MetricStatsDTO
        bufferLevel: #MetricStatsDTO
}

// GetTelemetryResponse — generated from openapi/schemas.go.
#GetTelemetryResponse: {
        frames: [...#TelemetryFrameDTO]
        stats: #TelemetryStatsDTO
}

// TelemetryEntry — generated from openapi/schemas.go.
#TelemetryEntry: {
        timestamp: string
        metrics: {[string]: _}
}

// TelemetryHistoryRequest — generated from openapi/schemas.go.
#TelemetryHistoryRequest: {
        device_id: string
        start_time: int64
        end_time: int64
        limit: int
}

// TelemetryHistoryResponse — generated from openapi/schemas.go.
#TelemetryHistoryResponse: {
        device_id: string
        entries: [...#TelemetryEntry]
        pagination: #Pagination
}

// TelemetryStatsResult — generated from openapi/schemas.go.
#TelemetryStatsResult: {
        deviceId: string
        latestEntry: string
        oldestEntry: string
        riskScore: #MetricAggregateResult
        bufferLevel: #MetricAggregateResult
        thermalTemp: #MetricAggregateResult
        sampleCount: int
}

// TelemetryHistoryEntry — generated from openapi/schemas.go.
#TelemetryHistoryEntry: {
        receivedAt: string
        id: string
        deviceId: string
        payload?: string
        riskScore?: int
        bufferLevel?: int
        thermalTemp?: float64
}

// TelemetryHistoryQueryResult — generated from openapi/schemas.go.
#TelemetryHistoryQueryResult: {
        deviceId: string
        entries: [...#TelemetryHistoryEntry]
        totalCount: int
        startTime: int64
        endTime: int64
        queryTime: int64
}
