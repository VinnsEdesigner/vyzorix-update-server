/**
 * Invitation domain types aligned with server models.
 */

import type { OrganizationRole } from "../organization/organization-entity";

// Invitation lifecycle states
export type InvitationLifecycle = "pending" | "approved" | "rejected" | "expired";

// Default invitation expiry in hours
export const DEFAULT_INVITATION_EXPIRY_HOURS = 168; // 7 days

/**
 * Invitation entity.
 */
export interface Invitation {
  id: string;
  organizationId: string;
  email: string;
  role: OrganizationRole;
  lifecycle: InvitationLifecycle;
  invitedBy: string;
  invitedAt: Date;
  expiresAt: Date;
  inviterNotes?: string;
  // Populated fields
  organizationName?: string;
  invitedByName?: string;
}

/**
 * Invitation list item (lightweight).
 */
export interface InvitationListItem {
  id: string;
  organizationId: string;
  organizationName: string;
  email: string;
  role: OrganizationRole;
  lifecycle: InvitationLifecycle;
  invitedAt: Date;
  expiresAt: Date;
}

/**
 * Create invitation request.
 */
export interface CreateInvitationRequest {
  organizationId: string;
  email: string;
  role: OrganizationRole;
  inviterNotes?: string;
}

/**
 * Invitation response request (accept/reject).
 */
export interface InvitationResponseRequest {
  notes?: string;
}

// API response types (raw from server)
export interface InvitationApiResponse {
  id: string;
  organization_id: string;
  email: string;
  role: string;
  lifecycle: string;
  invited_by: string;
  invited_at: string;
  expires_at: string;
  inviter_notes?: string;
  organization_name?: string;
  invited_by_name?: string;
}

/**
 * Check if invitation is pending.
 */
export function isInvitationPending(invitation: Invitation): boolean {
  return invitation.lifecycle === "pending";
}

/**
 * Check if invitation is expired.
 */
export function isInvitationExpired(invitation: Invitation): boolean {
  return new Date() > invitation.expiresAt;
}

/**
 * Check if invitation can be responded to.
 */
export function canRespondToInvitation(invitation: Invitation): boolean {
  return isInvitationPending(invitation) && !isInvitationExpired(invitation);
}
