package vyzorix

// TelemetryFrameDTO.
#TelemetryFrameDTO: {
        timestamp: int64
        uptime: int64
        riskScore: float64
        thermalTemp: float64
        bufferLevel: float64
}

// TelemetryStatsDTO.
#TelemetryStatsDTO: {
        riskScore: #MetricStatsDTO
        thermalTemp: #MetricStatsDTO
        bufferLevel: #MetricStatsDTO
}

// GetTelemetryResponse.
#GetTelemetryResponse: {
        frames: [...#TelemetryFrameDTO]
        stats: #TelemetryStatsDTO
}

// TelemetryEntry.
#TelemetryEntry: {
        timestamp: int64
        metrics: {[string]: float64}
        id?: string
        device_id?: string
        received_at: int64
        risk_score: int
        buffer_level: int
        thermal_temp: int
}

// TelemetryHistoryResponse.
#TelemetryHistoryResponse: {
        device_id: string
        entries: [...#TelemetryEntry]
        pagination: #Pagination
}

// TelemetryStatsResult.
#TelemetryStatsResult: {
        deviceId: string
        latestEntry: int64
        oldestEntry: int64
        riskScore: #MetricAggregateResult
        bufferLevel: #MetricAggregateResult
        thermalTemp: #MetricAggregateResult
        sampleCount: int
}

// TelemetryHistoryEntry.
#TelemetryHistoryEntry: {
        receivedAt: int64
        id: string
        deviceId: string
        payload?: string
        riskScore?: int
        bufferLevel?: int
        thermalTemp?: float64
}

// TelemetryHistoryQueryResult.
#TelemetryHistoryQueryResult: {
        deviceId: string
        entries: [...#TelemetryHistoryEntry]
        totalCount: int
        startTime: int64
        endTime: int64
        queryTime: int64
        exported_at: int64
        count: int
}

// TelemetryCleanupResult.
#TelemetryCleanupResult: {
        deleted: int
        older_than: int64
}
