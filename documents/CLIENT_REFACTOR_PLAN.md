# Client-Side Refactoring Plan: Multi-Tenant Organization Model

## Executive Summary

This document outlines all client-side changes needed to support the multi-tenant organization model on the frontend (React SPA) and API client (TypeScript package).

---

## Current Architecture

### API Client Structure (`packages/API_Client/src/`)

```
API_Client/src/
├── domain/                    # Domain entities & mappers
│   ├── auth/
│   │   ├── auth-entity.ts          # Operator, Session types
│   │   ├── auth-mappers.ts        # API → Domain mapping
│   │   └── auth-validators.ts
│   ├── oauth/
│   │   ├── oauth-entity.ts        # OAuth types (NEEDS UPDATE)
│   │   └── oauth-mappers.ts       # OAuth callback parsing
│   ├── device/
│   ├── apikey/
│   ├── session/
│   └── [other domains]
│
├── vyzorServer/
│   ├── rest/                    # REST API endpoints
│   │   ├── auth/
│   │   │   └── rest-auth-endpoints.ts
│   │   ├── device/
│   │   ├── apikey/
│   │   └── [other REST endpoints]
│   │
│   ├── graphql/                 # GraphQL queries/mutations
│   │   ├── device/
│   │   ├── apikey/
│   │   ├── commands/
│   │   ├── logs/
│   │   ├── updates/
│   │   ├── diagnostics/
│   │   ├── settings/
│   │   ├── registration/
│   │   └── apikey/
│   │
│   └── websocket/
│
└── index.ts                    # Exports
```

### Web App Structure (`apps/web/src/`)

```
web/src/
├── routes/                     # Page components
│   ├── __root.tsx             # Root layout
│   ├── dashboard.tsx          # Main dashboard (NEEDS UPDATE)
│   ├── device.tsx
│   ├── logs.tsx
│   ├── auth.callback.tsx      # OAuth callback (NEEDS UPDATE)
│   ├── login.tsx
│   ├── create-account.tsx
│   ├── settings.tsx           # Settings pages
│   └── [other routes]
│
├── components/
│   ├── layout/
│   │   └── AppLayout.tsx     # Main layout wrapper
│   └── [other components]
│
├── hooks/                     # Custom React hooks
├── lib/
│   ├── api/
│   │   └── graphql/           # GraphQL client hooks
│   └── vyzorix-api.ts        # API utilities
└── integrations/
```

---

## Required Changes

## PART 1: API Client Changes

### 1.1 NEW Domain Packages to Create

```
packages/API_Client/src/domain/
├── organization/                    # NEW
│   ├── organization-entity.ts      # Organization, OrgMember types
│   ├── organization-mappers.ts     # API → Domain mapping
│   ├── organization-validators.ts  # Validation functions
│   └── index.ts
│
├── membership/                      # NEW
│   ├── membership-entity.ts        # Membership type with role
│   ├── membership-mappers.ts
│   └── index.ts
│
└── invitation/                     # NEW
    ├── invitation-entity.ts        # Invitation, InvitationStatus types
    ├── invitation-mappers.ts
    ├── invitation-validators.ts
    └── index.ts
```

### 1.2 Files to MODIFY in Domain Layer

| File | Changes |
|------|---------|
| `domain/auth/auth-entity.ts` | Add `memberships` field, remove global role |
| `domain/device/device-entity.ts` | Add `organizationId` field |
| `domain/oauth/oauth-entity.ts` | Add error codes (email_required, etc.) |
| `domain/oauth/oauth-mappers.ts` | Update OAuth error handling |
| `domain/session/session-entity.ts` | Add `selectedOrganizationId` field |

### 1.3 NEW REST Endpoints to Create

```
packages/API_Client/src/vyzorServer/rest/
├── organization/                     # NEW
│   ├── index.ts
│   ├── organization-endpoints.ts    # POST/GET/PATCH/DELETE /v1/organizations
│   ├── member-endpoints.ts         # Member CRUD
│   └── invitation-endpoints.ts     # Invitation endpoints
│
├── invitation/                      # NEW (public endpoints)
│   ├── index.ts
│   └── invitation-endpoints.ts     # GET/POST /v1/invite/:token
│
└── me/                            # NEW
    ├── index.ts
    └── me-endpoints.ts             # GET /v1/me/memberships
```

