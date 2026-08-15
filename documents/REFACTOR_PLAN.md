# Server Refactoring Plan: Multi-Tenant Organization Model

## Executive Summary

Transform the flat role-based system into a multi-tenant organization model where:
- **Operators** are global identities (email, password, OAuth)
- **Organizations** are tenants that own resources
- **Memberships** link operators to organizations with scoped roles
- **Invitations** manage the join request flow

---

## Phase 1: Database Schema Changes

### 1.1 New Tables Required

```sql
-- organizations table
CREATE TABLE organizations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_by TEXT NOT NULL,  -- FK to operators.id (first super_admin)
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    deleted_at INTEGER,  -- soft delete
    is_active BOOLEAN DEFAULT true,
    max_members INTEGER DEFAULT 2,  -- limit to 2 active orgs per operator
    FOREIGN KEY (created_by) REFERENCES operators(id)
);

-- organization_members table (replaces global role)
CREATE TABLE organization_members (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    operator_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('super_admin', 'admin', 'operator', 'viewer')),
    invited_by TEXT,  -- FK to operators.id
    joined_at INTEGER NOT NULL,
    removed_at INTEGER,  -- soft delete
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'removed')),
    UNIQUE(organization_id, operator_id),
    FOREIGN KEY (organization_id) REFERENCES organizations(id),
    FOREIGN KEY (operator_id) REFERENCES operators(id)
);

-- invitations table
CREATE TABLE invitations (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    email TEXT NOT NULL,  -- invitee email
    role TEXT NOT NULL CHECK (role IN ('admin', 'operator', 'viewer')),
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'expired')),
    token TEXT NOT NULL UNIQUE,  -- secure token for invite link
    inviter_notes TEXT,  -- notes from inviter
    invitee_notes TEXT,  -- notes from invitee on accept/reject
    invited_by TEXT NOT NULL,  -- FK to operators.id
    invited_at INTEGER NOT NULL,
    responded_at INTEGER,
    expires_at INTEGER NOT NULL,  -- token expiry
    FOREIGN KEY (organization_id) REFERENCES organizations(id),
    FOREIGN KEY (invited_by) REFERENCES operators(id)
);

-- Indexes for performance
CREATE INDEX idx_org_members_operator ON organization_members(operator_id);
CREATE INDEX idx_org_members_org ON organization_members(organization_id);
CREATE INDEX idx_invitations_token ON invitations(token);
CREATE INDEX idx_invitations_email ON invitations(email);
CREATE INDEX idx_invitations_org_status ON invitations(organization_id, status);
```

### 1.2 Table Modifications

| Table | Change | Action |
|-------|--------|--------|
| `operators` | Remove `role` column | Migration drops column |
| `devices` | Add `organization_id` column | Migration adds nullable FK |
| `sessions` | Add `organization_id` column | Migration adds nullable FK |
| `api_keys` | Add `organization_id` column | Migration adds nullable FK |

### 1.3 Migration File: `040_organizations.sql`

```sql
-- Migration: Add organizations, memberships, invitations
-- This migration:
-- 1. Creates new tables
-- 2. Migrates existing global role to null (no org)
-- 3. Adds organization_id to devices, sessions, api_keys

-- Create organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_by TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    deleted_at INTEGER,
    is_active INTEGER DEFAULT 1,
    max_members INTEGER DEFAULT 2
);

-- Create organization_members table
CREATE TABLE IF NOT EXISTS organization_members (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    operator_id TEXT NOT NULL,
    role TEXT NOT NULL,
    invited_by TEXT,
    joined_at INTEGER NOT NULL,
    removed_at INTEGER,
    status TEXT DEFAULT 'active',
    UNIQUE(organization_id, operator_id)
);

-- Create invitations table
CREATE TABLE IF NOT EXISTS invitations (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    email TEXT NOT NULL,
    role TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    token TEXT NOT NULL UNIQUE,
    inviter_notes TEXT,
    invitee_notes TEXT,
    invited_by TEXT NOT NULL,
    invited_at INTEGER NOT NULL,
    responded_at INTEGER,
    expires_at INTEGER NOT NULL
);

-- Add organization_id to devices (nullable for now, required after migration)
ALTER TABLE devices ADD COLUMN organization_id TEXT;

-- Add organization_id to sessions (nullable for backward compat)
ALTER TABLE sessions ADD COLUMN organization_id TEXT;

-- Add organization_id to api_keys (nullable for backward compat)
ALTER TABLE api_keys ADD COLUMN organization_id TEXT;

-- Note: Global 'role' column in operators table will be deprecated
-- Use organization_members.role instead for permission checks
```

---

## Phase 2: Domain Layer Changes

### 2.1 New Domain Packages

```
internal/domain/
├── organization/           # NEW
│   ├── organization_entity.go
│   ├── member_entity.go
│   ├── invitation_entity.go
│   ├── organization_errors.go
│   ├── organization_repository.go
│   ├── member_repository.go
│   └── invitation_repository.go
```

### 2.2 Files to MODIFY

| File | Changes |
|------|---------|
| `domain/operator/operator_entity.go` | Remove `Role` field, add `Memberships` field |
| `domain/operator/role.go` | Add org-scoped role methods |
| `domain/operator/permissions.go` | Remove global permissions, add membership-based checks |
| `domain/device/device_entity.go` | Add `OrganizationID` field |

### 2.3 Files to CREATE

| File | Purpose |
|------|---------|
| `domain/organization/organization_entity.go` | Organization entity with validation |
| `domain/organization/member_entity.go` | OrganizationMember entity |
| `domain/organization/invitation_entity.go` | Invitation entity with token generation |
| `domain/organization/organization_errors.go` | Domain-specific errors |
| `domain/organization/organization_repository.go` | Repository interface |
| `domain/organization/member_repository.go` | Member repository interface |
| `domain/organization/invitation_repository.go` | Invitation repository interface |

---

## Phase 3: Infrastructure Layer Changes

### 3.1 Storage Files to CREATE

| File | Purpose |
|------|---------|
| `storage/040_organizations.go` | Migration for new tables |
| `storage/organization_storage.go` | OrganizationRepository implementation |
| `storage/member_storage.go` | MemberRepository implementation |
| `storage/invitation_storage.go` | InvitationRepository implementation |

### 3.2 Storage Files to MODIFY

| File | Changes |
|------|---------|
| `storage/device_storage.go` | Add organization_id to queries |
| `storage/session_storage.go` | Track org context in sessions |
| `storage/operator_storage.go` | Remove global role queries |

### 3.3 Email Templates to CREATE

| File | Purpose |
|------|---------|
| `email/templates/invitation.go` | Invitation received email |
| `email/templates/invitation_accepted.go` | Notify inviter of acceptance |
| `email/templates/invitation_rejected.go` | Notify inviter of rejection |

---

## Phase 4: Application Layer Changes

### 4.1 New Application Services

```
internal/application/
├── organization/           # NEW
│   ├── organization_service.go
│   ├── member_service.go
│   └── invitation_service.go
```

### 4.2 Application Files to MODIFY

| File | Changes |
|------|---------|
| `application/auth/auth_register.go` | Create operator without role |
| `application/auth/auth_google_oauth.go` | No auto-role assignment |
| `application/auth/auth_github_oauth.go` | No auto-role assignment |

### 4.3 Application Files to CREATE

| File | Purpose |
|------|---------|
| `application/organization/organization_service.go` | Org CRUD operations |
| `application/organization/member_service.go` | Member management |
| `application/organization/invitation_service.go` | Invitation flow |

---

## Phase 5: Handler Layer Changes

