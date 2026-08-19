# Authorization Model: Code-Verified Issues & Fixes

Code-level review of the vyzorix organization/authorization model, comparing
what the code actually does against a scoped-permission model. Each
issue includes the exact evidence (file + line) and the best fix.

Severity: CRITICAL (reachable security hole) > HIGH (missing protection) >
MEDIUM (design gap, no immediate hole).

---

## Issue 1 — GraphQL `sendCommand` bypasses the risk gate entirely (CRITICAL)

The REST command path (`POST /v1/device/:imei/command` → `command_execute.go`)
enforces the risk/confirmation/MFA system before dispatch: `authorizeCommand`
builds an `ActorContext` (operator, MFA-verified-from-session), runs
`RiskEvaluator.Evaluate`, and for high/critical-tier commands consumes a
single-use confirmation token. `device.factory_reset` requires MFA **and** a
confirmation token.

The GraphQL mutation does none of that.

`internal/api/graphql/resolver/mutation_resolver.go` — `SendCommand` (line 122):

```go
// Verify device exists in organization using org-scoped method.
_, err = r.DeviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
...
cmdResp, err := r.CommandService.SendCommand(ctx, cmdReq)   // <- straight to dispatch
```

It checks org membership (via middleware) and device-in-org, then calls
`CommandService.SendCommand` directly. `CommandService.SendCommand`
(`internal/application/command/command_service.go:37`) does device-exists +
pending-count + idempotency — **no risk evaluation, no confirmation token, no
MFA check.**

**Result:** any authenticated org member — including a `viewer` — can send
`device.factory_reset` (critical tier) over `POST /:org/graphql` with no
confirmation token and no MFA. The entire confirmation/MFA system built for the
REST path is bypassed on the GraphQL path.

**Best fix:** route GraphQL command mutations through the same authorization
gate as REST. The cleanest place is a single authorization function the command
service exposes (or a shared authorizer both layers call) so the REST handler
and the GraphQL resolver run identical checks: classify risk → require MFA for
critical → require/consume a confirmation token when the profile demands it.
Do not duplicate the logic in the resolver; call the shared gate.

---

## Issue 2 — Device transfer route has no authorization (HIGH)

`internal/api/server_routes.go:536`:

```go
orgs.POST("/:id/devices/:imei/transfer", s.transferHandler.Transfer)
```

This sits inside `setupOrganizationRoutes`, whose group is only
`orgs.Use(s.cookieAuth.Middleware())` (line 508). There is **no**
`OrganizationContext`, **no** `OrganizationMembership`, **no**
`RequireSuperAdmin`. The handler (`internal/api/handlers/device/
device_transfer.go`) checks only that the operator is authenticated, then calls
`TransferDevice`. `TransferDevice` (`device_service.go:980`) verifies the
device belongs to the source org and is offline — but **never checks the
operator's role in the source org, nor any membership in the target org.**

Contrast: the *member* endpoints self-enforce (`CheckCanManageMembers`,
super_admin check in `TransferOwnership`). Device transfer does not.

**Result:** any authenticated operator who knows a source org ID, a device IMEI,
and a target org ID can move a device out of an org — effectively removing it
from that org's control — regardless of their role, and into an org they may not
even belong to.

**Best fix:** require `super_admin` in the source org **and** membership in the
target org before a transfer is accepted. Add `OrganizationMembership` +
`RequireSuperAdmin` to the route, and validate target-org membership in the
service layer (defense in depth).

---

## Issue 3 — No per-device authorization within an org (HIGH)

Every device read and every command is scoped only to the org, not to the
operator or the device.

`internal/application/device/device_service.go:529` — `GetDevices`:

```go
allDevices, _, err := s.deviceRepo.ListByOrganizationPaginated(ctx, query.OrganizationID, 10000, 0)
```

Loads **all** devices in the org regardless of the caller. The repository has
`ListByOperator`, `ListActiveByOperator`, and `FindByIMEIAndOperator`
(`device_repository.go`), but **none are called on the read path.** Command
execution checks only `verifyDeviceInOrganization` — that the device is in the
org — never that the operator is authorized for *this* device.

**Result:** a `viewer` sees the same device list as a `super_admin`, and (once
Issue 1 is fixed) any operator-tier member can still send `factory_reset` to any
device in the org. The org is the only authorization boundary.

**Best fix:** introduce per-device authorization. Devices already carry an
`OperatorID` (the owning/registering operator, `device_storage.go:88`). Enforce
**owner-or-admin**: admin/super_admin can access any device in the org;
operator-tier members only devices they own (`device.OperatorID == op.ID`).
Apply it on the command-execution path first (the destructive surface), then the
read path.

---

## Issue 4 — Authorization is role-tier only, not scoped per resource (MEDIUM)

vyzorix: `RequireOrgRole(minRole)` (`rbac_authorize.go`) does
`membership.Role.Level() >= minRole.Level()` — a 4-tier numeric comparison.
`RequirePermission` doesn't check an independent grant: `HasPermission`
(`operator_entity.go:220`) re-derives permission from role via a hardcoded
switch, and `permissions.go` marks `DefaultPermissions()`/`AdminPermissions()`
`Deprecated` — "Permissions are now derived from organization membership roles."
The `Permission` constants are vestigial.