### 1.4 REST Endpoints Summary

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/v1/organizations` | Create organization |
| GET | `/v1/organizations` | List my organizations |
| GET | `/v1/organizations/:id` | Get organization details |
| PATCH | `/v1/organizations/:id` | Update organization |
| DELETE | `/v1/organizations/:id` | Delete organization |
| GET | `/v1/organizations/:id/members` | List members |
| POST | `/v1/organizations/:id/members` | Add member |
| DELETE | `/v1/organizations/:id/members/:memberId` | Remove member |
| PATCH | `/v1/organizations/:id/members/:memberId` | Update member role |
| POST | `/v1/invitations` | Create invitation |
| GET | `/v1/invitations` | List pending invitations (as inviter) |
| GET | `/v1/invite/:token` | Get invitation by token (public) |
| POST | `/v1/invite/:token/approve` | Accept invitation |
| POST | `/v1/invite/:token/reject` | Reject invitation |
| GET | `/v1/me/memberships` | List my organization memberships |

### 1.5 NEW GraphQL Files to Create

```
packages/API_Client/src/vyzorServer/graphql/
├── organization/                    # NEW
│   ├── index.ts
│   ├── graphql-organization-types.ts
│   ├── graphql-organization-queries.ts
│   ├── graphql-organization-mutations.ts
│   └── graphql-organization-fragments.ts
│
├── membership/                      # NEW
│   ├── index.ts
│   ├── graphql-membership-types.ts
│   ├── graphql-membership-queries.ts
│   └── graphql-membership-mutations.ts
│
└── invitation/                     # NEW
    ├── index.ts
    ├── graphql-invitation-types.ts
    ├── graphql-invitation-queries.ts
    ├── graphql-invitation-mutations.ts
    └── graphql-invitation-subscriptions.ts
```

### 1.6 GraphQL Schema Changes Required

```graphql
# NEW: Organization types for GraphQL

type Organization {
  id: ID!
  name: String!
  createdBy: Operator!
  members: [OrganizationMembership!]!
  memberCount: Int!
  createdAt: DateTime!
  isActive: Boolean!
}

type OrganizationMembership {
  id: ID!
  organization: Organization!
  operator: Operator!
  role: OrgRole!
  joinedAt: DateTime!
  invitedBy: Operator
}

enum OrgRole {
  SUPER_ADMIN
  ADMIN
  OPERATOR
  VIEWER
}

type Invitation {
  id: ID!
  organization: Organization!
  email: String!
  role: OrgRole!
  status: InvitationStatus!
  invitedBy: Operator!
  invitedAt: DateTime!
  expiresAt: DateTime!
  inviterNotes: String
}

enum InvitationStatus {
  PENDING
  APPROVED
  REJECTED
  EXPIRED
}

# UPDATED: Operator type
type Operator {
  id: ID!
  email: String!
  name: String!
  memberships: [OrganizationMembership!]!  # NEW
  # Remove global role field
}

# UPDATED: Device type
type Device {
  id: ID!
  imei: String!
  organization: Organization!  # NEW
  operator: Operator!
  # ... other fields
}

# UPDATED: Session type
type Session {
  id: ID!
  operator: Operator!
  selectedOrganization: Organization  # NEW
  memberships: [OrganizationMembership!]!  # NEW
  createdAt: DateTime!
}

# NEW: Queries
type Query {
  myMemberships: [OrganizationMembership!]!
  organizations: [Organization!]!
  organization(id: ID!): Organization
  invitation(token: String!): Invitation
}

