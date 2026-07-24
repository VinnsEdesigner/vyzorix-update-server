/**
 * Mappers for converting invitation API responses to domain entities.
 */

import type {
  Invitation,
  InvitationListItem,
  InvitationApiResponse,
  InvitationLifecycle,
} from "./invitation-entity";
import type { OrganizationRole } from "../organization/organization-entity";

/**
 * Map API invitation response to domain entity.
 */
export function mapApiToInvitation(api: InvitationApiResponse): Invitation {
  return {
    id: api.id,
    organizationId: api.organization_id,
    email: api.email,
    role: api.role as OrganizationRole,
    lifecycle: api.lifecycle as InvitationLifecycle,
    invitedBy: api.invited_by,
    invitedAt: new Date(api.invited_at),
    expiresAt: new Date(api.expires_at),
    inviterNotes: api.inviter_notes,
    organizationName: api.organization_name,
    invitedByName: api.invited_by_name,
  };
}

/**
 * Map API response to invitation list item (lightweight).
 */
export function mapApiToInvitationListItem(
  api: InvitationApiResponse
): InvitationListItem {
  return {
    id: api.id,
    organizationId: api.organization_id,
    organizationName: api.organization_name ?? "Unknown",
    email: api.email,
    role: api.role as OrganizationRole,
    lifecycle: api.lifecycle as InvitationLifecycle,
    invitedAt: new Date(api.invited_at),
    expiresAt: new Date(api.expires_at),
  };
}

/**
 * Map array of API responses to invitation list.
 */
export function mapApiToInvitationList(
  apiList: InvitationApiResponse[]
): InvitationListItem[] {
  return apiList.map(mapApiToInvitationListItem);
}
