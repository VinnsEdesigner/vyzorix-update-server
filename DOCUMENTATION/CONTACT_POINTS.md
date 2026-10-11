# Contact Points

A contact point is a named destination where notifications go. One place to
configure — email, a webhook, or a Slack channel — and every alert and system
message in the organization can be delivered there. Change the destination once
and everything that references it follows.

If [alerting](ALERTING.md) is the part that decides *when* to speak up, contact
points are the part that decides *where* the words land.

---

## What a contact point is

An org-scoped record: a name, a channel type, and the configuration that channel
needs. Every field is validated on save, so a broken destination fails at
creation time rather than silently at 3am when the alert fires.

| Field | Meaning |
|:------|:--------|
| `name` | Human label, unique within the org. |
| `channel` | `email`, `webhook`, or `slack`. |
| `config` | Channel-specific settings (see below). |
| `secret` | Optional shared secret, e.g. for signing or auth. Stored encrypted. |
| `template_id` | Optional template selector for the message body. |
| `enabled` | Disabled points are skipped by the dispatcher. |

### Channel types

**Email** — needs a `to` address.

```json
{ "channel": "email", "config": { "to": "oncall@example.com" } }
```

**Webhook** — needs a `url`. The URL must be HTTPS, or `http://localhost` for
local development; plain `http://` to a public host is rejected.

```json
{ "channel": "webhook", "config": { "url": "https://hooks.example.com/vyzorix" } }
```

**Slack** — needs a Slack incoming-webhook URL. The URL must start with
`https://hooks.slack.com/`.

```json
{ "channel": "slack", "config": { "webhook": "https://hooks.slack.com/services/…" } }
```

---

## How a message is delivered

When something needs to notify — a rule firing, a rule resolving — the
dispatcher takes the message and fans it out to every enabled contact point in
the organization.

For each point it:

1. Picks the channel implementation for the point's type.
2. Renders the message for that channel (Slack gets Block Kit, email gets
   subject plus body, webhooks get the raw JSON payload).
3. Sends it, with a timeout so one slow destination can't stall the rest.
4. **Records the attempt.** Every delivery — success or failure — is written to
   the delivery log with its status, channel, timestamp, and error text.

One destination failing does not stop the others. The dispatcher reports how many
points it reached, and each failure is visible in that point's delivery history.

### Templates

Slack messages are rendered from templates (`vyzorix_slack_templates.go`) so the
block layout stays consistent. The `template_id` on a contact point selects which
template is used; when it's empty, the channel's default is used.

---

## Testing a destination

Before you trust a contact point, send it a real message:

```
POST /v1/notifications/contact-points/:id/test
```

The server dispatches a test message through the point's channel and returns the
result — success, or the error the channel produced. This is the fastest way to
catch a typo'd webhook URL or a revoked Slack token.

---

## Delivery history

Every attempt is kept:

```
GET /v1/notifications/contact-points/:id/deliveries
```

Each entry carries the channel, the status, when it happened, and the error when
there was one. It's an audit trail for "did we actually notify anyone?" — useful
when a rule fired and nobody noticed.

---

## Managing contact points over REST

All routes live under `/v1/notifications/contact-points` and require organization
context.

| Method | Path | Does |
|:-------|:-----|:-----|
| `GET` | `/v1/notifications/contact-points` | List the org's points. |
| `POST` | `/v1/notifications/contact-points` | Create a point. |
| `GET` | `/v1/notifications/contact-points/:id` | Read one point. |
| `PATCH` | `/v1/notifications/contact-points/:id` | Update a point. |
| `DELETE` | `/v1/notifications/contact-points/:id` | Delete a point. |
| `POST` | `/v1/notifications/contact-points/:id/test` | Send a test message. |
| `GET` | `/v1/notifications/contact-points/:id/deliveries` | Read the delivery log. |

Reading needs `notifications.read`; writing and testing need
`notifications.write`, scoped to `org:*`.

---

## What delivers through a contact point

- **Alert transitions** — when a rule fires or resolves, its notification goes
  out through every enabled contact point in the org.
- **System messages** — any server-side event that needs to reach a human uses
  the same dispatcher, so there is one delivery path to reason about, one
  delivery log, and one place to configure destinations.

A rule can still carry its own `webhook_url` for a one-off destination, but the
intended shape is: configure contact points once, and let rules reference the
organization's set.