# NEW: Mutations
type Mutation {
  createOrganization(name: String!): Organization!
  updateOrganization(id: ID!, name: String!): Organization!
  deleteOrganization(id: ID!): DeleteOrganizationPayload!
  
  inviteMember(orgId: ID!, email: String!, role: OrgRole!, notes: String): Invitation!
  removeMember(orgId: ID!, memberId: ID!): Boolean!
  updateMemberRole(orgId: ID!, memberId: ID!, role: OrgRole!): OrganizationMembership!
  
  acceptInvitation(token: String!, notes: String): Invitation!
  rejectInvitation(token: String!, notes: String!): Invitation!
  
  transferOwnership(orgId: ID!, newOwnerId: ID!, notes: String): Organization!
  transferDevice(orgId: ID!, deviceId: ID!, targetOrgId: ID!, notes: String): Device!
}
```

---

## PART 2: Web App Changes

### 2.1 NEW Routes to Create

```
apps/web/src/routes/
├── organizations/                     # NEW directory
│   ├── index.tsx                    # Organizations list
│   ├── new.tsx                     # Create organization
│   ├── $orgId.tsx                  # Org layout wrapper
│   ├── $orgId/index.tsx            # Org dashboard
│   ├── $orgId/members.tsx          # Member management
│   ├── $orgId/invitations.tsx     # Invitation management
│   └── $orgId/settings.tsx        # Org settings
│
├── invite/
│   ├── $token.tsx                # Invitation page (public)
│   └── $token.accept.tsx          # Accept invitation
│   └── $token.reject.tsx         # Reject invitation
│
├── join-organization.tsx           # NEW: Join org by ID
└── transfer-ownership.tsx          # NEW: Transfer ownership
```

### 2.2 Routes to MODIFY

| Route | Changes |
|-------|---------|
| `dashboard.tsx` | Add org selector, filter by selected org |
| `auth.callback.tsx` | Handle new OAuth error codes (email_required, etc.) |
| `login.tsx` | Add "Join Organization" option |
| `create-account.tsx` | Show "Create or Join Organization" after signup |
| `settings.tsx` | Update operator settings page |
| `device.tsx` | Add org context to device operations |
| `__root.tsx` | Add org context provider |

### 2.3 NEW Components to Create

```
apps/web/src/components/
├── organization/                    # NEW
│   ├── OrganizationSelector.tsx    # Org switcher dropdown
│   ├── OrganizationCard.tsx       # Org display card
│   ├── CreateOrganizationModal.tsx
│   ├── JoinOrganizationModal.tsx
│   ├── MemberList.tsx
│   ├── MemberRow.tsx
│   ├── InviteMemberModal.tsx
│   ├── InvitationList.tsx
│   └── TransferOwnershipModal.tsx
│
├── invitation/                     # NEW
│   ├── InvitationBanner.tsx        # "You have pending invitation"
│   ├── InvitationCard.tsx
│   ├── AcceptInvitationModal.tsx
│   └── RejectInvitationModal.tsx
│
└── layout/
    ├── OrgLayout.tsx             # Layout wrapper for org pages
    └── OrgSwitcher.tsx          # Nav bar org selector
```

### 2.4 Components to MODIFY

| Component | Changes |
|-----------|---------|
| `layout/AppLayout.tsx` | Add org selector, org context |
| `layout/Sidebar.tsx` | Add organization management links |
| `layout/Header.tsx` | Add org indicator |
| `layout/Navbar.tsx` | Add org switcher |

### 2.5 NEW Hooks to Create

```
apps/web/src/hooks/
├── use-organization.ts             # Current org context
├── use-organizations.ts           # List user's orgs
├── use-memberships.ts            # User's memberships
├── use-invitation.ts             # Single invitation by token
├── use-pending-invitations.ts    # Invitations sent to user
└── use-org-permissions.ts       # Role-based permission checks
```

### 2.6 Context Providers to CREATE

```
apps/web/src/context/
├── OrganizationContext.tsx         # Selected org state
└── InvitationContext.tsx         # Pending invitation state
```

---

## PART 3: OAuth Error Handling Update

### 3.1 Current OAuth Callback Flow

```typescript
// Current: auth.callback.tsx
const oauth = params.get("oauth");
if (oauth === "success") {
  toast.success("Welcome!");
  navigate({ to: "/dashboard" });
}
// Missing: Error handling for email_required, etc.
```

### 3.2 Required OAuth Error Handling

```typescript
// NEW: Handle OAuth errors from server redirect

const oauth = params.get("oauth");
const code = params.get("code");
const message = params.get("message");
const provider = params.get("provider");
const retryable = params.get("retryable") !== "false";

