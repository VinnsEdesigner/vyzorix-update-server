/**
 * Validation functions for invitation domain.
 */

import type {
  CreateInvitationRequest,
  InvitationResponseRequest,
} from "./invitation-entity";
import { isValidEmail } from "../organization/organization-validators";
import { isValidRole } from "../organization/organization-validators";

const MAX_NOTES_LENGTH = 500;

/**
 * Validate create invitation request.
 */
export function validateCreateInvitation(
  req: CreateInvitationRequest
): { valid: boolean; errors: string[] } {
  const errors: string[] = [];

  // Organization ID validation
  if (!req.organizationId || req.organizationId.trim().length === 0) {
    errors.push("Organization ID is required");
  }

  // Email validation
  if (!req.email || !isValidEmail(req.email)) {
    errors.push("Valid email address is required");
  }

  // Role validation
  if (!isValidRole(req.role)) {
    errors.push("Invalid role specified");
  }

  // Inviter notes validation
  if (req.inviterNotes && req.inviterNotes.length > MAX_NOTES_LENGTH) {
    errors.push(`Inviter notes cannot exceed ${MAX_NOTES_LENGTH} characters`);
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Validate invitation response request.
 */
export function validateInvitationResponse(
  req: InvitationResponseRequest
): { valid: boolean; errors: string[] } {
  const errors: string[] = [];

  if (req.notes && req.notes.length > MAX_NOTES_LENGTH) {
    errors.push(`Notes cannot exceed ${MAX_NOTES_LENGTH} characters`);
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Validate invitation token format.
 */
export function isValidInvitationToken(token: string): boolean {
  // Token should be a non-empty string with reasonable length
  return token.length > 0 && token.length < 500;
}
