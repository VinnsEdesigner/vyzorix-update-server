# Organizations & Multi-Tenancy

The server is multi-tenant. Every resource (devices, commands, updates, API keys) belongs to an organization. Operators are global identities that belong to one or more organizations with scoped roles.

## The model

- **Operator** — a person identified by email + password (or OAuth). Global across organizations.
- **Organization** — a tenant. Owns devices, updates, API keys, telemetry.
- **Membership** — links an operator to an organization with a role.
- **Invitation** — the flow for joining an organization.

## Roles

Four roles, defined as a CHECK constraint in the `organization_members` table:

| Role | What they can do |
|------|-----------------|
| `super_admin` | Everything — manage members, delete org, manage all devices |
| `admin` | Manage devices, API keys, updates — cannot manage members |
| `operator` | Execute commands, view devices — cannot manage |
| `viewer` | Read-only access to dashboard and devices |

The `RequireSuperAdmin` middleware checks if the operator is a super_admin in the current org. The `RBACAuthorize` middleware does role-based checks on specific routes.

## Organization context middleware

The `NewOrganizationContext` middleware in `internal/api/middleware/org_context.go` extracts the org ID from the request. It checks:
- URL path parameters (e.g., `/:org/graphql`)
- Query parameters
- The session's `SelectedOrganizationID`

It sets the org ID in the gin context. If no org ID is found and the middleware isn't configured to skip missing orgs, it returns 400.

## Membership middleware

`NewOrganizationMembership` verifies the authenticated operator is an active member of the current organization. It takes a `MembershipChecker` that queries the `organization_members` table. If the operator isn't a member or is removed, it returns 403.

## Selecting an organization

`POST /v1/auth/organizations/select` sets the `SelectedOrganizationID` on the operator's session. This is how the dashboard switches between organizations.

## Invitations

Operators join organizations via invitations:

1. An admin creates an invitation: `POST /v1/organizations/:id/invitations` with the invitee's email and role
2. The invitee receives an email with an invitation link
3. The invitee accepts: `POST /v1/organizations/:id/invitations/:id/accept`
4. A membership is created with the specified role

Invitations can expire (configurable TTL) and be revoked. The invitation status is tracked: `pending` → `approved` / `rejected` / `expired`.

## Organization settings

Each org has settings stored in `organization_settings`:

- Webhook URL (for event notifications)
- Webhook secret (for signing webhook payloads)
- Notification preferences
- Password policy (min length, min age, history)

Managed via `GET/PATCH /v1/organizations/:id/settings`.

## Device ownership

Devices have an `OrganizationID` field. When a device registers, it's assigned to the organization of the registering operator (or the org specified in the registration request). All device queries are org-scoped — the `DevicesHandler.GetDevices` method filters by the org ID from the context.

## Transfer

Devices can be transferred between organizations: `POST /v1/organizations/:id/devices/:imei/transfer`. This changes the `OrganizationID` on the device. Only super_admins can transfer.

## API key scoping

API keys are scoped to an organization. The `TenantAPIKeyAuth` middleware sets both the operator ID and org ID in the context. API key scope (read/write/admin) determines what operations the key can perform within that org.