if (oauth === "error" && code) {
  switch (code) {
    case "email_required":
      // Show dialog explaining email requirement
      showEmailRequiredDialog({ message, provider, helpUrl });
      break;
    
    case "email_not_verified":
      // Show dialog about verification
      showEmailNotVerifiedDialog({ message, provider, helpUrl });
      break;
    
    case "state_invalid":
      // Retryable - show retry option
      showRetryDialog({ message });
      break;
    
    default:
      // Generic error
      showErrorDialog({ message, retryable });
  }
  return;
}
```

### 3.3 NEW Dialog Components

```
apps/web/src/components/
├── dialogs/
│   ├── EmailRequiredDialog.tsx    # NEW
│   ├── EmailNotVerifiedDialog.tsx # NEW
│   ├── OAuthErrorDialog.tsx       # Generic OAuth error
│   └── RetryDialog.tsx            # Retryable error
```

---

## PART 4: User Flow Changes

### 4.1 Signup → First Login Flow

```
┌─────────────────────────────────────────────────────────────┐
│  1. User signs up (OAuth or Email)                          │
│                                                              │
│  2. Operator created (no org membership)                    │
│                                                              │
│  3. Redirect to "Join or Create Organization" page         │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐  │
│  │  You have no organization                               │  │
│  │                                                        │  │
│  │  [Create Organization]     [Join Organization]        │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                              │
│  4a. Create Organization                                    │
│      └── Enter: Org name, Your role (admin/super_admin)     │
│      └── Creates org, adds self as first member             │
│      └── Redirect to org dashboard                         │
│                                                              │
│  4b. Join Organization                                     │
│      └── Enter: Organization ID                           │
│      └── If member: Redirect to org dashboard              │
│      └── If not member: Show "Request Invitation" option │
└─────────────────────────────────────────────────────────────┘
```

### 4.2 Invitation Flow (New)

```
┌─────────────────────────────────────────────────────────────┐
│  INVITATION SENDER (Admin)                                 │
├─────────────────────────────────────────────────────────────┤
│  1. Go to Organization → Members → Invite                  │
│  2. Enter: Email, Role (admin/operator/viewer), Notes     │
│  3. Click "Send Invitation"                             │
│  4. Server sends email with link: /invite/:token         │
└─────────────────────────────────────────────────────────────┘

                              ↓

┌─────────────────────────────────────────────────────────────┐
│  INVITATION RECEIVER (Invitee)                             │
├─────────────────────────────────────────────────────────────┤
│  1. Receives email with invitation link                    │
│  2. Clicks link → /invite/:token page                     │
│                                                              │
│  3. If not logged in:                                     │
│     └── Redirect to login                                  │
│     └── After login, return to invitation page             │
│                                                              │
│  4. If logged in:                                         │
│     └── Show invitation dialog:                           │
│                                                              │
│     ┌─────────────────────────────────────────────────┐  │
│     │  You've been invited to [Org Name]                  │  │
│     │  Role: [Role]                                     │  │
│     │  Invited by: [Inviter Name]                       │  │
│     │  Notes: [Inviter notes]                           │  │
│     │                                                      │  │
│     │  [Approve]              [Reject]                   │  │
│     │  + Required notes         + Required notes         │  │
│     └─────────────────────────────────────────────────┘  │
│                                                              │
│  5. Approve:                                              │
│     └── Creates membership                                  │
│     └── Email sent to inviter: "[User] accepted invitation" │
│     └── Redirect to org dashboard                          │
│                                                              │
│  6. Reject:                                                │
│     └── Updates invitation status                          │
│     └── Email sent to inviter: "[User] declined"           │
│     └── Redirect to home                                  │
└─────────────────────────────────────────────────────────────┘
```

### 4.3 Organization Switch Flow

```
┌─────────────────────────────────────────────────────────────┐
│  ORG SWITCHING (Navbar or Dashboard Header)                  │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────────────────────────────────────────┐  │
│  │  🏢 Acme Corp [Super Admin]  ▼                        │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                              │
│  Click → Dropdown:                                        │
│  ┌─────────────────────────────────────────────────────┐  │
│  │  🏢 Acme Corp                           [Current ✓]  │  │
│  │  🏢 Personal Projects                      [Admin]  │  │
│  │  ──────────────────────────────────────────────────  │  │
│  │  + Create Organization                                  │  │
│  │  Leave Organization                                    │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                              │
│  Selecting different org:                                   │
│  └── Updates session context (X-Org-ID header)              │
│  └── Refreshes all queries with new org filter              │
│  └── Updates URL: /dashboard?org=org-uuid                   │
└─────────────────────────────────────────────────────────────┘
```

---

## PART 5: Component Details

### 5.1 OrganizationSelector Component

```typescript
// components/organization/OrganizationSelector.tsx
interface OrganizationSelectorProps {
  onCreateNew?: () => void;
}

