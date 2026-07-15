## Routes Still Without Org Context

These routes are **intentionally without org context** because they use different authentication methods:

| Route Group | Reason |
|-------------|--------|
| `/admin/*` | Super admin only routes |
| `/device/*` (management) | Device-to-server HMAC/signing authentication |
| `/command/*` | Device-to-server HMAC/signing authentication |
| `/invite/*` | Public invite acceptance routes |
| `/me/*` | Operator's own data (not org-scoped) |
| `/auth/*` | Authentication routes |

---

## GraphQL Schema Files Needing Updates

**✅ Already Updated:**
- `inbox`, `inboxEntry` queries
- `deviceInspection`, `deviceTimeline` queries
- `ackInbox`, `deregisterDevice` mutations

**❌ Still Needs `organizationId` argument:**

| Query | File | Lines |
|-------|------|-------|
| `device` | schema.go | 103-109 |
| `devices` | schema.go | 111-118 |
| `deviceCount` | schema.go | 120-124 |
| `command` | schema.go | 130-136 |
| `pendingCommands` | schema.go | 138-145 |
| `telemetryHistory` | schema.go | 151-160 |
| `latestTelemetry` | schema.go | 162-168 |
| `telemetryStats` | schema.go | 170-177 |
| `connectionStatus` | schema.go | 183-189 |
| `allConnections` | schema.go | 191-195 |
| `deviceMetrics` | schema.go | 201-211 |
| `deviceLogs` | schema.go | 213-224 |
| `deviceCommandHistory` | schema.go | 226-237 |
| `dashboardStats` | schema.go | 239-243 |
| `updatesStatus` | schema.go | 249-255 |
| `updatesVersions` | schema.go | 257-265 |
| `updatesChangelog` | schema.go | 267-274 |
| `updatesHistory` | schema.go | 276-284 |
| `updatesHistoryDetail` | schema.go | 286-292 |
| `updatesSyncStatus` | schema.go | 294-298 |

**❓ Mutations needing review:**
- All device mutations (updateFCMToken, deregisterDevice, etc.)
- All command mutations (execute, cancel, retry)
- All updates mutations (push, sync, cancel)

---

**Would you like me to update the remaining GraphQL queries and mutations with `organizationId` arguments?**
