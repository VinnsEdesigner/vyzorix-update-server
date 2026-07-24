/**
 * Validation functions for organization domain.
 */

import type {
  OrganizationRole,
  CreateOrganizationRequest,
  UpdateOrganizationRequest,
  CreateMemberRequest,
  UpdateMemberRoleRequest,
} from "./organization-entity";

const VALID_ROLES: OrganizationRole[] = [
  "super_admin",
  "admin",
  "operator",
  "viewer",
];

const MIN_NAME_LENGTH = 2;
const MAX_NAME_LENGTH = 100;
const MIN_PASSWORD_LENGTH = 8;
const MAX_MEMBERS_LIMIT = 1000;

/**
 * Validate organization role string.
 */
export function isValidRole(role: string): role is OrganizationRole {
  return VALID_ROLES.includes(role as OrganizationRole);
}

/**
 * Validate create organization request.
 */
export function validateCreateOrganization(
  req: CreateOrganizationRequest
): { valid: boolean; errors: string[] } {
  const errors: string[] = [];

  // Name validation
  if (!req.name || req.name.trim().length === 0) {
    errors.push("Organization name is required");
  } else {
    if (req.name.length < MIN_NAME_LENGTH) {
      errors.push(`Organization name must be at least ${MIN_NAME_LENGTH} characters`);
    }
    if (req.name.length > MAX_NAME_LENGTH) {
      errors.push(`Organization name must be at most ${MAX_NAME_LENGTH} characters`);
    }
  }

  // Max members validation
  if (req.maxMembers !== undefined) {
    if (req.maxMembers < 1) {
      errors.push("Max members must be at least 1");
    }
    if (req.maxMembers > MAX_MEMBERS_LIMIT) {
      errors.push(`Max members cannot exceed ${MAX_MEMBERS_LIMIT}`);
    }
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Validate update organization request.
 */
export function validateUpdateOrganization(
  req: UpdateOrganizationRequest
): { valid: boolean; errors: string[] } {
  const errors: string[] = [];

  // Name validation (if provided)
  if (req.name !== undefined) {
    if (req.name.trim().length === 0) {
      errors.push("Organization name cannot be empty");
    } else {
      if (req.name.length < MIN_NAME_LENGTH) {
        errors.push(`Organization name must be at least ${MIN_NAME_LENGTH} characters`);
      }
      if (req.name.length > MAX_NAME_LENGTH) {
        errors.push(`Organization name must be at most ${MAX_NAME_LENGTH} characters`);
      }
    }
  }

  // Max members validation
  if (req.maxMembers !== undefined) {
    if (req.maxMembers < 1) {
      errors.push("Max members must be at least 1");
    }
    if (req.maxMembers > MAX_MEMBERS_LIMIT) {
      errors.push(`Max members cannot exceed ${MAX_MEMBERS_LIMIT}`);
    }
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Validate email format.
 */
export function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

/**
 * Validate create member request.
 */
export function validateCreateMember(
  req: CreateMemberRequest
): { valid: boolean; errors: string[] } {
  const errors: string[] = [];

  // Email validation
  if (!req.email || !isValidEmail(req.email)) {
    errors.push("Valid email address is required");
  }

  // Role validation
  if (!isValidRole(req.role)) {
    errors.push(`Invalid role. Must be one of: ${VALID_ROLES.join(", ")}`);
  }

  // Inviter notes length validation
  if (req.inviterNotes && req.inviterNotes.length > 500) {
    errors.push("Inviter notes cannot exceed 500 characters");
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Validate update member role request.
 */
export function validateUpdateMemberRole(
  req: UpdateMemberRoleRequest
): { valid: boolean; errors: string[] } {
  const errors: string[] = [];

  if (!isValidRole(req.role)) {
    errors.push(`Invalid role. Must be one of: ${VALID_ROLES.join(", ")}`);
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Sanitize organization name for display.
 */
export function sanitizeOrganizationName(name: string): string {
  return name.trim().replace(/\s+/g, " ");
}
