# Background Workers

The server runs several background workers that handle periodic tasks. They start at boot and stop gracefully on shutdown.

## Device deletion worker

`internal/infrastructure/worker/device_deletion_worker.go`

Cleans up deregistered devices. When a device is deregistered, its `DeletionScheduledAt` is set to 30 days in the future. This worker periodically checks for devices past that timestamp and permanently deletes them from the database.

- Interval: configurable (default 5 minutes in providers, 1 hour in api_main.go)
- Enabled via `DeviceDeletionEnabled` config flag
- Implements the 30-day retention policy for deregistered devices

## FCM retry worker

`internal/infrastructure/worker/fcm_retry_worker.go`

Retries failed FCM (Firebase Cloud Messaging) deliveries. When a command can't be delivered via WebSocket (device offline), the server sends an FCM silent push. If that FCM delivery fails, it's queued for retry.

- Interval: 30 seconds
- Uses the FCM circuit breaker to avoid hammering a failing FCM service

## Command outbox worker

`internal/application/command/command_outbox.go`

Polls for pending commands and attempts delivery. If a command was created but the device was offline and FCM failed, the outbox retries.

- Poll interval: 1 second (configurable)
- Max retries: 5
- Backoff: exponential (base delay * 2^retryCount)
- Batch size: configurable (commands per poll cycle)

## Worker lifecycle

All workers follow the same pattern:

```go
type Worker struct {
    stopCh chan struct{}  // signal to stop
    doneCh chan struct{}  // signal that the goroutine finished
}

func (w *Worker) Start() {
    go w.run()
}

func (w *Worker) Stop() {
    close(w.stopCh)
    <-doneCh  // wait for the goroutine to finish
}
```

In `api_main.go`, workers are started after the server starts and stopped before the HTTP server shuts down:

```go
deviceDeletionWorker.Start()
// ... server runs ...
<-ctx.Done()  // SIGINT/SIGTERM
deviceDeletionWorker.Stop()
```

The stop is synchronous — it blocks until the worker's goroutine has finished its current iteration. This prevents data corruption during shutdown.
