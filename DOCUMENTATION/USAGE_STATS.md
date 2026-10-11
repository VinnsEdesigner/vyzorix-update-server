# Usage Stats

The server keeps a small, periodic snapshot of itself — how many devices,
operators, organizations, alert rules, contact points, and annotations exist, and
which feature toggles are on. It's self-telemetry: the platform measuring its own
shape.

The snapshot isn't a dashboard. It exists for two things: the admin update
checker, which reports what version is running alongside what the deployment
contains, and operator support, where "how big is this install?" is the first
question in any triage.

---

## What the snapshot holds

| Field | Meaning |
|:------|:--------|
| `counts.devices` | Devices across the deployment. |
| `counts.operators` | Operator accounts. |
| `counts.organizations` | Organizations. |
| `counts.service_accounts` | Service accounts. |
| `counts.alert_rules` | Alert rules. |
| `counts.contact_points` | Contact points. |
| `counts.annotations` | Annotations. |
| `toggles` | The current state of each feature toggle. |
| `collected_at` | When the snapshot was taken. |

---

## How it's collected

A background worker registered with the [service lifecycle](SERVICE_LIFECYCLE.md)
runs on a **six-hour** interval. Each pass takes the server lock, queries the
counts, reads the feature-toggle manager, and stores the result as the latest
snapshot.

The interval is deliberately long. These numbers move slowly, the query touches
every table, and there's nothing to gain from asking more often. The lock means a
multi-instance deployment collects once, not once per instance.

---

## Where it surfaces

The admin update checker embeds the latest snapshot in its response as
`usage_stats`. That single call answers "what version is this, and what does this
deployment contain?" — the two facts you want when checking whether an upgrade is
available and whether it's the right size for the fleet.

```
GET /v1/admin/updates/check
{
  "current_version": "2.0.0",
  "latest_version": "2.1.0",
  "update_available": true,
  "usage_stats": {
    "counts": { "devices": 1284, "operators": 37, ... },
    "toggles": { "scoped_rbac": true, "device_groups": true, ... },
    "collected_at": 1767225600000
  }
}
```

---

## Why it's cheap and safe

The counts are `SELECT COUNT(*)` queries against indexed tables — no joins, no
scans of row data. Nothing collected is personally identifiable; it's the shape
of the install, not its contents. And because it's collected on a timer rather
than computed per request, the cost never lands on a user's request path.

The build version that accompanies it flows from the binary via `-ldflags -X
main.Version` — there's no hardcoded stamp to forget to bump, so the version in
the snapshot is always the version that's actually running.