### 5.1 New Handler Directories

```
internal/api/handlers/
├── organization/           # NEW
│   ├── organization_handler.go
│   ├── member_handler.go
│   └── invitation_handler.go
```

### 5.2 Handler Files to MODIFY

| File | Changes |
|------|---------|
| `handlers/auth/auth_register.go` | No role assignment on register |
| `handlers/auth/auth_admin.go` | Refactor for org-scoped admin operations |
| `handlers/device/device_register.go` | Require org_id for registration |
| `handlers/device/device_list.go` | Filter by org context |

### 5.3 Handler Files to CREATE

| File | Purpose |
|------|---------|
| `handlers/organization/organization_handler.go` | POST/GET/PATCH/DELETE /v1/organizations |
| `handlers/organization/member_handler.go` | /v1/organizations/:id/members |
| `handlers/organization/invitation_handler.go` | /v1/invitations and /v1/invite/:token |

---

## Phase 6: API Routes Changes

### 6.1 New Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/v1/organizations` | Create organization |
| GET | `/v1/organizations` | List my organizations |
| GET | `/v1/organizations/:id` | Get organization details |
| PATCH | `/v1/organizations/:id` | Update organization |
| DELETE | `/v1/organizations/:id` | Delete organization |
| GET | `/v1/organizations/:id/members` | List members |
| POST | `/v1/organizations/:id/members` | Add member (from invite) |
| DELETE | `/v1/organizations/:id/members/:memberId` | Remove member |
| PATCH | `/v1/organizations/:id/members/:memberId` | Change role |
| POST | `/v1/invitations` | Create invitation (send email) |
| GET | `/v1/invitations` | List pending invitations (as inviter) |
| GET | `/v1/invite/:token` | Get invitation details (public) |
| POST | `/v1/invite/:token/approve` | Accept invitation |
| POST | `/v1/invite/:token/reject` | Reject invitation |
| GET | `/v1/me/memberships` | List my organization memberships |
| GET | `/v1/me/organizations` | Quick list of my orgs |

### 6.2 Modified Routes

| Endpoint | Change |
|----------|--------|
| `/v1/dashboard/devices` | Filter by org_id from session context |
| `/v1/devices` | Filter by org_id |
| `/v1/device/register` | Require org_id in payload |

### 6.3 Deprecated Routes

| Endpoint | Replacement |
|----------|-------------|
| `/v1/auth/admin/operators` | `/v1/organizations/:id/members` |

---

## Phase 7: Middleware Changes

### 7.1 New Middleware

| File | Purpose |
|------|---------|
| `middleware/org_context.go` | Extract and validate org_id from request/session |
| `middleware/org_membership.go` | Verify operator is member of target org |

### 7.2 Modified Middleware

| File | Changes |
|------|---------|
| `middleware/super_admin.go` | Refactor to check org membership + super_admin role in that org |
| `middleware/rbac_authorize.go` | Add org-scoped permission checks |
| `middleware/cookie_auth.go` | Load operator with memberships |

### 7.3 Middleware Flow

```
Request
  ↓
cookie_auth → Load Operator + Memberships
  ↓
org_context → Extract org_id from URL/body/header
  ↓
org_membership → Verify operator is member of that org
  ↓
org_role_check → Verify role level for action
  ↓
Handler
```

---

## Phase 8: Client-Side Changes (Brief)

### 8.1 New API Client Directories

```
packages/API_Client/src/
├── domain/
│   ├── organization/           # NEW
│   │   ├── organization-entity.ts
│   │   ├── organization-mappers.ts
│   │   ├── member-entity.ts
│   │   └── invitation-entity.ts
│   └── invitation/             # NEW
│       └── invitation-entity.ts
└── vyzorServer/
    └── rest/
        ├── organization/       # NEW
        │   ├── organization-endpoints.ts
        │   └── member-endpoints.ts
        └── invitation/        # NEW
            └── invitation-endpoints.ts
```

---

## Summary of All Changes

### Files to CREATE (New)

```
apps/api/internal/
├── domain/organization/
│   ├── organization_entity.go
│   ├── member_entity.go
│   ├── invitation_entity.go
│   ├── organization_errors.go
│   ├── organization_repository.go
│   ├── member_repository.go
│   └── invitation_repository.go
├── application/organization/
│   ├── organization_service.go
│   ├── member_service.go
│   └── invitation_service.go
├── handlers/organization/
│   ├── organization_handler.go
│   ├── member_handler.go
│   └── invitation_handler.go
├── infrastructure/storage/
│   ├── 040_organizations.go
│   ├── organization_storage.go
│   ├── member_storage.go
│   └── invitation_storage.go
└── infrastructure/email/templates/
    ├── invitation.go
    ├── invitation_accepted.go
    └── invitation_rejected.go
```

### Files to MODIFY

```
apps/api/internal/
├── domain/
│   ├── operator/
│   │   ├── operator_entity.go      # Remove Role, add Memberships
│   │   ├── role.go                # Add org-scoped methods
│   │   └── permissions.go         # Remove global, add org-scoped
│   └── device/
│       └── device_entity.go        # Add OrganizationID
├── application/auth/
│   ├── auth_register.go           # No role on register
│   ├── auth_google_oauth.go      # No auto-role
│   └── auth_github_oauth.go      # No auto-role
├── handlers/
│   ├── auth/auth_admin.go         # Refactor for org-scoped
│   ├── device/device_register.go  # Require org_id
│   └── device/device_list.go     # Filter by org
├── api/
│   ├── server_routes.go          # Add new routes
│   └── middleware/
│       ├── super_admin.go        # Org-scoped
│       ├── rbac_authorize.go     # Org-scoped
│       ├── cookie_auth.go        # Load memberships
│       └── [NEW] org_context.go
│       └── [NEW] org_membership.go
└── infrastructure/storage/
    ├── device_storage.go          # Add org_id queries
    └── session_storage.go        # Track org context
```

### Files to DELETE

```
apps/api/internal/
├── handlers/auth/auth_admin.go    # Replaced by org handlers
```

### Migration Files

```
apps/api/internal/infrastructure/storage/
├── 040_organizations.sql         # Create tables
└── 041_device_org_relations.sql  # Add org_id to devices
```

---

## Implementation Order

```
1. Migration 040_organizations.sql
   - Create organizations table
   - Create organization_members table
   - Create invitations table
   - Add organization_id to devices (nullable)
   - Add organization_id to sessions (nullable)

2. Domain Layer
   - Create organization package
   - Create invitation package
   - Modify operator entity (remove Role)

3. Storage Layer
   - Create organization_storage.go
   - Create member_storage.go
   - Create invitation_storage.go
   - Modify device_storage.go
   - Modify session_storage.go

4. Application Layer
   - Create organization_service.go
   - Create member_service.go
   - Create invitation_service.go
   - Modify auth services (remove auto-role)

5. Handler Layer
   - Create organization_handler.go
   - Create member_handler.go
   - Create invitation_handler.go
   - Modify device handlers

6. Middleware Layer
   - Create org_context.go
   - Create org_membership.go
   - Modify super_admin.go
   - Modify cookie_auth.go

7. Routes
   - Add organization routes
   - Add invitation routes
   - Add member routes

8. Email Templates
   - Create invitation.go
   - Create invitation_accepted.go
   - Create invitation_rejected.go

9. Testing
   - Unit tests for new services
   - Integration tests for new routes
   - Migration rollback tests
```

---

## Key Design Decisions Documented

