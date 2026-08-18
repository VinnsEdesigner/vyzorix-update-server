# Command Execution

Commands are how the server tells a device to do something — check for updates, wake up, reboot, factory reset. The command system handles creation, signing, delivery, retries, and status tracking.

## Command lifecycle

A command goes through these states (defined in `internal/domain/command/command_entity.go`):

```
pending → delivered → completed
                   ↘ → failed
pending → cancelled
delivered → cancelled
```

- **pending** — created, waiting to be sent to the device
- **delivered** — sent to the device via WebSocket or FCM, awaiting response
- **completed** — device confirmed the command executed successfully
- **failed** — device reported failure, or the command timed out
- **cancelled** — operator cancelled the command before completion

Terminal states (completed, failed, cancelled) cannot transition further.

## Creating a command

`POST /v1/device/:imei/command` is the entry point. The request includes the command name, optional arguments, and (for risky commands) a confirmation token.

The `ExecuteHandler.Handle` method in `internal/api/handlers/command/command_execute.go` does this:

1. Validates the request (deviceId, command, nonce)
2. Checks the device belongs to the caller's organization
3. Runs the risk gate (`authorizeCommand`) — classifies the command, checks MFA, consumes confirmation tokens if needed
4. Calls `commandService.SendCommand` which creates a `Command` entity in the database with status `pending`
5. Builds a `CommandFrame` (the wire format for the device)
6. Signs the frame with the device's `CommandSecretHash` (HMAC-SHA512)
7. Delivers via WebSocket if the device is online, or FCM silent wake if not
8. If delivered via WebSocket, marks the command as `delivered`
9. Audits the execution attempt

## Command signing

Every command sent to a device is HMAC-signed. The server generates a nonce and timestamp, computes `HMAC-SHA512(frame, deviceSecret)`, and includes the signature in the frame. The device verifies the signature before executing — this prevents command injection.

The `CommandSigner` in `internal/infrastructure/crypto/` handles this. The device's secret is stored as a SHA-256 hash (not plaintext), and both the server and device can compute the same HMAC key from it.

## Delivery: WebSocket vs FCM

The `deliverCommand` method tries WebSocket first:

```go
if h.hub != nil && h.hub.Online(imei) {
    if sent := h.hub.Send(imei, frame); sent {
        delivery = "sent"
    }
}
```

If the device isn't connected via WebSocket (offline), it falls back to FCM:

```go
if delivery == "queued" && h.fcmNotifier != nil {
    // Send a silent push notification to wake the device
    h.fcmNotifier.SendSilentWake(ctx, wake)
}
```

The FCM notification tells the device to connect and fetch pending commands. The command stays in `pending` status until the device connects and receives it.

## Command outbox

The `Outbox` in `internal/application/command/command_outbox.go` is a background worker that polls for pending commands every second and attempts delivery. It handles retries with exponential backoff.

```go
type OutboxConfig struct {
    PollInterval   time.Duration  // 1 second
    MaxRetries     int            // 5
    RetryBaseDelay time.Duration  // base for exponential backoff
    BatchSize      int            // commands per poll cycle
}
```

Started in the wire layer via `ProvideCommandOutbox`.

## Command status

`GET /v1/command/:dispatchId/status` returns the current status of a command. The handler verifies the command belongs to the caller's organization.

## Command retry

`POST /v1/command/:dispatchId/retry` creates a new command with the same parameters as the original failed command. The original stays in its terminal state; the retry is a fresh command.

## Command cancellation

`DELETE /v1/command/:dispatchId` cancels a pending or delivered command. If the device is online, a cancellation notification is sent.

## Pending commands

`GET /v1/device/:imei/commands/pending` returns all pending commands for a device. Used by the device when it connects to fetch anything it missed.

## Risk integration

The risk gate runs before command creation. If the command requires confirmation and no valid token is presented, the handler returns 425 Too Early and never creates the command. See [Risk & Confirmation](RISK_AND_CONFIRMATION.md) for details.

## Audit integration

Every command execution attempt — whether it succeeds, is blocked, or fails — generates an audit entry. The audit includes the command name, device, risk tier, result, and the state transition (e.g., pending → delivered). See [Audit System](AUDIT_SYSTEM.md) for details.
