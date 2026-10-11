# Service Lifecycle

The server runs more than it serves. Behind the HTTP layer there are background
services — workers that clean up deleted devices, retry failed pushes, evaluate
alert rules, collect usage stats. Each of them starts at boot and needs to stop
cleanly at shutdown.

The lifecycle registry is how that's organized: named services, started in
dependency order, stopped in reverse, each reporting its own health.

---

## What a background service is

Anything that runs on its own schedule and outlives a single request. It
implements three methods:

| Method | Called when | Does |
|:-------|:------------|:-----|
| `Start()` | Boot | Launches the service (usually a goroutine). |
| `Stop()` | Shutdown | Shuts it down gracefully. |
| `Healthy()` | On demand | Reports whether it's running correctly. |

---

## Registration and ordering

A service is registered under a name, with optional dependencies:

```go
registry.Register("device-deletion", deletionWorker, "outbox")
registry.Register("invite-cleanup", cleanupWorker, "device-deletion")
```

The registry computes the order at start time:

- A service starts **after** everything it depends on.
- Independent services start in a stable, deterministic order.
- Shutdown is the reverse — dependents stop before their dependencies.
- A **circular dependency is an error**, reported at startup rather than
  discovered as a hang later.

The point of ordering is that a worker can assume its dependencies are already
running when it starts, and still running when it shuts down. The invitation
cleanup worker, for instance, depends on the device deletion worker because both
take the same lock.

---

## Graceful shutdown

On `SIGTERM` or `SIGINT`, the server stops its background services in reverse
dependency order before it exits. Each service gets a chance to finish the work
in flight and release its resources — which matters most for services holding
the [server lock](BACKGROUND_WORKERS.md) or a database transaction.

Nothing is killed mid-operation. A worker either finishes its current pass or
stops at the next safe point.

---

## Health

Every registered service reports its own health, and `/healthz` reflects the
aggregate. If a worker has died or stalled, the health endpoint says so — which
is what a load balancer or an uptime check needs to decide whether the instance
should take traffic.

This is the difference between "the process is up" and "the process is working."
A server whose alert evaluator has silently stopped is not healthy, even though
it still answers HTTP.

---

## What's registered

The workers wired onto the registry at boot:

| Service | Does |
|:--------|:-----|
| Device deletion | Purges soft-deleted devices after their retention window. |
| Invitation cleanup | Expires stale invitations. |
| FCM retry | Re-sends push notifications that failed. |
| Command outbox | Drains queued commands to devices. |
| Alert evaluation | Ticks the alert evaluator. |
| Usage stats | Collects periodic self-telemetry. |
| Service-account expiry / leak scan | Expires old tokens and scans for leaked keys. |

Each of these takes the server lock before it runs, so a second instance never
runs the same worker at the same time.
