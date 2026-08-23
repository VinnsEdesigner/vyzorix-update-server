// Auth domain — generated types + zod structural validation + hand-rolled
// business rules. Entity types and Raw mappers eliminated.
import {
  LoginRequestSchema,
  RegisterRequestSchema,
  MFAVerifyRequestSchema,
} from '../../generated/vyzorixUpdateServerAPI.zod';
import type {
  LoginRequest,
  LoginResult,
  LoginWithTokensResult,
  RegisterRequest,
  RegisterResult,
  MeResult,
  RefreshTokenRequest,
  RefreshTokenResult,
  OrganizationInfo,
  MFAVerifyRequest,
  MFAVerifyResult,
} from '../../generated/vyzorixUpdateServerAPI.schemas';

export type {
  LoginRequest,
  LoginResult,
  LoginWithTokensResult,
  RegisterRequest,
  RegisterResult,
  MeResult,
  RefreshTokenRequest,
  RefreshTokenResult,
  OrganizationInfo,
  MFAVerifyRequest,
  MFAVerifyResult,
};

// ---- Constants (hand-rolled, not in OpenAPI) ----

export type OperatorRole = 'super_admin' | 'admin' | 'operator' | 'viewer';

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const MFA_CODE_REGEX = /^\d{6}$/;

// ---- Login validator (business rules on generated zod) ----

export const loginValidator = LoginRequestSchema
  .refine((r) => r.email && EMAIL_REGEX.test(r.email), {
    message: 'Invalid email format',
    path: ['email'],
  })
  .refine((r) => r.password && r.password.length >= 8, {
    message: 'Password must be at least 8 characters',
    path: ['password'],
  });

export function validateLogin(input: unknown) {
  return loginValidator.safeParse(input);
}

// ---- Register validator ----

export const registerValidator = RegisterRequestSchema
  .refine((r) => r.email && EMAIL_REGEX.test(r.email), {
    message: 'Invalid email format',
    path: ['email'],
  })
  .refine((r) => r.password && r.password.length >= 8, {
    message: 'Password must be at least 8 characters',
    path: ['password'],
  })
  .refine((r) => r.name && r.name.trim().length >= 2, {
    message: 'Name must be at least 2 characters',
    path: ['name'],
  });

export function validateRegister(input: unknown) {
  return registerValidator.safeParse(input);
}

// ---- MFA validator ----

export function validateMFACode(code: string): boolean {
  return code !== undefined && MFA_CODE_REGEX.test(code);
}

export const mfaVerifyValidator = MFAVerifyRequestSchema
  .refine((r) => r.code && MFA_CODE_REGEX.test(r.code), {
    message: 'MFA code must be 6 digits',
    path: ['code'],
  });

// ---- Individual field validators (for form-level validation) ----

export function validateEmail(email: string): string | null {
  if (!email) return 'Email is required';
  if (!EMAIL_REGEX.test(email)) return 'Invalid email format';
  return null;
}

export function validatePassword(password: string): string | null {
  if (!password) return 'Password is required';
  if (password.length < 8) return 'Password must be at least 8 characters';
  return null;
}

export function validateName(name: string): string | null {
  if (!name || name.trim().length < 2) return 'Name must be at least 2 characters';
  return null;
}
