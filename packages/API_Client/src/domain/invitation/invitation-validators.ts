import type { InvitationResponseRequest } from "./invitation-entity";

export function validateInvitationResponse(
  req: InvitationResponseRequest
): { valid: boolean; errors: string[] } {
  const errors: string[] = [];
  if (req.notes && req.notes.length > 500) {
    errors.push("notes cannot exceed 500 characters");
  }
  return { valid: errors.length === 0, errors };
}

export function isValidInvitationToken(token: string): boolean {
  return token.length > 0 && token.length < 500;
}
