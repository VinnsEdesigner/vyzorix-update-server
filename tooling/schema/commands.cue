package vyzorix

// CommandRequest.
#CommandRequest: {
        args?: {[string]: _}
        command: string
        confirmation_token?: string
        dispatch_id?: string
        nonce: string
        signature?: string
        timestamp: int64
}

// CommandDispatchResult.
#CommandDispatchResult: {
        status: string
        dispatchId: string
        command_id: string
        serverTime: int64
        device_online: bool
}

// CommandStatus.
#CommandStatus: {
        dispatchId: string
        command_id: string
        device_id: string
        command: string
        status: string
        serverTime: int64
}

// CommandRetryResult.
#CommandRetryResult: {
        dispatchId: string
        command_id: string
        retried: bool
        serverTime: int64
}

// CommandCancelResult.
#CommandCancelResult: {
        dispatchId: string
        cancelled: bool
        serverTime: int64
}

// CommandPendingResult.
#CommandPendingResult: {
        commands: [...#CommandResponse]
}

// CommandResponse.
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

// CommandHistoryEntry.
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

// CommandHistoryResult.
#CommandHistoryResult: {
        commands: [...#CommandHistoryEntry]
        pagination: #Pagination
}

// CommandConfirmRequest.
#CommandConfirmRequest: {
        command: string
}

// CommandConfirmResult.
#CommandConfirmResult: {
        confirmation_token?: string
        risk_tier: string
        trace_id?: string
        expires_at?: int64
        confirmation_required: bool
        ttl_seconds: int
}
