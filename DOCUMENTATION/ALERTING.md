# Alerting

Vyzorix watches your fleet and tells you when something goes wrong. You write a
rule — "more than ten devices offline", "more than a fifth of the fleet is
offline", "commands are failing more than 5% of the time" — and the server
evaluates it on a fixed cadence. When the condition holds long enough, the rule
fires. When it clears, it resolves. Every transition is recorded, and firing
alerts are delivered through your contact points.

This is the part of the server that turns raw device data into "hey, look at
this."

---

## What a rule is

A rule is an org-scoped definition. It belongs to one organization, and only
that organization's devices are measured against it.

| Field | Meaning |
|:------|:--------|
| `name` | Human label for the rule. |
| `metric` | What to measure (see the table below). |
| `condition` | Comparator: `gt`, `gte`, `lt`, or `lte`. |
| `threshold` | The number the measurement is compared against. |
| `for_seconds` | How long the breach must hold before firing. Zero means "fire on the first breach". |
| `notify_interval_seconds` | Minimum gap between repeat notifications while firing. Zero notifies on every transition only. |
| `on_no_data` | What to do when the metric has no signal: `ignore`, `no_data`, or `resolve`. |
| `on_error` | What to do when the metric query fails: `ignore`, `error`, or `resolve`. |
| `webhook_url` | Optional single webhook destination, kept for simple setups. |
| `enabled` | Off rules are stored but never evaluated. |

### The three metrics

| Metric | Measures | Fans out per |
|:-------|:---------|:-------------|
| `device_offline_count` | Devices currently offline. | `device_class`, plus a fleet-wide total. |
| `device_offline_percent` | Share of the fleet offline, 0–100. | `device_class`, plus a fleet-wide total. |
| `command_failure_rate` | Percent of commands that failed inside the observation window, 0–100. | Fleet-wide only. |

The offline metrics are point-in-time — they read the current `online` flag on
each device. `command_failure_rate` is windowed; the window defaults to one hour
and is configurable with `ALERT_METRIC_WINDOW_SECONDS`. When a window carries no
commands at all, the metric reports *no data* rather than a misleading zero.

---

## How a rule moves through its states

Every rule instance lives in exactly one state, and only ever moves along these
edges:

```
              breach           for_seconds elapsed
  inactive ──────────▶ pending ──────────────────▶ firing
     ▲                    │                          │
     │   cleared          │  cleared                 │  cleared
     └────────────────────┴──────────────────────────┘

  any state ──▶ no_data     (on_no_data = no_data, metric went quiet)
  any state ──▶ error       (on_error = error, metric query failed)
  no_data/error ──▶ inactive (on_no_data / on_error = resolve, or the metric returns)
```

- **inactive** — the condition is not breached.
- **pending** — the condition is breached, but hasn't held for `for_seconds`
  yet. This is what stops a single noisy sample from paging you.
- **firing** — the condition held long enough. The alert is live.
- **no_data** — the metric produced no signal this tick and the rule asked to be
  told about that (`on_no_data: no_data`).
- **error** — the metric query failed and the rule asked to be told
  (`on_error: error`).

`ignore` leaves the instance wherever it was — a quiet metric simply doesn't
change the state. `resolve` treats the missing or failed signal as "the problem
went away" and returns the instance to inactive.

### Instances, labels, and fan-out

A rule doesn't have one state — it has one **instance** per labeled series. The
offline metrics fan out by `device_class`, so a single rule tracks one instance
per class plus an unlabeled fleet aggregate. That's what lets you write "any
class is more than 20% offline" without writing a rule per class.

Instances are keyed by a deterministic hash of their labels, so the same labels
always land on the same instance across ticks. A rule with no labels collapses
to a single fleet-wide instance.

---

## What happens when a rule fires

A transition is the interesting moment, and the server does four things with it:

1. **Records history.** Every `from → to` transition is written to the alert
   history with the observed value and a timestamp. The live instance table
   holds only current state; history is the audit trail.
2. **Notifies.** Firing and resolved transitions are dispatched to the
   organization's contact points (see [Contact Points](CONTACT_POINTS.md)). If
   the rule has a `webhook_url`, that destination is used too. Repeat
   notifications respect `notify_interval_seconds`; the last-attempt time is
   stamped so a slow channel can't be hammered.
3. **Pushes to the live channel.** The transition is published to the
   organization's `stream/<org>/alerts` channel, so dashboards subscribed over
   the websocket light up the moment the state changes. See
   [Live Channels](LIVE_CHANNELS.md).
4. **Drops an annotation.** The evaluator can drop an annotation onto the
   timeline for the transition, so a firing alert is visible next to the fleet
   events that surrounded it. See [Annotations](ANNOTATIONS.md).

---

## Evaluation

Evaluation runs on a background worker registered with the
[service lifecycle registry](SERVICE_LIFECYCLE.md). It ticks on an interval
(`ALERT_EVAL_INTERVAL_SECONDS`, default 60s), and the first tick carries a random
jitter so multiple server instances don't all evaluate at the same instant.

Each tick evaluates every enabled rule in every organization. The worker holds
the server lock, so if you ever run more than one instance, only one evaluates at
a time — rules are never double-fired.

If a metric query fails, the rule's `on_error` policy decides what happens. The
failure is contained to that rule; other rules keep evaluating.

---

## Managing rules over REST

All alert routes live under `/v1/alerts` and require organization context.

| Method | Path | Does |
|:-------|:-----|:-----|
| `GET` | `/v1/alerts/rules` | List the org's rules. |
| `POST` | `/v1/alerts/rules` | Create a rule. |
| `GET` | `/v1/alerts/rules/:id` | Read one rule. |
| `PATCH` | `/v1/alerts/rules/:id` | Update a rule. |
| `DELETE` | `/v1/alerts/rules/:id` | Delete a rule. |
| `POST` | `/v1/alerts/rules/:id/evaluate` | Evaluate a rule right now, on demand. |
| `GET` | `/v1/alerts/history` | List transitions for the org. |
| `GET` | `/v1/alerts/rules/:id/history` | List transitions for one rule. |

Reading needs the `alert.read` action; writing needs `alert.write`, both scoped
to `org:*`. The manual evaluate endpoint is useful for testing a rule's
definition before you trust it in the loop.

---

## Configuration

| Variable | Default | Meaning |
|:---------|:--------|:--------|
| `ALERT_EVAL_INTERVAL_SECONDS` | `60` | How often the evaluator ticks. |
| `ALERT_METRIC_WINDOW_SECONDS` | `3600` | Observation window for `command_failure_rate`. |
| `ALERT_WEBHOOK_SECRET` | *(empty)* | HMAC secret used to sign outbound alert webhooks. |

---

## A worked example

You want to know when a class of devices starts dropping off.

```json
POST /v1/alerts/rules
{
  "name": "Warehouse units going offline",
  "metric": "device_offline_count",
  "condition": "gt",
  "threshold": 10,
  "for_seconds": 300,
  "notify_interval_seconds": 3600,
  "on_no_data": "ignore",
  "on_error": "error",
  "enabled": true
}
```

The evaluator reads the offline count per `device_class` each minute. When the
`warehouse` class crosses ten offline and stays there for five minutes, that
class's instance goes `pending → firing`, a history event is written, the org's
contact points are notified, and the alert channel gets a frame. If it drops back
under ten, the instance resolves and a resolution is recorded and delivered.