A scoped engine evaluates `EvalPermission(action, scopes...)` evaluates
`permissions[action]` against per-resource scopes with wildcard prefix matching
(`datasources:uid:abc` matches `datasources:uid:*`). Permissions are
independently grantable, persisted, and evaluated per resource at the route.

**Result:** vyzorix can't express "operator A can command device X but not
device Y" or "this API key can only push updates, not reboot." Authorization
granularity went *backwards* (a permission system was deleted for role bundles).

**Best fix:** long-term, move to action+scope permission evaluation like
Short-term, Issue 3's owner-or-admin device scoping is the highest-value
increment and does not require the full RBAC rewrite.

---

## Issue 5 — No teams / intra-org partitioning (MEDIUM)

No `Team` concept exists anywhere in vyzorix (`find internal/domain -iname
'*team*'` → empty). a mature system partitions org access via teams + scoped
folder/dashboard permissions. vyzorix has operators ↔ orgs and nothing between.

**Result:** no way to say "field-ops team manages these 200 devices, NOC manages
these 50." Every device is accessible to every member.

**Best fix:** future feature — add a team/device-group entity scoped within an
org, and let permissions target a group. Defer until Issue 4's scoped model
exists; teams without scoped permissions would just be another coarse role.

---

## Issue 6 — Org context is client-header-driven per request (MEDIUM, not a hole)

`internal/api/middleware/org_context.go` reads the active org from
`?organization_id=` query → `X-Organization-ID` header → session. So a client
can shift org context **on every request** by sending a different header. The
`OrganizationMembership` middleware then validates active membership in whatever
org was claimed (`GetMembership` returns `ErrNotOrgMember` when
`!member.IsActive()`) — so this is **not** a cross-org access hole.

A server-driven model treats org ID as a server-side property of the signed-in identity: a
`?orgId=` that differs from the session **forces re-login** (`middleware/
auth.go:208`), and org-switching is an explicit validated `POST /user/using/:id`
(`api/user.go:499`, `validateUsingOrg`).

**Result:** in vyzorix, "which org was the operator acting in" is whatever the
client claimed that request — auditable via the membership check, but not a
stable session property. Audit accuracy depends on correctly logging the
resolved org every time.

**Best fix:** keep the header (it's useful for multi-org API clients) but treat
the resolved orgID as authoritative context and log it on every mutation; for
browser sessions, prefer the session's `SelectedOrganizationID` and validate a
header that disagrees.

---

## Issue 7 — No role gate on the command execution path (MEDIUM)

Neither the REST command route (`deviceMgmt`, line 341) nor the GraphQL
`sendCommand` applies `RequireOrgRole`. A `viewer` (read-only intent) reaches
the command handler. The REST path's risk gate is the only thing that might stop
a destructive command — and the GraphQL path (Issue 1) has no risk gate at all.

**Best fix:** gate command execution on at least operator-tier role
(`RequireOrgRole(RoleOperator)`), in addition to the per-device owner-or-admin
check from Issue 3.

---

## Fix priority

1. **Issue 1** — GraphQL risk-gate bypass. Critical, reachable, defeats the
   confirmation/MFA system. Fix first.
2. **Issue 2** — transfer authorization. Any member can remove a device from an
   org. Close the hole.
3. **Issue 3** — per-device owner-or-admin authorization. The core MDM gap.
4. **Issue 7** — role gate on command path. Cheap, closes the viewer-can-command gap.
5. **Issues 4–6** — architectural (scoped permissions, teams, session-bound org).
   Documented, phased; don't big-bang them.

## Implementation status

Implemented and verified (build + vet + golangci-lint + full test suite all pass):

- **Issue 1** — Shared `command.Authorizer` (single gate for REST + GraphQL).
  GraphQL `sendCommand` runs the same risk/MFA/confirmation gate as REST; added
  `confirmationToken` schema arg and bridged session MFA into the GraphQL context.
- **Issue 2** — `POST /v1/organizations/:id/devices/:imei/transfer` now requires
  super_admin in the source org AND active membership in the target org; the dead
  `RegisterRoutes` method was removed.
- **Issue 3** — Per-device owner-or-admin on the command path: non-admins may
  only command devices they own (`device.OperatorID`) or that are assigned to a
  group they belong to (device groups).
- **Issue 7** — Role gate: a viewer-tier member cannot execute commands (enforced
  via the scoped permission engine, not just a role-level check).
- **Issue 4** — Scoped permission engine (`internal/domain/permission`): Action +
  Scope with trailing-wildcard matching, role→scoped defaults, a persisted
  `permission_grants` table (migration 60) for admin-assigned custom scopes, and
  an Evaluator that unions defaults + grants. `operator.HasPermission` and the
  command authorization helper evaluate scoped permissions.
- **Issue 5** — Device groups: `device_groups`, `device_group_members`,
  `device_group_devices` tables (migration 61) with `storage.DeviceGroupRepository`
  and real in-memory SQLite integration tests; the command authorization helper
  grants access to group members.
- **Issue 6** — Server-driven org state: `api_keys.organization_id` is bound at
  key creation; the tenant API-key middleware sets org context from the key, and
  the org-context middleware treats session/key binding as authoritative,
  rejecting a conflicting client header/query with 400 (mirrors a
  validated org-switch, not a per-request header override).
