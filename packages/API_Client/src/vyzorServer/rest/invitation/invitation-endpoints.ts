

import { restClient } from '../_shared/rest-client';
import type { Invitation, InvitationLifecycle, InvitationApiResponse } from '../../../domain/invitation';
import { mapInvitation } from '../../../domain/invitation';

export interface PublicInvitationResponse {
  organizationId: string;
  organizationName: string;
  email: string;
  role: string;
  status: InvitationLifecycle;
  invitedAt: string;
  expiresAt: string;
}

export interface InvitationApproveRequest {
  notes?: string;
}

export interface InvitationRejectRequest {
  notes?: string;
}

export async function getPublicInvitation(token: string): Promise<Invitation> {
  const response = await restClient.get<InvitationApiResponse>(`/v1/invite/${token}`);
  return mapInvitation(response);
}

export async function approveInvitation(
  token: string,
  request?: InvitationApproveRequest
): Promise<Invitation> {
  const response = await restClient.post<InvitationApiResponse>(
    `/v1/invite/${token}/approve`,
    request || {}
  );
  return mapInvitation(response);
}

export async function rejectInvitation(
  token: string,
  request?: InvitationRejectRequest
): Promise<Invitation> {
  const response = await restClient.post<InvitationApiResponse>(
    `/v1/invite/${token}/reject`,
    request || {}
  );
  return mapInvitation(response);
}
