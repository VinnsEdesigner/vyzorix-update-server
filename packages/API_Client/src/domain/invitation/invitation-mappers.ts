import type { Invitation, InvitationApiResponse, InvitationLifecycle } from "./invitation-entity";
import type { OrganizationRole } from "../organization/organization-entity";

export function mapInvitation(raw: InvitationApiResponse): Invitation {
  return {
    id: raw.id,
    organization_id: raw.organization_id,
    organization_name: raw.organization_name,
    email: raw.email,
    role: raw.role as OrganizationRole,
    status: raw.status as InvitationLifecycle,
    token: raw.token,
    invited_by: raw.invited_by,
    inviter_name: raw.inviter_name,
    invited_at: raw.invited_at,
    responded_at: raw.responded_at,
    expires_at: raw.expires_at,
    notes: raw.notes,
  };
}