const OrganizationSelector: React.FC<OrganizationSelectorProps> = ({ onCreateNew }) => {
  const { memberships } = useMemberships();
  const { selectedOrg, selectOrg } = useOrganization();
  
  const dropdownItems = memberships.map((m) => ({
    id: m.organization.id,
    label: m.organization.name,
    role: m.role,
    isSelected: m.organization.id === selectedOrg?.id,
    onClick: () => selectOrg(m.organization),
  }));
  
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" className="gap-2">
          <BuildingIcon />
          <span>{selectedOrg?.name || "Select Organization"}</span>
          <ChevronDownIcon />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        {dropdownItems.map((item) => (
          <DropdownMenuItem 
            key={item.id}
            onClick={item.onClick}
            className={item.isSelected ? "bg-accent" : ""}
          >
            {item.label}
            <Badge>{item.role}</Badge>
            {item.isSelected && <CheckIcon />}
          </DropdownMenuItem>
        ))}
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={onCreateNew}>
          <PlusIcon /> Create Organization
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
```

### 5.2 InviteMemberModal Component

```typescript
// components/organization/InviteMemberModal.tsx
interface InviteMemberModalProps {
  organizationId: string;
  onInvite: (invitation: Invitation) => void;
  onClose: () => void;
}

const InviteMemberModal: React.FC<InviteMemberModalProps> = ({ organizationId, onInvite, onClose }) => {
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<OrgRole>("operator");
  const [notes, setNotes] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  
  const { mutateAsync: inviteMember } = useInviteMember();
  
  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    try {
      const invitation = await inviteMember({
        organizationId,
        email,
        role,
        notes,
      });
      toast.success(`Invitation sent to ${email}`);
      onInvite(invitation);
      onClose();
    } catch (error) {
      toast.error(getErrorMessage(error));
    } finally {
      setIsLoading(false);
    }
  };
  
  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Invite Member</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="space-y-4">
            <div>
              <Label>Email</Label>
              <Input 
                type="email" 
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div>
              <Label>Role</Label>
              <Select value={role} onValueChange={setRole}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="admin">Admin</SelectItem>
                  <SelectItem value="operator">Operator</SelectItem>
                  <SelectItem value="viewer">Viewer</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Notes (optional)</Label>
              <Textarea 
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                placeholder="Add a personal message..."
              />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? "Sending..." : "Send Invitation"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};
```

### 5.3 Invitation Page (Public)

```typescript
// routes/invite/$token.tsx
const InvitationPage: React.FC = () => {
  const { token } = useParams();
  const { user } = useAuth();
  const { invitation, isLoading, error } = useInvitation(token);
  
  const [notes, setNotes] = useState("");
  const [action, setAction] = useState<"approve" | "reject" | null>(null);
  
  const { mutateAsync: acceptInvitation } = useAcceptInvitation();
  const { mutateAsync: rejectInvitation } = useRejectInvitation();
  
  const handleApprove = async () => {
    setAction("approve");
    try {
      await acceptInvitation({ token, notes });
      toast.success("You've joined the organization!");
      navigate({ to: `/dashboard?org=${invitation.organization.id}` });
    } catch (error) {
      toast.error(getErrorMessage(error));
      setAction(null);
    }
  };
  
  const handleReject = async () => {
    if (!notes.trim()) {
      toast.error("Please provide a reason for declining");
      return;
    }
    setAction("reject");
    try {
      await rejectInvitation({ token, notes });
      toast.info("Invitation declined");
      navigate({ to: "/dashboard" });
    } catch (error) {
      toast.error(getErrorMessage(error));
      setAction(null);
    }
  };
  
  if (!user) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Card className="max-w-md w-full p-6">
          <CardHeader>
            <CardTitle>Sign in Required</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground mb-4">
              Please sign in or create an account to view this invitation.
            </p>
            <Button asChild>
              <Link to="/login">Sign In</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }
  
  if (invitation?.email !== user.email) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Card className="max-w-md w-full p-6 border-destructive">
          <CardHeader>
            <CardTitle className="text-destructive">Wrong Account</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground">
              This invitation was sent to {invitation.email}, 
              but you're signed in as {user.email}. 
              Please sign in with the correct account.
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }
  
  return (
    <div className="flex items-center justify-center min-h-screen p-4">
      <Card className="max-w-lg w-full">
        <CardHeader>
          <CardTitle>You're Invited!</CardTitle>
          <CardDescription>
            You've been invited to join {invitation.organization.name}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-muted-foreground">Role:</span>
              <Badge>{invitation.role}</Badge>
            </div>
            <div>
              <span className="text-muted-foreground">Invited by:</span>
              <p>{invitation.invitedBy.name}</p>
            </div>
          </div>
          {invitation.inviterNotes && (
            <div>
              <span className="text-muted-foreground">Notes:</span>
              <p className="mt-1 p-3 bg-muted rounded-md">
                {invitation.inviterNotes}
              </p>
            </div>
          )}
          <div>
            <Label>Your response notes (required):</Label>
            <Textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Add a note (required for declining)..."
            />
          </div>
        </CardContent>
        <CardFooter className="gap-2">
          <Button 
            variant="outline" 
            className="flex-1"
            onClick={handleReject}
            disabled={action !== null}
          >
            {action === "reject" ? "Declining..." : "Decline"}
          </Button>
          <Button 
            className="flex-1"
            onClick={handleApprove}
            disabled={action !== null}
          >
            {action === "approve" ? "Joining..." : "Accept Invitation"}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
};
```

---

## PART 6: API Client Type Definitions

### 6.1 Organization Types

```typescript
// domain/organization/organization-entity.ts

