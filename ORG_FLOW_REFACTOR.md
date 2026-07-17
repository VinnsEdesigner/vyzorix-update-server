# Organization Flow Refactor - Issue Tracker

> **Created:** 2026-07-16
> **Updated:** 2026-07-16
> **Status:** In Progress
> **Target:** Complete organization flow implementation
> **Model:** Multi-Tenant Organization (per REFACTOR_PLAN.md)

---

## 🎯 NEW MODEL OVERVIEW

Per REFACTOR_PLAN.md, the system follows this model:
- **Operators** are global identities (email, password, OAuth) - **NO global role**
- **Organizations** are tenants that own resources
- **Memberships** link operators to organizations with scoped roles
- **Role is ONLY determined when an organization is selected**
- **MFA/Backup Codes remain per-operator (global)**

---

## 🔴 CRITICAL ISSUES

### Issue #1: Code References Non-Existent `op.Role` Field
**Severity:** Critical (Runtime Bug)
**Files Affected:**
- `apps/api/internal/domain/operator/op_responses.go:34`
- `apps/api/internal/application/auth/auth_login_session.go:57,71`
- `apps/api/internal/application/auth/auth_device_recognition.go:235,255`
- `apps/api/internal/api/handlers/auth/auth_me.go:48`
- `apps/api/internal/api/handlers/auth/auth_settings.go:191,236`
- `apps/api/internal/api/handlers/auth/auth_mfa.go:321,346`
- `apps/api/internal/api/middleware/updates_admin_auth.go:60`
- `apps/api/internal/api/handlers/telemetry_history.go:600`

**Problem:** 
Code references `op.Role` but the `Operator` struct intentionally has NO `Role` field (per new model).

**Required Fix:**
1. Remove all references to `op.Role` from these files
2. Role should ONLY come from organization membership when org is selected
3. Update DTOs to NOT return role at login/auth level

---

### Issue #2: File Corruption in `auth_constructors.go`
**Severity:** Critical (Compilation Error)
**File:** `apps/api/internal/application/auth/auth_constructors.go`

**Problem:**
Lines are improperly concatenated/corrupted.

**Required Fix:**
Rewrite the file properly with:
- `ResolveOrganizationForOperator()` - auto-select org logic
- `SelectOrganization()` - switch organization
- `ValidateOrganizationMembership()` - validate org access
- `SelectOrganizationResult` struct

---

## 🟡 MISSING FEATURES

### Issue #3: CreateOrganizationRequest Missing Fields
**File:** `apps/api/internal/domain/organization/organization_entity.go`

**Current:**
```go
type CreateOrganizationRequest struct {
    Name       string
    MaxMembers int
}
```

**Required:**
```go
type CreateOrganizationRequest struct {
    Name        string   // Optional - defaults to "personal"
    Description string   // Optional - organization description
    MaxMembers  int      // Optional - max members limit
    Role        string   // Creator's role in this org - "super_admin" or "admin" only
}
```

---

### Issue #4: Login Response Wrong Format
**File:** `apps/api/internal/application/dto/auth.go`

**Current (WRONG for new model):**
```go
type LoginResponse struct {
    OperatorID string `json:"operator_id"`
    Email      string `json:"email"`
    Name       string `json:"name"`
    Role       string `json:"role"`  // WRONG - no global role!
    MFAEnabled bool   `json:"mfa_enabled"`
}
```

**Required (New Model):**
```go
type LoginResponse struct {
    OperatorID            string               `json:"operator_id"`
    Email                 string               `json:"email"`
    Name                  string               `json:"name"`
    MFAEnabled            bool                 `json:"mfa_enabled"`
    NeedsOrganization     bool                 `json:"needs_organization"`
    Organizations         []OrganizationInfo   `json:"organizations,omitempty"`
    LastOrganizationID    string               `json:"last_organization_id,omitempty"`
    SelectedOrganization  *OrganizationInfo    `json:"selected_organization,omitempty"`
}

type OrganizationInfo struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    Role string `json:"role"`  // Role IN THIS organization
}
```

---

### Issue #5: Auth/Me Endpoint Wrong Response
**File:** `apps/api/internal/api/handlers/auth/auth_me.go`

**Problem:**
Returns `role: op.Role` (doesn't exist) and no organization context.

**Required Fix:**
Return same organization-aware response as login.

---

## ✅ CORRECTLY IMPLEMENTED

### Dashboard Routes Protected Without Organization
**File:** `apps/api/internal/api/server_routes.go:176-178`
```go
dashboard := router.Group("/dashboard")
dashboard.Use(middleware.NewOrganizationContext(nil).Middleware())
dashboard.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
```

### Organization Auto-Selection Logic (exists, but corrupted)
Functions in `auth_constructors.go`:
- `ResolveOrganizationForOperator()` - returns org based on LastOrganizationID or single org
- `SelectOrganization()` - allows switching orgs

### Organization Members Repository
**File:** `apps/api/internal/domain/organization/member_repository.go`

---

## 📋 IMPLEMENTATION CHECKLIST

### Phase 1: Fix Critical Bugs
- [x] Operator struct has NO Role field (correct per new model)
- [x] Fix corruption in `auth_constructors.go` - rewritten properly
- [x] Remove remaining `op.Role` references from codebase

### Phase 2: Update DTOs
- [x] Update `LoginResponse` - removed Role, added organization fields
- [x] Update `LoginWithTokensResponse` - same changes
- [x] Update `OperatorResponse` - removed Role, added organization fields
- [x] Add `OrganizationInfo` struct to dto (done)

### Phase 3: Update Auth Handlers
- [x] Update `auth_me.go` - returns organization context
- [x] Update login handlers to populate organization data
- [x] Ensure MFA responses don't include role

### Phase 4: Organization Flow Backend
- [x] Update `CreateOrganizationRequest` - added Description, made Name optional, added Role
- [x] Update `OrganizationService.CreateOrganization` - accepts new fields
- [x] Update `OrganizationHandler.Create` - passes new fields
- [x] Organization entity has Description field

### Phase 5: Organization Selection
- [x] `ResolveOrganizationForOperator` - implemented in auth_constructors.go
- [x] `SelectOrganization` - implemented in auth_constructors.go
- [x] `GetOperatorOrganizations` - implemented in auth_constructors.go
- [ ] Add `POST /v1/auth/organizations/select` endpoint
- [ ] Update session to track selected organization

### Remaining Issues to Fix
1. `auth_device_recognition.go` - LoginWithDevice function still references `op.Role`
2. Add `POST /v1/auth/organizations/select` endpoint for switching orgs
3. Update session entity to store selected organization ID

---

## 📁 KEY FILES

| Component | Location |
|-----------|----------|
| Operator Entity | `internal/domain/operator/operator_entity.go` |
| Organization Entity | `internal/domain/organization/organization_entity.go` |
| Member Entity | `internal/domain/organization/member_entity.go` |
| Create Org Handler | `internal/api/handlers/organization/organization_handler.go` |
| Create Org Service | `internal/application/organization/organization_service.go` |
| Login Handler | `internal/api/handlers/auth/auth_login.go` |
| Auth/Me Handler | `internal/api/handlers/auth/auth_me.go` |
| Auth Constructors | `internal/application/auth/auth_constructors.go` (CORRUPTED) |
| Org Context Middleware | `internal/api/middleware/org_context.go` |
| Org Membership Middleware | `internal/api/middleware/org_membership.go` |
| Server Routes | `internal/api/server_routes.go` |
| DTOs | `internal/application/dto/auth.go` |