1. **Max 2 Active Organizations**: Capped at 2 per operator as specified. Only ACTIVE orgs count toward max.
2. **Invitation Required for Join**: Users cannot join by org_id alone
3. **Org-Context from Session**: Selected org stored in session, not JWT
4. **Soft Delete**: Organizations and members use soft delete
5. **Token-Based Invitations**: Secure random tokens, 256-bit entropy
6. **Email Required**: Invitation emails require verified email
7. **Device Transfer Required**: Must delete device before registering to new org. **NEW: Device transfer feature to be implemented.**
8. **Invitation TTL**: 7 days default, configurable
9. **Rate Limit**: 5 invitations per hour per org
10. **Pagination**: Cursor-based for all list endpoints

---

## ️ CRITICAL GAPS IDENTIFIED & ADDRESSED

### 1. Race Condition: Max 2 Orgs Check

**Problem**: Two simultaneous requests could both pass the "max 2 orgs" check.

**Solution**: Use database transaction with row-level locking:

```go
func (s *OrganizationService) CreateOrganization(ctx context.Context, opID, name string, role OrgRole) (*Organization, error) {
    return s.db.Transaction(ctx, func(tx *sql.Tx) (*Organization, error) {
        // Count with FOR UPDATE lock prevents race
        count, err := s.countActiveOrgsForOperatorTx(ctx, tx, opID)
        if err != nil {
            return nil, err
        }
        if count >= MaxActiveOrgs {
            return nil, ErrMaxOrgsReached
        }
        // Proceed with creation
        return s.createOrgTx(ctx, tx, opID, name, role)
    })
}
```

### 2. Token Security: Invitation Token Generation

**Problem**: Are tokens cryptographically secure? Can they be guessed?

**Solution**: Use crypto/rand with 32 bytes (256 bits):

```go
func GenerateInvitationToken() (string, error) {
    bytes := make([]byte, 32)  // 256 bits - unguessable
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}

// Token entropy: 32 bytes = 256 bits
// Collision probability: negligible (1 in 2^256)
// Time to guess (10B attempts/sec): longer than universe exists
```

### 3. Email Enumeration Prevention

**Problem**: Can attackers enumerate valid invitation tokens?

**Solution**: 
- Tokens are high-entropy (256 bits) - unguessable
- Rate limit invitation creation (max 5/hour per org)
- Return same error for "invalid token" vs "expired token"
- Log failed token attempts for abuse detection

```go
// Never reveal if token exists
func (s *InvitationService) GetInvitationByToken(ctx context.Context, token string) (*Invitation, error) {
    inv, err := s.repo.FindByToken(ctx, token)
    if err == ErrNotFound {
        // Return same error for security
        return nil, ErrInvitationNotFound
    }
    return inv, err
}
```

### 4. Invitation Token TTL & Expiration

**Problem**: How long should tokens be valid?

**Decision**: 
- **Default TTL**: 7 days (configurable)
- **Max TTL**: 30 days
- **Expired tokens**: Status changed to "expired", cannot be used

```go
const (
    InvitationDefaultTTL = 7 * 24 * time.Hour  // 7 days
    InvitationMaxTTL    = 30 * 24 * time.Hour // 30 days
)
```

### 5. Concurrent Invitation Acceptance

**Problem**: Same token accepted twice simultaneously.

**Solution**: Use database unique constraint + atomic status update:

```go
func (s *InvitationService) ApproveInvitation(ctx context.Context, token string, inviteeID string, notes string) error {
    // Atomic update with status check
    result, err := s.db.ExecContext(ctx, `
        UPDATE invitations 
        SET status = 'approved', invitee_notes = ?, responded_at = ?, responder_id = ?
        WHERE token = ? AND status = 'pending' AND expires_at > ?
    `, notes, time.Now(), inviteeID, token, time.Now())
    
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return ErrInvitationAlreadyProcessed
    }
    return nil
}
```

### 6. Org Deletion with Pending Invitations

**Problem**: What happens to invitations when org is deleted?

**Solution**: 
- Delete ALL pending invitations when org is deleted (cascade)
- Send cancellation email to pending invitees
- Log audit event

```go
func (s *OrganizationService) DeleteOrganization(ctx context.Context, orgID string) error {
    return s.db.Transaction(ctx, func(tx *sql.Tx) error {
        // Cancel all pending invitations first
        if err := s.cancelPendingInvitationsTx(ctx, tx, orgID); err != nil {
            return err
        }
        // Soft delete org
        return s.deleteOrgTx(ctx, tx, orgID)
    })
}
```

### 7. Device Transfer Between Orgs

**Problem**: Can devices be moved between orgs?

**Decision**: 
- **NO device transfer** - devices are permanently bound to org at registration
- Device deletion required before registering to new org
- Or: Use "device claiming" flow (separate feature)

```go
// In device registration
func (s *DeviceService) RegisterDevice(ctx context.Context, req *RegisterRequest, orgID string) (*Device, error) {
    // Check device not already registered
    existing, _ := s.repo.FindByIMEI(ctx, req.IMEI)
    if existing != nil && existing.OrganizationID != "" {
        return nil, ErrDeviceAlreadyRegistered
    }
    // Proceed with registration to org
}
```

### 8. Session Invalidation on Role Change

**Problem**: User demoted while having active session.

**Solution**:
- **Do NOT invalidate sessions** on role change (too disruptive)
- Check role on EVERY request via middleware
- Sessions store org_context, role checked at request time

```go
// Middleware checks role on EVERY request
func OrgMembershipMiddleware(ctx context.Context, op *Operator, orgID string) error {
    membership := op.GetMembership(orgID)
    if membership == nil {
        return ErrNotOrgMember
    }
    // Role checked here - no session invalidation needed
    if !membership.Role.CanPerform(action) {
        return ErrInsufficientPermissions
    }
    return nil
}
```

### 9. Operator Account Deletion

**Problem**: What happens to org memberships when operator deletes account?

**Solution**:
- **Cascade delete memberships** when operator deleted
- **Cancel pending invitations** sent by operator
- **Delete operator** - cascades handle rest
- Log audit event

```go
func (s *OperatorService) DeleteOperator(ctx context.Context, operatorID string) error {
    return s.db.Transaction(ctx, func(tx *sql.Tx) error {
        // Remove from all orgs (cascade)
        if err := s.removeFromAllOrgsTx(ctx, tx, operatorID); err != nil {
            return err
        }
        // Cancel invitations sent by this operator
        if err := s.cancelSentInvitationsTx(ctx, tx, operatorID); err != nil {
            return err
        }
        // Delete operator (other cascades handle devices, sessions, etc)
        return s.deleteOperatorTx(ctx, tx, operatorID)
    })
}
```

### 10. Pagination on List Endpoints

**Problem**: What if org has 1000 members?

**Solution**: Cursor-based pagination (efficient for large datasets)

```go
// GET /v1/organizations/:id/members?limit=20&cursor=abc123
type ListMembersRequest struct {
    OrgID   string `param:"id"`
    Limit   int    `query:"limit"`   // default 20, max 100
    Cursor  string `query:"cursor"`   // opaque cursor
}

type ListMembersResponse struct {
    Members    []Member `json:"members"`
    NextCursor string   `json:"nextCursor,omitempty"`
    HasMore    bool     `json:"hasMore"`
}
```

### 11. Org Context Persistence

**Problem**: How does selected org persist?

**Solution**:
- Store `selected_org_id` in session cookie
- Default to first org if only one membership
- Frontend can override via `X-Org-ID` header
- Session refreshed on org switch

