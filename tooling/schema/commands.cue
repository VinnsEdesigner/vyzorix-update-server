package vyzorix

// CommandRequest — generated from openapi/schemas.go.
#CommandRequest: {
        args?: {[string]: _}
        command: string
        confirmation_token?: string
        dispatch_id?: string
        nonce: string
        signature?: string
        timestamp: int64
}

// CommandDispatchResult — generated from openapi/schemas.go.
#CommandDispatchResult: {
        status: string
        dispatchId: string
        command_id: string
        serverTime: int64
        device_online: bool
}

// CommandStatus — generated from openapi/schemas.go.
#CommandStatus: {
        dispatchId: string
        command_id: string
        device_id: string
        command: string
        status: string
        serverTime: int64
}

// CommandRetryResult — generated from openapi/schemas.go.
#CommandRetryResult: {
        dispatchId: string
        command_id: string
        retried: bool
        serverTime: int64
}

// CommandCancelResult — generated from openapi/schemas.go.
#CommandCancelResult: {
        dispatchId: string
        cancelled: bool
        serverTime: int64
}

// CommandPendingResult — generated from openapi/schemas.go.
#CommandPendingResult: {
        commands: [...#CommandResponse]
}

// CommandResponse — generated from openapi/schemas.go.
#CommandResponse: {
        id?: string
        deviceId?: string
        command?: string
        dispatchId?: string
        status?: string
        delivery?: string
        args?: [...string]
        serverTime?: int64
}

// CommandHistoryEntry — generated from openapi/schemas.go.
#CommandHistoryEntry: {
        id?: string
        dispatchId: string
        deviceId: string
        command: string
        status: string
        failureReason?: string
        createdAt: int64
        sentAt: int64
        deliveredAt?: int64
        completedAt?: int64
        latencyMs?: int64
}

// CommandHistoryResult — generated from openapi/schemas.go.
#CommandHistoryResult: {
        commands: [...#CommandHistoryEntry]
        pagination: #Pagination
}

// CommandConfirmRequest — generated from openapi/schemas.go.
#CommandConfirmRequest: {
        command: string
}

// CommandConfirmResult — generated from openapi/schemas.go.
#CommandConfirmResult: {
        confirmation_token?: string
        risk_tier: string
        trace_id?: string
        expires_at?: int64
        confirmation_required: bool
}
