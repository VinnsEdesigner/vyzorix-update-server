# Config Versioning

Every time someone changes an organization's settings or thresholds, the server
keeps the previous version. Not a diff you have to reconstruct — a full snapshot,
with who made the change and when. You can list the history, look at any past
version, and put it back.

It's the difference between "settings look wrong" and "settings look wrong, and
here's exactly what changed at 09:14 on Tuesday, and here's the button to undo
it."

---

## What gets versioned

Two org-scoped config surfaces today:

| Resource | What it covers |
|:---------|:---------------|
| `org_settings` | The organization's settings blob. |
| `org_thresholds` | The organization's default alert thresholds. |

Each write to one of these appends a new version. Nothing is overwritten, so the
history is complete and ordered.

---

## What a version holds

| Field | Meaning |
|:------|:--------|
| `version` | Monotonic number, scoped to the resource. Starts at 1. |
| `snapshot` | The full config as it stood after the write. |
| `changed_by` | The operator who made the change. |
| `created_at` | When it happened. |
| `resource_type` | Which surface this version belongs to. |

Because the snapshot is the whole config, restoring is exact — there's no
replay of a chain of diffs, and no guessing what the result will be.

---

## Restoring a version

```
POST /v1/config-versions/:resource/:version/restore
```

Restoring applies that version's snapshot back onto the live config. The restore
is itself a write, so it produces a **new** version — the history never rewinds,
it only grows. You can always see that a restore happened, and who did it, and
undo the undo.

`resource` is the resource type (`org_settings` or `org_thresholds`); `version`
is the number you want back.

---

## Reading the history

| Method | Path | Does |
|:-------|:-----|:-----|
| `GET` | `/v1/config-versions/:resource` | List versions of a resource, newest first. |
| `GET` | `/v1/config-versions/:resource/:version` | Read one version. |
| `POST` | `/v1/config-versions/:resource/:version/restore` | Restore that version. |

Reading needs `config_versions.read`; restoring needs `config_versions.write`,
scoped to `org:*`.

---

## Why snapshots instead of diffs

A diff chain is fragile: to know what the config looked like at version 4, you
replay 1→2→3→4 and hope nothing was lost along the way. A snapshot is the answer
itself. It costs a little more to store and saves a lot of reasoning — and for
config that changes rarely, that's the right trade.

The write path is wired into the settings and thresholds services, so versions
are captured automatically. You don't call a "record a version" endpoint; you
change a setting, and the version appears.