```go
// Session stores selected org
type Session struct {
    OperatorID    string
    SelectedOrgID string  // Current org context
    // ...
}

// Middleware extracts org from header or session
func OrgContextMiddleware(c *gin.Context) {
    orgID := c.GetHeader("X-Org-ID")
    if orgID == "" {
        orgID = c.Session.Get("selected_org_id")
    }
    // Validate membership, set in context
}
```

### 12. Graceful Degradation During Migration

**Problem**: How to deploy migration without breaking existing users?

**Solution**: Multi-phase rollout:

| Phase | Migration | Behavior |
|-------|-----------|----------|
| Phase 1 | Add nullable columns | All existing users work (null org_id) |
| Phase 2 | Backfill existing data | Assign devices to "default org" |
| Phase 3 | Require org_id for NEW operations | New registrations need org |
| Phase 4 | Strict mode | All operations require org_id |

```go
// Config flag for migration phases
type Config struct {
    RequireOrganizationID bool // Phase 3+
    StrictOrgMode         bool // Phase 4
}

// Middleware checks based on phase
func RequireOrgMiddleware(c *gin.Context) {
    if !c.Config.RequireOrganizationID {
        return // Allow null during migration
    }
    // Strict enforcement
}
```

### 13. Invitation Rate Limiting

**Problem**: Can someone spam invitations?

**Solution**:
- **Per org**: Max 20 pending invitations
- **Per operator**: Max 5 invitations per hour
- **Per email**: No duplicate pending invitations

```go
type InvitationLimits struct {
    MaxPendingPerOrg    = 20
    MaxPerOperatorPerHour = 5
    RequireVerifiedEmail = true
}

func (s *InvitationService) CreateInvitation(ctx context.Context, req *CreateInvitationRequest) error {
    // Check per-org limit
    count, _ := s.countPendingInvitations(ctx, req.OrgID)
    if count >= InvitationLimits.MaxPendingPerOrg {
        return ErrMaxInvitationsReached
    }
    
    // Check per-operator rate limit (redis sliding window)
    if err := s.rateLimiter.Check(ctx, "invite:"+req.InvitedBy, 5, time.Hour); err != nil {
        return ErrRateLimitExceeded
    }
    
    // Check no duplicate pending
    existing, _ := s.findPendingByEmail(ctx, req.OrgID, req.Email)
    if existing != nil {
        return ErrInvitationAlreadyPending
    }
    
    // Create invitation
}
```

### 14. Audit Logging Requirements

**Problem**: Need audit trail for compliance.

**Solution**: Comprehensive event logging:

```go
type AuditEvent struct {
    ID          string    `json:"id"`
    Timestamp   time.Time `json:"timestamp"`
    ActorID     string    `json:"actor_id"`      // Who did it
    Action      string    `json:"action"`        // What happened
    Resource    string    `json:"resource"`      // org, member, invitation
    ResourceID  string    `json:"resource_id"`
    OrgID       string    `json:"org_id"`
    Details     map[string]any `json:"details"`  // Extra context
    IPAddress   string    `json:"ip_address"`
}

// Events to log:
const (
    AuditOrgCreated           = "org.created"
    AuditOrgDeleted           = "org.deleted"
    AuditMemberInvited        = "member.invited"
    AuditMemberJoined         = "member.joined"
    AuditMemberRemoved        = "member.removed"
    AuditMemberRoleChanged    = "member.role_changed"
    AuditInvitationAccepted   = "invitation.accepted"
    AuditInvitationRejected   = "invitation.rejected"
    AuditInvitationExpired    = "invitation.expired"
)
```

### 15. Emergency Recovery Procedures

**Problem**: What if all admins are locked out?

**Solution**:
- **Bootstrap token**: One-time use secret set via env var
- **CLI command**: `vyzorix-cli admin recover --bootstrap-token=xxx`
- **Procedure**: Creates emergency admin membership

```bash
# Set bootstrap token in environment
VYORIX_BOOTSTRAP_TOKEN="one-time-secret-token-here"

# Run recovery
vyzorix-cli admin recover --email=admin@company.com --role=super_admin
```

### 16. Email Failure Handling

**Problem**: What if email sending fails after creating invitation?

**Solution**:
- **Transaction**: Create invitation, THEN send email
- **Retry queue**: Failed emails queued for retry (max 3 attempts)
- **Manual fallback**: Admin can resend invitation

```go
func (s *InvitationService) CreateInvitation(ctx context.Context, req *CreateInvitationRequest) error {
    // 1. Create invitation in DB (committed)
    inv, err := s.repo.Create(ctx, req)
    if err != nil {
        return err
    }
    
    // 2. Send email (non-transactional)
    if err := s.emailService.SendInvitation(ctx, inv); err != nil {
        // Log error but don't rollback - invitation exists
        s.logger.Error("failed to send invitation email", "invitation_id", inv.ID, "err", err)
        
        // Queue for retry
        s.retryQueue.Add(EmailJob{
            Type:      "invitation",
            InvitationID: inv.ID,
            Attempts:  0,
        })
    }
    
    return nil
}
```

---

## Complete Error Codes

```go
// Organization errors
var (
    ErrOrgNotFound          = errors.New("organization not found")
    ErrOrgDeleted          = errors.New("organization has been deleted")
    ErrMaxOrgsReached      = errors.New("maximum 2 active organizations allowed")
    ErrOrgNameRequired     = errors.New("organization name is required")
    ErrOrgNameTooLong      = errors.New("organization name exceeds 255 characters")
    ErrCannotDeleteLastOrg = errors.New("cannot delete last organization")
)

// Membership errors
var (
    ErrNotOrgMember        = errors.New("operator is not a member of this organization")
    ErrAlreadyOrgMember    = errors.New("operator is already a member of this organization")
    ErrCannotRemoveSelf    = errors.New("cannot remove yourself from organization")
    ErrCannotRemoveOwner   = errors.New("cannot remove organization owner")
    ErrInsufficientRole    = errors.New("insufficient role for this action")
    ErrCannotModifyHigher  = errors.New("cannot modify member with equal or higher role")
)

// Invitation errors
var (
    ErrInvitationNotFound      = errors.New("invitation not found")
    ErrInvitationExpired       = errors.New("invitation has expired")
    ErrInvitationAlreadyUsed   = errors.New("invitation has already been processed")
    ErrInvitationPending       = errors.New("invitation is still pending")
    ErrMaxInvitationsReached  = errors.New("maximum pending invitations reached")
    ErrRateLimitExceeded       = errors.New("rate limit exceeded, try again later")
    ErrEmailAlreadyInvited    = errors.New("user has already been invited to this organization")
    ErrCannotInviteSelf       = errors.New("cannot invite yourself")
    ErrEmailRequired          = errors.New("email is required for invitation")
    ErrEmailNotVerified       = errors.New("invitee email must be verified")
)

// Session errors
var (
    ErrOrgContextRequired = errors.New("organization context is required for this operation")
    ErrOrgNotSelected     = errors.New("no organization selected")
)
```

---

## Configuration Requirements

```go
type OrganizationConfig struct {
    // Limits
    MaxActiveOrgsPerOperator int           `env:"MAX_ACTIVE_ORGS" default:"2"`
    MaxMembersPerOrg         int           `env:"MAX_MEMBERS_PER_ORG" default:"100"`
    MaxPendingInvitations    int           `env:"MAX_PENDING_INVITATIONS" default:"20"`
    
    // Invitation
    InvitationTTL           time.Duration `env:"INVITATION_TTL" default:"168h"` // 7 days
    InvitationRateLimit     int           `env:"INVITATION_RATE_LIMIT" default:"5"` // per hour
    
    // Bootstrap
    BootstrapToken          string         `env:"VYORIX_BOOTSTRAP_TOKEN"` // Emergency recovery
    
    // Naming
    OrgNameMinLength        int           `env:"ORG_NAME_MIN_LENGTH" default:"2"`
    OrgNameMaxLength        int           `env:"ORG_NAME_MAX_LENGTH" default:"255"`
}
```

