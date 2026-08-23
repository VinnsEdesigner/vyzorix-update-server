// Invitations domain — generated types + zod validation.
import { CreateInvitationRequestSchema } from '../../generated/vyzorixUpdateServerAPI.zod';
import type {
  Invitation,
  InvitationListResult,
  InvitationByTokenResult,
  CreateInvitationRequest,
} from '../../generated/vyzorixUpdateServerAPI.schemas';

export type {
  Invitation,
  InvitationListResult,
  InvitationByTokenResult,
  CreateInvitationRequest,
};

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const VALID_ROLES = ['super_admin', 'admin', 'operator', 'viewer'];

export const createInvitationValidator = CreateInvitationRequestSchema
  .refine((r) => r.email && EMAIL_REGEX.test(r.email), {
    message: 'Invalid email format',
    path: ['email'],
  })
  .refine((r) => r.role && VALID_ROLES.includes(r.role), {
    message: `Role must be one of: ${VALID_ROLES.join(', ')}`,
    path: ['role'],
  });

export function validateCreateInvitation(input: unknown) {
  return createInvitationValidator.safeParse(input);
}