export type OrgRole = "super_admin" | "admin" | "operator" | "viewer";

export interface Organization {
  id: string;
  name: string;
  createdBy: string;
  memberCount: number;
  createdAt: string;
  isActive: boolean;
}

export interface CreateOrganizationRequest {
  name: string;
  role: "admin" | "super_admin";
}

export interface UpdateOrganizationRequest {
  name?: string;
}

export interface OrganizationWithMembers extends Organization {
  members: OrganizationMembership[];
}
```

### 6.2 Membership Types

```typescript
// domain/membership/membership-entity.ts

import type { OrgRole, Organization } from "../organization";
import type { Operator } from "../auth";

export interface OrganizationMembership {
  id: string;
  organization: Organization;
  operator: Operator;
  role: OrgRole;
  joinedAt: string;
  invitedBy?: Operator;
  status: "active" | "removed";
}

export interface UpdateMemberRoleRequest {
  role: OrgRole;
}
```

### 6.3 Invitation Types

```typescript
// domain/invitation/invitation-entity.ts

import type { OrgRole, Organization } from "../organization";
import type { Operator } from "../auth";

export type InvitationStatus = "pending" | "approved" | "rejected" | "expired";

export interface Invitation {
  id: string;
  organization: Organization;
  email: string;
  role: OrgRole;
  status: InvitationStatus;
  token: string;
  inviterNotes?: string;
  inviteeNotes?: string;
  invitedBy: Operator;
  invitedAt: string;
  respondedAt?: string;
  expiresAt: string;
}

export interface CreateInvitationRequest {
  organizationId: string;
  email: string;
  role: OrgRole;
  notes?: string;
}

export interface InvitationResponse {
  token: string;
  notes?: string;
}
```

### 6.4 REST Endpoint Functions

```typescript
// vyzorServer/rest/organization/organization-endpoints.ts

import type { Organization, CreateOrganizationRequest, UpdateOrganizationRequest } from "@/domain/organization";
import type { OrganizationMembership, UpdateMemberRoleRequest } from "@/domain/membership";
import type { Invitation, CreateInvitationRequest, InvitationResponse } from "@/domain/invitation";

export async function createOrganization(
  request: CreateOrganizationRequest,
  baseUrl: string
): Promise<Organization> {
  const response = await fetch(`${baseUrl}/v1/organizations`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(request),
  });
  
  if (!response.ok) {
    throw new Error(`Failed to create organization: ${response.statusText}`);
  }
  
  return response.json();
}

export async function listOrganizations(baseUrl: string): Promise<Organization[]> {
  const response = await fetch(`${baseUrl}/v1/organizations`);
  
  if (!response.ok) {
    throw new Error(`Failed to list organizations: ${response.statusText}`);
  }
  
  return response.json();
}

export async function getOrganization(
  id: string,
  baseUrl: string
): Promise<Organization> {
  const response = await fetch(`${baseUrl}/v1/organizations/${id}`);
  
  if (!response.ok) {
    throw new Error(`Failed to get organization: ${response.statusText}`);
  }
  
  return response.json();
}

export async function deleteOrganization(
  id: string,
  baseUrl: string
): Promise<void> {
  const response = await fetch(`${baseUrl}/v1/organizations/${id}`, {
    method: "DELETE",
  });
  
  if (!response.ok) {
    throw new Error(`Failed to delete organization: ${response.statusText}`);
  }
}

// Member endpoints
export async function listMembers(
  organizationId: string,
  baseUrl: string
): Promise<OrganizationMembership[]> {
  const response = await fetch(`${baseUrl}/v1/organizations/${organizationId}/members`);
  
  if (!response.ok) {
    throw new Error(`Failed to list members: ${response.statusText}`);
  }
  
  return response.json();
}