---

## Monitoring & Metrics

```go
// Metrics to track
const (
    // Counters
    MetricOrgsCreated        = "vyzorix_orgs_created_total"
    MetricOrgsDeleted       = "vyzorix_orgs_deleted_total"
    MetricInvitationsSent   = "vyzorix_invitations_sent_total"
    MetricInvitationsAccepted = "vyzorix_invitations_accepted_total"
    MetricInvitationsRejected = "vyzorix_invitations_rejected_total"
    MetricInvitationsExpired  = "vyzorix_invitations_expired_total"
    
    // Histograms
    MetricInvitationLatency = "vyzorix_invitation_create_duration_seconds"
    
    // Gauges
    MetricActiveOrgs        = "vyzorix_active_organizations"
    MetricPendingInvitations = "vyzorix_pending_invitations"
)
```

---

## Testing Requirements

### Unit Tests
- [ ] OrganizationService.CreateOrganization with race condition prevention
- [ ] InvitationService with token generation entropy
- [ ] Membership role hierarchy checks
- [ ] Max orgs enforcement
- [ ] Rate limiting

### Integration Tests
- [ ] Full invitation flow: create → email → accept → membership
- [ ] Org deletion cascades invitations
- [ ] Migration: nullable → required org_id
- [ ] Concurrent invitation acceptance (same token)
- [ ] Session org context persistence

### Load Tests
- [ ] 1000 invitations/second creation
- [ ] 100 orgs with 1000 members each
- [ ] Invitation token generation collision

### Security Tests
- [ ] Token entropy verification (no weak tokens)
- [ ] Rate limiting bypass attempts
- [ ] SQL injection in org name
- [ ] XSS in invitation notes
- [ ] CSRF on invitation endpoints

---

## Migration Checklist

### Pre-Migration
- [ ] Backup database
- [ ] Test migration on staging with copy of production data
- [ ] Prepare rollback procedure
- [ ] Notify users of maintenance window
- [ ] Disable new org creation during migration

### Migration Execution
- [ ] Phase 1: Run 040_organizations.sql
- [ ] Phase 2: Backfill existing operators to "default org"
- [ ] Phase 3: Enable require_org_id for new device registrations
- [ ] Phase 4: Enable strict mode (all operations require org)

### Post-Migration
- [ ] Verify all existing users have memberships
- [ ] Verify all devices have organization_id
- [ ] Enable org-switching UI
- [ ] Test invitation flow end-to-end
- [ ] Enable monitoring dashboards
- [ ] Update documentation

### Rollback Procedure
```bash
# If critical error found:
# 1. Stop API server
# 2. Restore database from backup
# 3. Disable new migration-enabled code paths
# 4. Deploy previous version
# 5. Investigate issue
```

---

##  ADDITIONAL DESIGN QUESTIONS REQUIRING CLARIFICATION

### 1. Device Registration Without Organization

**Problem**: What if user wants to test the app without creating an org?

**Options**:
```
Option A: Device registration requires org
├── User must create/join org before registering devices
├── Existing users without org prompted to create one
└── Cleaner model, devices always belong to org

Option B: Personal device space + org later
├── User can register device to "personal" (no org)
├── Personal devices visible only to that user
├── Can later "promote" personal device to org
└── More complex, but better onboarding

Option C: Automatic org on first device registration
├── User registers device → auto-creates "Personal" org
├── User becomes super_admin of their personal org
└── Simplest UX
```

**Recommendation**: Option A - requires org, but show "Create or Join" flow immediately on signup.

### 2. API Key Organization Scope

**Problem**: Currently API keys are per-operator. How do they work with orgs?

**Options**:
```
Option A: API keys are org-scoped
├── API key = "org_api_key"
├── Can only register devices to that specific org
├── Key includes org_id in payload
└── Cleaner isolation

Option B: API keys are operator-scoped, inherit operator's org
├── Same as current
├── Operator with multiple orgs: which org for device?
└── Need to specify org on each registration
```

**Recommendation**: Option A - org-scoped API keys for cleaner isolation.

### 3. Device IMEI Uniqueness

**Problem**: Can the same IMEI exist in multiple organizations?

**Options**:
```
Option A: Globally unique IMEI
├── Same IMEI cannot exist in two orgs
├── Device must be deregistered before registering to new org
└── Prevents confusion, matches real-world

Option B: Per-org unique IMEI
├── Same IMEI can exist in multiple orgs (different devices)
└── Complex - how to distinguish same physical device?
```

**Recommendation**: Option A - globally unique IMEI. Device must be deregistered to register in new org.

### 4. Invitation Email Validation

**Problem**: What if invitee signs up with a different email?

**Scenario**: Invitation sent to john@example.com, but user signs up as johnny@example.com (typo)

**Options**:
```
Option A: Strict email match required
├── Invitation email must match signed-in operator's email exactly
├── Prevents wrong-account acceptance
└── User must correct invitation or admin must re-invite

Option B: Allow email mismatch
├── User can accept invitation regardless of email
├── Accept based on token only
└── Less secure but more forgiving
```

**Recommendation**: Option A - strict email match. Admin should verify email and re-invite if typo.

### 5. Last Super Admin Protection

**Problem**: What if only super_admin wants to leave org?

**Options**:
```
Option A: Block self-removal if last super_admin
├── Cannot remove last super_admin from org
├── Must promote another member first
└── Prevents orphaned orgs

Option B: Allow, org becomes unmanageable
├── Last super_admin can leave
├── Org has no super_admin
└── Admin can request super_admin status
```

**Recommendation**: Option A - block self-removal if last super_admin.

### 6. Organization Deletion Prerequisites

**Problem**: What must happen before deleting an org?

**Checklist before delete**:
```
□ All devices deregistered OR transferred
□ All members removed except deleter
□ All pending invitations canceled
□ All API keys deleted
□ All webhook configurations removed
□ Audit logs: Delete OR archive
```

**Recommendation**: Enforce all prerequisites before delete. Return clear error messages listing what's blocking deletion.

### 7. Role Hierarchy Enforcement

**Problem**: Who can invite/demote whom within org?

**Invitation Permissions**:
```
super_admin can invite → admin, operator, viewer (NOT super_admin)
admin can invite → operator, viewer (NOT admin or super_admin)
operator cannot invite
viewer cannot invite
```

**Demotion Permissions**:
```
super_admin can demote → admin, operator, viewer
admin can demote → operator, viewer (NOT super_admin or other admin)
operator cannot demote
viewer cannot demote
```

**Protection**: Cannot modify members with equal or higher role.

### 8. Organization Name Uniqueness

**Problem**: Can two operators have orgs with the same name?

**Options**:
```
Option A: Globally unique org names
├── "My Company" can only exist once
└── Cleaner, but conflicts possible

Option B: Per-operator unique org names
├── Each operator can have "My Company"
├── Org identity = operator_id + name
└── More flexible
```

**Recommendation**: Option B - per-operator unique names. Two users can both have "Personal" org.

### 9. Webhook Scope

**Problem**: Currently webhooks are per-operator. How do they work with orgs?

**Options**:
```
Option A: Webhooks are org-scoped
├── Each org has its own webhook URLs
├── Device events in org trigger org's webhook
└── Clean separation

Option B: Operator webhooks span all org memberships
├── Same as current
├── Events include org_id in payload
└── Less change, but more complex routing
```

