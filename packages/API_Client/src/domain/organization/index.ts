// Organization domain — generated types + hand-rolled business rules.
import { CreateOrganizationRequestSchema } from '../../generated/vyzorixUpdateServerAPI.zod';
import type {
  Organization,
  OrganizationListResult,
  OrganizationMember,
  OrganizationMemberListResult,
  CreateOrganizationRequest,
  UpdateOrganizationRequest,
  SelectOrganizationRequest,
  SelectOrganizationResult,
  UpdateMemberRoleRequest,
  MessageResult,
  OrganizationSettingsResult,
  UpdateOrganizationSettingsRequest,
} from '../../generated/vyzorixUpdateServerAPI.schemas';

export type {
  Organization,
  OrganizationListResult,
  OrganizationMember,
  OrganizationMemberListResult,
  CreateOrganizationRequest,
  UpdateOrganizationRequest,
  SelectOrganizationRequest,
  SelectOrganizationResult,
  UpdateMemberRoleRequest,
  MessageResult,
  OrganizationSettingsResult,
  UpdateOrganizationSettingsRequest,
};

// ---- Constants (hand-rolled) ----

export type OrganizationRole = 'super_admin' | 'admin' | 'operator' | 'viewer';
export type MemberLifecycle = 'active' | 'suspended' | 'removed';

const VALID_ROLES: OrganizationRole[] = ['super_admin', 'admin', 'operator', 'viewer'];
const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

// ---- Validators (business rules on generated zod) ----

export const createOrganizationValidator = CreateOrganizationRequestSchema
  .refine((r) => r.name && r.name.trim().length > 0, {
    message: 'name is required',
    path: ['name'],
  })
  .refine((r) => !r.name || r.name.length <= 100, {
    message: 'name must be at most 100 characters',
    path: ['name'],
  })
  .refine((r) => r.description && r.description.trim().length > 0, {
    message: 'description is required',
    path: ['description'],
  })
  .refine((r) => r.role === 'super_admin' || r.role === 'admin', {
    message: 'role must be super_admin or admin',
    path: ['role'],
  });

export function validateCreateOrganization(input: unknown) {
  return createOrganizationValidator.safeParse(input);
}

// ---- Field validators ----

export function isValidRole(role: string): role is OrganizationRole {
  return VALID_ROLES.includes(role as OrganizationRole);
}

export function isValidEmail(email: string): boolean {
  return EMAIL_REGEX.test(email);
}
