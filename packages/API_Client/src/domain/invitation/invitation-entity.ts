import type { OrganizationRole } from "../organization/organization-entity";

export type InvitationLifecycle = "pending" | "approved" | "rejected" | "expired";

export interface Invitation {
  id: string;
  organization_id: string;
  organization_name?: string;
  email: string;
  role: OrganizationRole;
  status: InvitationLifecycle;
  token?: string;
  invited_by: string;
  inviter_name?: string;
  invited_at: string;
  responded_at?: string;
  expires_at: string;
  notes?: string;
}

export interface InvitationResponseRequest {
  notes?: string;
}

export interface InvitationApiResponse {
  id: string;
  organization_id: string;
  organization_name?: string;
  email: string;
  role: string;
  status: string;
  token?: string;
  invited_by: string;
  inviter_name?: string;
  invited_at: string;
  responded_at?: string;
  expires_at: string;
  notes?: string;
}