export async function removeMember(
  organizationId: string,
  memberId: string,
  baseUrl: string
): Promise<void> {
  const response = await fetch(
    `${baseUrl}/v1/organizations/${organizationId}/members/${memberId}`,
    { method: "DELETE" }
  );
  
  if (!response.ok) {
    throw new Error(`Failed to remove member: ${response.statusText}`);
  }
}

export async function updateMemberRole(
  organizationId: string,
  memberId: string,
  request: UpdateMemberRoleRequest,
  baseUrl: string
): Promise<OrganizationMembership> {
  const response = await fetch(
    `${baseUrl}/v1/organizations/${organizationId}/members/${memberId}`,
    {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    }
  );
  
  if (!response.ok) {
    throw new Error(`Failed to update member role: ${response.statusText}`);
  }
  
  return response.json();
}

// Invitation endpoints
export async function createInvitation(
  request: CreateInvitationRequest,
  baseUrl: string
): Promise<Invitation> {
  const response = await fetch(`${baseUrl}/v1/invitations`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(request),
  });
  
  if (!response.ok) {
    throw new Error(`Failed to create invitation: ${response.statusText}`);
  }
  
  return response.json();
}

export async function getInvitationByToken(
  token: string,
  baseUrl: string
): Promise<Invitation> {
  const response = await fetch(`${baseUrl}/v1/invite/${token}`);
  
  if (!response.ok) {
    throw new Error(`Failed to get invitation: ${response.statusText}`);
  }
  
  return response.json();
}

export async function acceptInvitation(
  request: InvitationResponse,
  token: string,
  baseUrl: string
): Promise<Invitation> {
  const response = await fetch(`${baseUrl}/v1/invite/${token}/approve`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(request),
  });
  
  if (!response.ok) {
    throw new Error(`Failed to accept invitation: ${response.statusText}`);
  }
  
  return response.json();
}

export async function rejectInvitation(
  request: InvitationResponse,
  token: string,
  baseUrl: string
): Promise<Invitation> {
  const response = await fetch(`${baseUrl}/v1/invite/${token}/reject`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(request),
  });
  
  if (!response.ok) {
    throw new Error(`Failed to reject invitation: ${response.statusText}`);
  }
  
  return response.json();
}

// My memberships
export async function getMyMemberships(baseUrl: string): Promise<OrganizationMembership[]> {
  const response = await fetch(`${baseUrl}/v1/me/memberships`);
  
  if (!response.ok) {
    throw new Error(`Failed to get memberships: ${response.statusText}`);
  }
  
  return response.json();
}
```

---

## PART 7: Implementation Order

### Phase C1: API Client Foundation
1. Create `domain/organization` package
2. Create `domain/membership` package  
3. Create `domain/invitation` package
4. Create REST endpoints for organizations
5. Create REST endpoints for invitations
6. Update `domain/auth` with memberships field
7. Update GraphQL types for organizations

### Phase C2: Web App Foundation
1. Create `use-organizations` hook
2. Create `use-memberships` hook
3. Create `OrganizationContext`
4. Create `OrganizationSelector` component
5. Update `AppLayout` with org selector
6. Update routes to use org context

### Phase C3: Organization Management UI
1. Create organizations list page
2. Create organization creation modal
3. Create join organization modal
4. Create member management page
5. Create invitation management page

### Phase C4: Invitation Flow
1. Create invitation API endpoints
2. Create `useInvitation` hook
3. Create invitation page `/invite/:token`
4. Create invitation dialog components
5. Update OAuth callback with error handling

### Phase C5: Dashboard Integration
1. Update dashboard to filter by org
2. Add org switcher to navbar
3. Update device operations with org context
4. Add "no org" empty states

### Phase C6: Polish
1. Add loading states
2. Add error handling
3. Add confirmations for destructive actions
4. Update documentation
5. Test full flows

---

## Summary of All Files to Create

### API Client - Domain Layer (NEW)
```
packages/API_Client/src/domain/organization/
├── organization-entity.ts
├── organization-mappers.ts
├── organization-validators.ts
└── index.ts

packages/API_Client/src/domain/membership/
├── membership-entity.ts
├── membership-mappers.ts
└── index.ts

