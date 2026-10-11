# Live Channels

A live channel is a named stream of events scoped to one organization. A
dashboard subscribes to the channels it cares about, and the server pushes events
the moment they happen — an alert firing, a command changing status, a member
joining, a device's telemetry arriving.

It's the addressing and permission layer on top of the WebSocket hub. The hub
moves bytes; channels decide who is allowed to hear them.

---

## The address

Every channel has a three-part address:

```
stream/<org>/<scope>
```

- `stream` — the fixed prefix.
- `<org>` — the organization ID. **This is the security boundary.** A channel
  belongs to exactly one org.
- `<scope>` — what the channel carries. Either a named scope or a sub-scope.

### Scopes

| Scope | Carries |
|:------|:--------|
| `alerts` | Alert state transitions (firing, resolved). |
| `commands` | Command dispatch and status changes. |
| `members` | Organization membership changes. |
| `telemetry` | Device telemetry frames. |
| `telemetry/device/<id>` | Telemetry for one device. |

The org prefix is not decoration. A subscriber can never receive events from
another organization's channel, because the address can't reach it.

---

## Subscribing is gated

A subscribe attempt is checked before the stream opens. Two authorizers exist:

- **Membership** — the subject must be a member of the channel's organization.
- **Permission** — membership *plus* the RBAC action for the scope. This is the
  stricter gate, and it's the one used for scope-level channels.

If the check fails, the subscribe is refused and no stream is opened. The
permission gate means an operator with telemetry access can't subscribe to the
commands channel just because they're in the same org.

---

## Streams open and close with their subscribers

A channel's stream has a lifecycle tied to its audience:

- The **first** subscriber opens the stream.
- More subscribers attach to the same stream.
- The **last** subscriber leaving closes it.

This means the server holds no resources for a channel nobody is listening to,
and it doesn't need a reaper — the audience is the signal.

---

## Managing subscriptions over REST

| Method | Path | Does |
|:-------|:-----|:-----|
| `GET` | `/v1/channels/status` | Current channel/connection status. |
| `POST` | `/v1/channels/subscribe` | Subscribe to a channel address. |
| `POST` | `/v1/channels/unsubscribe` | Leave a channel. |

All require organization context.

---

## What publishes to channels

- **The alert evaluator** publishes transitions to `stream/<org>/alerts`. When a
  rule fires or resolves, every subscriber sees it immediately — the dashboard
  doesn't poll for alert state.
- **Command and membership events** publish to their scopes as they happen.
- **Telemetry** flows on the device-scoped sub-channels, so a dashboard watching
  one device only receives that device's frames.

This is what replaced broadcast-to-all. Instead of every connected client
receiving every event and filtering locally, the server publishes to a named
channel and the hub fans out only to that channel's subscribers.

---

## Why this shape

Two design choices carry the weight here:

**The org is in the address.** Authorization and routing are the same decision.
There's no separate "can this client see this event?" check at delivery time,
because a client can only ever be subscribed to channels in its own org.

**Streams are audience-driven.** The server never decides when a channel should
exist. It exists while someone is listening and stops when they stop. That keeps
the resource model honest: no idle streams, no reaping, no configuration to keep
in sync with reality.
