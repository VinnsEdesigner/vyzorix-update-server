/**
 * Organization domain package.
 * Multi-tenant organization model with roles and memberships.
 */

// Entity types
export type {
  Organization,
  OrganizationListItem,
  OrganizationMember,
  MemberListItem,
  OrganizationLifecycle,
  OrganizationRole,
  MemberLifecycle,
  CreateOrganizationRequest,
  UpdateOrganizationRequest,
  CreateMemberRequest,
  UpdateMemberRoleRequest,
  OrganizationApiResponse,
  MemberApiResponse,
} from "./organization-entity";

// Utility functions
export {
  ROLE_LEVELS,
  getRoleLevel,
  isAdminRole,
  isSuperAdminRole,
  canAcceptMembers,
  isOrganizationDeleted,
  canMemberAccessResources,
  canManageMember,
} from "./organization-entity";

// Mappers
export {
  mapApiToOrganization,
  mapApiToOrganizationListItem,
  mapApiToMember,
  mapApiToMemberListItem,
  mapApiToOrganizationList,
  mapApiToMemberList,
} from "./organization-mappers";

// Validators
export {
  isValidRole,
  validateCreateOrganization,
  validateUpdateOrganization,
  validateCreateMember,
  validateUpdateMemberRole,
  isValidEmail,
  sanitizeOrganizationName,
} from "./organization-validators";