**Recommendation**: Option A - org-scoped webhooks. Each org has its own webhook configuration.

### 10. Cross-Organization Dashboard

**Problem**: If operator in multiple orgs, how do they view data?

**Options**:
```
Option A: Always org-scoped
├── Must select org first
├── Dashboard shows only selected org's data
├── No aggregate view
└── Clear separation

Option B: Multi-org aggregate view
├── "All Organizations" view
├── Shows combined stats from all orgs
├── Can drill down to specific org
└── More complex, but convenient

Option C: Both views
├── Default to selected org
├── Toggle to "All Organizations"
└── Most flexible
```

**Recommendation**: Option A initially - always org-scoped. Can add aggregate view later.

### 11. Device Transfer Feature

**Problem**: How does device transfer between orgs work?

**Design**:
```
Pre-transfer requirements:
□ Device must be deregistered (offline, no active sessions)
OR
□ Device must be online and accept transfer command

Transfer flow:
1. Source org super_admin initiates transfer
2. Select target org (must be member of target)
3. Device receives transfer command (if online)
4. Device deregisters from source, registers to target
5. Audit log records transfer

Post-transfer:
□ Device history/telemetry stays with device (transferred)
□ New org has access to all device data
□ Old org loses access immediately
```

**Transfer endpoint**: `POST /v1/organizations/:id/devices/:deviceId/transfer`
```json
{
  "target_org_id": "uuid",
  "notes": "Transferring to client org"
}
```

### 12. Membership Limits Per Org

**Problem**: Should there be a max members per org?

**Configuration**:
```go
type OrganizationConfig struct {
    // Limits
    MaxMembersPerOrg int `env:"MAX_MEMBERS_PER_ORG" default:"100"`
}
```

**Recommendation**: Default 100 members per org, configurable. Warn at 80%, block at 100%.

### 13. Device Limits Per Org

**Problem**: Should there be a max devices per org?

**Options**:
```
Option A: Unlimited devices
├── No limit on device count
└── May hit performance issues at scale

Option B: Per-org device limits
├── Configurable max devices per org
├── Tiered limits based on subscription
└── Better resource management
```

**Recommendation**: Add later when needed. Start unlimited, monitor performance.

### 14. Bootstrap/First Org Creation

**Problem**: How is the very first org created without UI?

**Options**:
```
Option A: First OAuth signup auto-creates personal org
├── User signs up with Google/GitHub
├── System auto-creates "Personal" org
├── User becomes super_admin
└── Simplest flow

Option B: CLI command for first org
├── `vyzorix-cli org create --name="My Org" --bootstrap-token=xxx`
├── Creates org and makes you super_admin
└── More control

Option C: Invitation required even for first user
├── Admin sends invite to first user
├── First user signs up, accepts invite
└── Most restrictive
```

**Recommendation**: Option A - first OAuth signup auto-creates personal org. Simplest onboarding.

---

## Summary of Required Clarifications

| # | Question | Options | Recommendation | Status |
|---|----------|---------|---------------|--------|
| 1 | Device registration without org | Require org / Personal space / Auto-create | Require org |  CONFIRMED |
| 2 | API key org scope | Org-scoped / Operator-scoped | Org-scoped |  CONFIRMED |
| 3 | IMEI uniqueness | Global / Per-org | Global (IMEI is device hardware ID) |  CONFIRMED |
| 4 | Invitation email match | Strict / Loose | Strict |  CONFIRMED |
| 5 | Last super_admin leaving | Block / Allow | Block |  CONFIRMED |
| 6 | Org deletion prerequisites | Checklist enforced | Enforce all |  CONFIRMED |
| 7 | Role hierarchy | As specified | As specified |  CONFIRMED |
| 8 | Org name uniqueness | Global / Per-operator | Per-operator |  CONFIRMED |
| 9 | Webhook scope | Org-scoped / Operator-scoped | Org-scoped |  CONFIRMED |
| 10 | Cross-org dashboard | Org-scoped only / Aggregate / Both | Org-scoped only |  CONFIRMED |
| 11 | Device transfer feature | Required | Implement |  CONFIRMED |
| 12 | Membership limits | Default 100 | Implement |  CONFIRMED |
| 13 | Device limits | Unlimited / Tiered | Unlimited initially |  CONFIRMED |
| 14 | First org creation | Auto / CLI / Invite | Auto-create |  CONFIRMED |

### All 14 Questions CONFIRMED 

---

##  GRAPHQL SCHEMA IMPACT ANALYSIS

### Files to Check and Refactor

```
graphql/
├── schema/
│   ├── objects.go           # Add Organization, Membership types
│   ├── enums.go            # Add OrgRole enum, update Role enum
│   ├── schema.go           # Add org queries/mutations
│   ├── subscription.go     # Add org-scoped subscriptions
│   └── scalars.go          # May need new scalars
│
├── resolver/
│   ├── query_resolver.go       # Add organization queries
│   ├── mutation_resolver.go    # Add org mutations
│   ├── resolver.go            # Add organization resolver
│   ├── helpers.go             # Add org context helpers
│   ├── inbox_resolver.go     # Update to filter by org
│   ├── updates_resolver.go   # Update to filter by org
│   └── subscription_resolver.go # Add org subscriptions
│
├── middleware/
│   └── gql_auth.go         # Update role checks to org-scoped
│
├── errors/
│   └── gql_errors.go       # Add org-specific errors
│
├── context/
│   └── context.go          # Add org context to resolver
│
└── adapters/
    └── gql_presenter.go    # Update for org responses
```

### Detailed File Changes

#### 1. schema/objects.go

```go
// ADD: Organization type
type Organization struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    CreatedBy *Operator  `json:"createdBy"`
    Members   []*Member `json:"members"`
    CreatedAt time.Time `json:"createdAt"`
    IsActive  bool      `json:"isActive"`
}

// ADD: OrganizationMembership type
type OrganizationMembership struct {
    Organization *Organization `json:"organization"`
    Operator    *Operator     `json:"operator"`
    Role        OrgRole       `json:"role"`
    JoinedAt    time.Time     `json:"joinedAt"`
    InvitedBy   *Operator     `json:"invitedBy,omitempty"`
}

// ADD: Invitation type (for GraphQL)
type Invitation struct {
    ID            string    `json:"id"`
    Organization  *Organization `json:"organization"`
    Email         string    `json:"email"`
    Role          OrgRole   `json:"role"`
    Status        string    `json:"status"`
    InvitedAt     time.Time `json:"invitedAt"`
    ExpiresAt     time.Time `json:"expiresAt"`
}

// UPDATE: Operator type - add memberships field
type Operator struct {
    // ... existing fields
    Memberships []*OrganizationMembership `json:"memberships"`
}

// UPDATE: Device type - add organization field
type Device struct {
    // ... existing fields
    OrganizationID string `json:"organizationId"`
    Organization  *Organization `json:"organization"`
}
```

#### 2. schema/enums.go

```go
// ADD: OrgRole enum
var OrgRoleEnum = graphql.NewEnum(graphql.EnumConfig{
    Name: "OrgRole",
    Values: graphql.EnumValueConfigMap{
        "SUPER_ADMIN": &graphql.EnumValueConfig{
            Value: "super_admin",
        },
        "ADMIN": &graphql.EnumValueConfig{
            Value: "admin",
        },
        "OPERATOR": &graphql.EnumValueConfig{
            Value: "operator",
        },
        "VIEWER": &graphql.EnumValueConfig{
            Value: "viewer",
        },
    },
})

// UPDATE: Deprecate global Role enum (or keep for backward compat)
// Keep Role enum but mark as deprecated in docs
```

