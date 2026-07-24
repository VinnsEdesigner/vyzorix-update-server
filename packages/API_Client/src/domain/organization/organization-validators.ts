import type {
  OrganizationRole,
  CreateOrganizationRequest,
  UpdateOrganizationRequest,
  CreateInvitationRequest,
  UpdateMemberRoleRequest,
} from "./organization-entity";

const VALID_ROLES: OrganizationRole[] = [
  "super_admin",
  "admin",
  "operator",
  "viewer",
];

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function isValidRole(role: string): role is OrganizationRole {
  return VALID_ROLES.includes(role as OrganizationRole);
}

export function isValidEmail(email: string): boolean {
  return EMAIL_REGEX.test(email);
}

export function validateCreateOrganization(
  req: CreateOrganizationRequest
): { valid: boolean; errors: string[] } {
  const errors: string[] = [];
  if (!req.name || req.name.trim().length === 0) {
    errors.push("name is required");
  }
  if (req.name && req.name.length > 100) {
    errors.push("name must be at most 100 characters");
  }
  if (!req.description || req.description.trim().length === 0) {
    errors.push("description is required");
  }
  if (req.role !== "super_admin" && req.role !== "admin") {
    errors.push("role must be super_admin or admin");
  }
  return { valid: errors.length === 0, errors };
}

export function validateUpdateOrganization(
  req: UpdateOrganizationRequest
): { valid: boolean; errors: string[] } {
  const errors: string[] = [];
  if (req.name !== undefined && req.name.trim().length === 0) {
    errors.push("name cannot be empty");
  }
  if (req.maxMembers !== undefined && req.maxMembers < 1) {
    errors.push("maxMembers must be at least 1");
  }
  return { valid: errors.length === 0, errors };
}

export function validateCreateInvitation(
  req: CreateInvitationRequest
): { valid: boolean; errors: string[] } {
  const errors: string[] = [];
  if (!req.organizationId) {
    errors.push("organizationId is required");
  }
  if (!req.email || !isValidEmail(req.email)) {
    errors.push("valid email is required");
  }
  if (!isValidRole(req.role)) {
    errors.push("invalid role");
  }
  return { valid: errors.length === 0, errors };
}

export function validateUpdateMemberRole(
  req: UpdateMemberRoleRequest
): { valid: boolean; errors: string[] } {
  const errors: string[] = [];
  if (!isValidRole(req.role)) {
    errors.push("invalid role");
  }
  return { valid: errors.length === 0, errors };
}
