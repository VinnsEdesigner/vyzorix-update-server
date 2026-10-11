# Annotations

An annotation is a note pinned to a point — or a span — on your fleet's
timeline. "Firmware v2.3 rollout started." "Warehouse network maintenance."
"Switched the telemetry pipeline." They exist so that when you look back at a
spike in failures, you can see what you did right before it.

Graphs tell you *what* happened. Annotations tell you *why*.

---

## What an annotation is

An org-scoped marker with a time range, a title, freeform text, and tags.

| Field | Meaning |
|:------|:--------|
| `title` | Short label. Required. |
| `text` | The detail — what happened, who did it, what to know. |
| `start_time` | When the event starts. Required. |
| `end_time` | Optional. Set it for a span (a rollout window); leave it out for a point in time. |
| `tags` | Freeform labels for filtering — `rollout`, `v2.3`, `maintenance`. |
| `source` | Where the annotation came from (`operator`, or the system). |

The only validation rule worth remembering: `end_time` may not be before
`start_time`.

---

## Where annotations show up

- **On the timeline**, next to the device and fleet events, so a firing alert
  sits right beside the change that may have caused it.
- **Alongside alerts.** The alert evaluator can drop an annotation for a
  transition, so an alert's firing moment is itself a marker on the timeline.
- **Filterable by tag and range**, so you can ask "show me everything tagged
  `rollout` in the last week."

---

## Managing annotations over REST

All routes require organization context.

| Method | Path | Does |
|:-------|:-----|:-----|
| `GET` | `/v1/annotations` | List annotations, filtered by time range and tag. |
| `POST` | `/v1/annotations` | Create one. |
| `GET` | `/v1/annotations/:id` | Read one. |
| `PATCH` | `/v1/annotations/:id` | Update one. |
| `DELETE` | `/v1/annotations/:id` | Delete one. |

Listing accepts a time window (start and end) and an optional tag. Reading needs
`annotations.read`; writing needs `annotations.write`, scoped to `org:*`.

---

## A worked example

A rollout goes out at 14:00 and you want the timeline to say so.

```json
POST /v1/annotations
{
  "title": "Firmware v2.3 rollout",
  "text": "Pushed to the warehouse group; watching for regressions.",
  "start_time": 1767225600000,
  "tags": ["rollout", "v2.3", "warehouse"]
}
```

Later, when a device failure alert fires at 14:20, the annotation is sitting
right there at 14:00 on the same timeline. You don't have to remember what
changed — the timeline remembers for you.

For a window rather than a moment, add `end_time`. The annotation then covers the
whole span, which is what you want for a maintenance period or a staged rollout
that runs over hours.