#### 3. schema/schema.go

```go
// ADD: Organization queries
organization: &graphql.Field{
    Type: organizationType,
    Args: graphql.FieldConfigArgument{
        "id": &graphql.ArgumentConfig{
            Type: graphql.NewNonNull(graphql.ID),
        },
    },
    Resolve: func(p graphql.ResolveParams) (interface{}, error) {
        // Check membership, return org
    },
}

organizations: &graphql.Field{
    Type: graphql.NewList(organizationType),
    Resolve: func(p graphql.ResolveParams) (interface{}, error) {
        // Return user's organizations
    },
}

myMemberships: &graphql.Field{
    Type: graphql.NewList(memberType),
    Resolve: func(p graphql.ResolveParams) (interface{}, error) {
        // Return user's memberships
    },
}

// ADD: Organization mutations
createOrganization: &graphql.Field{
    Type: createOrganizationPayloadType,
    Args: graphql.FieldConfigArgument{
        "name": &graphql.ArgumentConfig{
            Type: graphql.NewNonNull(graphql.String),
        },
    },
    Resolve: CreateOrganizationResolver,
}

// ADD: Membership mutations
inviteMember: &graphql.Field{...}
removeMember: &graphql.Field{...}
updateMemberRole: &graphql.Field{...}

// ADD: Invitation mutations
acceptInvitation: &graphql.Field{...}
rejectInvitation: &graphql.Field{...}

// ADD: Device transfer mutation
transferDevice: &graphql.Field{...}
```

#### 4. schema/subscription.go

```go
// ADD: Organization-scoped subscriptions
organizationEvent: &graphql.Field{
    Type: organizationEventUnion,
    Args: graphql.FieldConfigArgument{
        "orgId": &graphql.ArgumentConfig{
            Type: graphql.NewNonNull(graphql.ID),
        },
    },
    Subscribe: OrganizationEventSubscription,
}

memberEvent: &graphql.Field{
    Type: memberEventType,
    Args: graphql.FieldConfigArgument{
        "orgId": &graphql.ArgumentConfig{
            Type: graphql.NewNonNull(graphql.ID),
        },
    },
    Subscribe: MemberEventSubscription,
}
```

#### 5. resolver/query_resolver.go

```go
// ADD: Organization queries
func (r *Resolver) OrganizationQuery(p graphql.ResolveParams) (interface{}, error) {
    orgID := p.Args["id"].(string)
    operator := GetOperatorFromContext(p.Context)
    
    // Check membership
    membership := operator.GetMembership(orgID)
    if membership == nil {
        return nil, gqlerror.Errorf("Not a member of this organization")
    }
    
    return GetOrganization(orgID)
}

func (r *Resolver) OrganizationsQuery(p graphql.ResolveParams) (interface{}, error) {
    operator := GetOperatorFromContext(p.Context)
    return operator.GetOrganizations(), nil
}

func (r *Resolver) MyMembershipsQuery(p graphql.ResolveParams) (interface{}, error) {
    operator := GetOperatorFromContext(p.Context)
    return operator.GetMemberships(), nil
}

// UPDATE: Existing device queries - add org filter
func (r *Resolver) DevicesQuery(p graphql.ResolveParams) (interface{}, error) {
    operator := GetOperatorFromContext(p.Context)
    orgID := p.Args["orgId"] // Optional, defaults to selected org
    
    if orgID == nil {
        orgID = operator.GetSelectedOrgID()
    }
    
    // Check membership and return org-scoped devices
    return GetDevicesByOrganization(orgID.(string))
}
```

#### 6. resolver/mutation_resolver.go

```go
// ADD: Organization mutations
func (r *Resolver) CreateOrganizationMutation(p graphql.ResolveParams) (interface{}, error) {
    operator := GetOperatorFromContext(p.Context)
    name := p.Args["name"].(string)
    
    return CreateOrganization(operator, name)
}

func (r *Resolver) DeleteOrganizationMutation(p graphql.ResolveParams) (interface{}, error) {
    operator := GetOperatorFromContext(p.Context)
    orgID := p.Args["id"].(string)
    
    // Check: must be super_admin of org
    // Check: must have no devices (or transfer/delete first)
    // Check: must have no other members (or remove first)
    
    return DeleteOrganization(operator, orgID)
}

// ADD: Membership mutations
func (r *Resolver) InviteMemberMutation(p graphql.ResolveParams) (interface{}, error) {
    operator := GetOperatorFromContext(p.Context)
    orgID := p.Args["orgId"].(string)
    email := p.Args["email"].(string)
    role := p.Args["role"].(OrgRole)
    
    // Check: operator must be admin+
    // Create invitation, send email
}

func (r *Resolver) AcceptInvitationMutation(p graphql.ResolveParams) (interface{}, error) {
    operator := GetOperatorFromContext(p.Context)
    token := p.Args["token"].(string)
    notes := p.Args["notes"].(string)
    
    // Validate token
    // Check email matches operator
    // Create membership
    // Send notification to inviter
}

func (r *Resolver) RejectInvitationMutation(p graphql.ResolveParams) (interface{}, error) {
    operator := GetOperatorFromContext(p.Context)
    token := p.Args["token"].(string)
    notes := p.Args["notes"].(string)
    
    // Validate token
    // Update invitation status
    // Send notification to inviter
}
```

#### 7. resolver/helpers.go

```go
// ADD: Org context helpers
func GetOperatorFromContext(ctx context.Context) *Operator {
    // Extract operator from context
}

func GetOrgFromContext(ctx context.Context) (*Organization, error) {
    // Get selected org from context
}

func RequireOrgMembership(operator *Operator, orgID string, minRole OrgRole) error {
    membership := operator.GetMembership(orgID)
    if membership == nil {
        return ErrNotOrgMember
    }
    if !membership.Role.IsAtLeast(minRole) {
        return ErrInsufficientRole
    }
    return nil
}

func GetOrgDevice(operator *Operator, deviceID string) (*Device, error) {
    // Get device, verify operator is org member
}
```

#### 8. middleware/gql_auth.go

```go
// UPDATE: Role checks to org-scoped
func RequireOrgRole(orgID string, minRole OrgRole) func(graphql.ResolveParams) (interface{}, error) {
    return func(p graphql.ResolveParams) (interface{}, error) {
        operator := GetOperatorFromContext(p.Context)
        
        membership := operator.GetMembership(orgID)
        if membership == nil {
            return nil, gqlerror.Errorf("Not a member of this organization")
        }
        
        if !membership.Role.IsAtLeast(minRole) {
            return nil, gqlerror.Errorf("Insufficient permissions")
        }
        
        return nil, nil
    }
}

// Deprecate old global role checks
// Keep RequireSuperAdmin() for backward compat but log deprecation warning
```

### GraphQL Impact Summary

| Component | File | Changes Required | Effort |
|-----------|------|----------------|--------|
| Organization Type | `objects.go` | Add Organization, Membership, Invitation structs | Medium |
| Enums | `enums.go` | Add OrgRole enum | Low |
| Queries | `schema.go` | Add 4 new queries | Medium |
| Mutations | `schema.go` | Add 8 new mutations | High |
| Subscriptions | `subscription.go` | Add org-scoped subscriptions | Medium |
| Query Resolver | `query_resolver.go` | Add org queries, update device queries | High |
| Mutation Resolver | `mutation_resolver.go` | Add all org mutations | High |
| Helpers | `helpers.go` | Add org context helpers | Medium |
| Auth Middleware | `middleware/gql_auth.go` | Update role checks | Medium |
| Context | `context.go` | Add org to context | Low |
| Errors | `errors/gql_errors.go` | Add org-specific errors | Low |
| Inbox Resolver | `inbox_resolver.go` | Filter by org | Medium |
| Updates Resolver | `updates_resolver.go` | Filter by org | Medium |