packages/API_Client/src/domain/invitation/
├── invitation-entity.ts
├── invitation-mappers.ts
├── invitation-validators.ts
└── index.ts
```

### API Client - REST Endpoints (NEW)
```
packages/API_Client/src/vyzorServer/rest/organization/
├── index.ts
├── organization-endpoints.ts
├── member-endpoints.ts
└── invitation-endpoints.ts

packages/API_Client/src/vyzorServer/rest/invitation/
├── index.ts
└── invitation-endpoints.ts

packages/API_Client/src/vyzorServer/rest/me/
├── index.ts
└── me-endpoints.ts
```

### API Client - GraphQL (NEW)
```
packages/API_Client/src/vyzorServer/graphql/organization/
├── index.ts
├── graphql-organization-types.ts
├── graphql-organization-queries.ts
├── graphql-organization-mutations.ts
└── graphql-organization-fragments.ts

packages/API_Client/src/vyzorServer/graphql/membership/
├── index.ts
├── graphql-membership-types.ts
├── graphql-membership-queries.ts
└── graphql-membership-mutations.ts

packages/API_Client/src/vyzorServer/graphql/invitation/
├── index.ts
├── graphql-invitation-types.ts
├── graphql-invitation-queries.ts
├── graphql-invitation-mutations.ts
└── graphql-invitation-subscriptions.ts
```

### Web App - Routes (NEW)
```
apps/web/src/routes/organizations/
├── index.tsx
├── new.tsx
├── $orgId.tsx
├── $orgId/index.tsx
├── $orgId/members.tsx
├── $orgId/invitations.tsx
└── $orgId/settings.tsx

apps/web/src/routes/invite/
├── $token.tsx
├── $token.accept.tsx
└── $token.reject.tsx

apps/web/src/routes/
├── join-organization.tsx
└── transfer-ownership.tsx
```

### Web App - Components (NEW)
```
apps/web/src/components/organization/
├── OrganizationSelector.tsx
├── OrganizationCard.tsx
├── CreateOrganizationModal.tsx
├── JoinOrganizationModal.tsx
├── MemberList.tsx
├── MemberRow.tsx
├── InviteMemberModal.tsx
├── InvitationList.tsx
└── TransferOwnershipModal.tsx

apps/web/src/components/invitation/
├── InvitationBanner.tsx
├── InvitationCard.tsx
├── AcceptInvitationModal.tsx
└── RejectInvitationModal.tsx

apps/web/src/components/dialogs/
├── EmailRequiredDialog.tsx
├── EmailNotVerifiedDialog.tsx
├── OAuthErrorDialog.tsx
└── RetryDialog.tsx

apps/web/src/components/layout/
├── OrgLayout.tsx
└── OrgSwitcher.tsx
```

### Web App - Hooks (NEW)
```
apps/web/src/hooks/
├── use-organization.ts
├── use-organizations.ts
├── use-memberships.ts
├── use-invitation.ts
├── use-pending-invitations.ts
└── use-org-permissions.ts
```

### Web App - Context (NEW)
```
apps/web/src/context/
├── OrganizationContext.tsx
└── InvitationContext.tsx
```

---

## Files to MODIFY

### API Client
| File | Changes |
|------|---------|
| `domain/auth/auth-entity.ts` | Add memberships, update types |
| `domain/auth/auth-mappers.ts` | Map memberships |
| `domain/device/device-entity.ts` | Add organizationId |
| `domain/oauth/oauth-entity.ts` | Add error codes |
| `domain/oauth/oauth-mappers.ts` | Handle email errors |
| `vyzorServer/index.ts` | Export new modules |

### Web App
| File | Changes |
|------|---------|
| `routes/__root.tsx` | Add org context provider |
| `routes/dashboard.tsx` | Filter by org, add org selector |
| `routes/auth.callback.tsx` | Handle OAuth errors |
| `routes/login.tsx` | Add join org option |
| `routes/create-account.tsx` | Redirect to create/join org |
| `routes/device.tsx` | Add org context |
| `components/layout/AppLayout.tsx` | Add org selector |
| `hooks/use-auth.ts` | Include memberships |
| `lib/api/graphql/index.ts` | Export new hooks |

---

## Total File Count

| Category | New Files | Modified Files |
|----------|-----------|---------------|
| API Client - Domain | 12 | 6 |
| API Client - REST | 8 | 1 |
| API Client - GraphQL | 15 | 2 |
| Web App - Routes | 10 | 6 |
| Web App - Components | 18 | 4 |
| Web App - Hooks | 6 | 3 |
| Web App - Context | 2 | 1 |
| **TOTAL** | **71** | **23** |