**Total Files to Modify: 12**
**Estimated Effort: High (2-3 weeks)**

---

##  MIGRATION STRATEGY (SIMPLIFIED)

### Important: No Existing Data

**️ ASSUMPTION: There are NO existing users, devices, or organizations in the system.**

This is a greenfield implementation for org features. All data is created fresh.

```
CURRENT STATE (Greenfield):
┌─────────────────┐
│  EMPTY DATABASE  │
│  (fresh start)   │
└─────────────────┘

TARGET STATE (With Organizations):
┌────────────┐     ┌────────────┐     ┌────────────┐
│  Operator  │────│ Membership │────│Organization│
│           │     │  (role)   │     │            │
└────────────┘     └────────────┘     └────────────┘
                         │
                         ▼
                    ┌────────────┐
                    │   Device   │
                    └────────────┘
```

### Migration Phase (Single Phase)

Since there's no existing data, we just need ONE migration:

```sql
-- Migration 040_organizations.sql
-- Creates all new tables at once

BEGIN;

-- Create organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_by TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    deleted_at INTEGER,
    is_active INTEGER DEFAULT 1,
    max_members INTEGER DEFAULT 100,
    UNIQUE(created_by, name)  -- Per-operator unique names
);

-- Create organization_members table
CREATE TABLE IF NOT EXISTS organization_members (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    operator_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('super_admin', 'admin', 'operator', 'viewer')),
    invited_by TEXT,
    joined_at INTEGER NOT NULL,
    removed_at INTEGER,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'removed')),
    UNIQUE(organization_id, operator_id),
    FOREIGN KEY (organization_id) REFERENCES organizations(id),
    FOREIGN KEY (operator_id) REFERENCES operators(id),
    FOREIGN KEY (invited_by) REFERENCES operators(id)
);

-- Create invitations table
CREATE TABLE IF NOT EXISTS invitations (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    email TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'operator', 'viewer')),
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'expired')),
    token TEXT NOT NULL UNIQUE,
    inviter_notes TEXT,
    invitee_notes TEXT,
    invited_by TEXT NOT NULL,
    invited_at INTEGER NOT NULL,
    responded_at INTEGER,
    expires_at INTEGER NOT NULL,
    responder_id TEXT,
    FOREIGN KEY (organization_id) REFERENCES organizations(id),
    FOREIGN KEY (invited_by) REFERENCES operators(id),
    FOREIGN KEY (responder_id) REFERENCES operators(id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_org_members_operator ON organization_members(operator_id);
CREATE INDEX IF NOT EXISTS idx_org_members_org ON organization_members(organization_id);
CREATE INDEX IF NOT EXISTS idx_invitations_token ON invitations(token);
CREATE INDEX IF NOT EXISTS idx_invitations_email ON invitations(email);
CREATE INDEX IF NOT EXISTS idx_invitations_org_status ON invitations(organization_id, status);

-- Add organization_id to devices (NOT NULL - required from now on)
ALTER TABLE devices ADD COLUMN organization_id TEXT NOT NULL;

-- Add organization_id to sessions (nullable for backward compat)
ALTER TABLE sessions ADD COLUMN organization_id TEXT;

-- Add organization_id to api_keys (nullable for backward compat)
ALTER TABLE api_keys ADD COLUMN organization_id TEXT;

COMMIT;
```

### Migration Verification

```sql
-- Verify tables created
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'main' 
AND table_name IN ('organizations', 'organization_members', 'invitations');

-- Should return: organizations, organization_members, invitations

-- Verify columns added
PRAGMA table_info(devices);
-- Should show organization_id column

-- Verify constraints
PRAGMA foreign_key_list('organization_members');
-- Should show FK to organizations and operators
```

### Migration Rollback (if needed)

```sql
-- If we need to rollback:
BEGIN;

ALTER TABLE devices DROP COLUMN organization_id;
ALTER TABLE sessions DROP COLUMN organization_id;
ALTER TABLE api_keys DROP COLUMN organization_id;

DROP TABLE IF EXISTS invitations;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;

COMMIT;
```

---

##  IMPLEMENTATION PHASES

### Phase 1: Foundation (Database + Domain)
1. Create migration 040_organizations.sql
2. Create domain/organization package
3. Create domain/invitation package  
4. Create storage implementations
5. **DO NOT enable org features yet**

### Phase 2: Application Layer
1. Create organization_service.go
2. Create invitation_service.go
3. Create member_service.go
4. Modify auth services (no auto-role on signup)

### Phase 3: API Endpoints (REST)
1. Create organization handlers
2. Create invitation handlers
3. Create member handlers
4. Add new routes
5. Add middleware (org_context, org_membership)

### Phase 4: GraphQL Updates
1. Update schema (add Organization types)
2. Update resolvers (add org queries/mutations)
3. Update auth middleware
4. Add org-scoped subscriptions

### Phase 5: Device Transfer Feature
1. Design transfer API
2. Implement transfer service
3. Add transfer endpoint
4. Handle online/offline device transfer

### Phase 6: Migration Execution
1. Run migration M1 (create tables)
2. Run migration M2 (backfill data)
3. Verify data integrity
4. Enable org context (Phase M3)
5. Monitor for issues
6. Enable strict mode (Phase M4)

### Phase 7: Client Updates
1. Update API client (add organization types)
2. Update frontend (org switcher)
3. Update invitation flow UI
4. Test full invitation flow

### Phase 8: Cleanup
1. Remove deprecated global role column
2. Remove old admin endpoints
3. Update documentation
4. Update OpenAPI spec

---

## MFA & BACKUP CODES (Per-Operator)

### Design Decision

**MFA and Backup Codes remain per-operator, NOT org-scoped.**

- MFA/Backup codes are GLOBAL - work across ALL org memberships
- Stored in `operators` table (already exists)
- **No changes needed**

---

## OPERATOR DELETION FLOW

### Deletion Flow

1. User initiates account deletion
2. System validates:
   - Cannot delete if last super_admin of any org
   - Warn about orphaned devices
   - Cancel pending invitations sent by user
3. User confirms with password
4. System cascade deletes:
   - Remove from all org memberships
   - Cancel all invitations sent by operator
   - Delete sessions
   - Delete operator record
5. Devices remain but orphaned (historical reference kept)

### Transfer Ownership (Optional)

Before deletion, user can transfer org ownership:
1. Select new owner (must be admin or operator)
2. New owner receives confirmation email
3. New owner accepts transfer
4. Old owner becomes admin
5. Old owner can then leave org

---

## DEVICE TRANSFER FEATURE

### Prerequisites

- Device must be OFFLINE
- User must have permission in source AND target orgs

### Transfer Flow

1. Admin initiates transfer
2. System validates prerequisites
3. Updates `device.organization_id`
4. Creates audit log
5. Notifies target org admins

### Transfer API

```
POST /v1/organizations/:id/devices/:imei/transfer
{
    "target_org_id": "uuid",
    "notes": "optional notes"
}
```

### Error Codes

| Error | Meaning |
|-------|---------|
| `device_must_be_offline` | Device must be disconnected |
| `device_already_in_org` | Device is already in target org |
| `cannot_transfer_to_self` | Source and target are same |
| `org_at_member_limit` | Target org has reached limit |
