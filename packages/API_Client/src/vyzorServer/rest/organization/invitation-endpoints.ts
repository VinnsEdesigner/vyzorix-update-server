/**
 * Organization invitation REST API endpoints.
 */

import { restClient } from "../_shared/rest-client";
import {
  mapApiToInvitation,
  mapApiToInvitationListItem,
  type InvitationApiResponse,
  type Invitation,
  type InvitationListItem,
  type CreateInvitationRequest,
  type InvitationResponseRequest,
} from "@/domain/invitation";

const PATHS = {
  invitations: "/v1/invitations",
  invitation: (token: string) => `/v1/invite/${token}`,
  invitationApprove: (token: string) => `/v1/invite/${token}/approve`,
  invitationReject: (token: string) => `/v1/invite/${token}/reject`,
} as const;

export const invitations = {
  /**
   * List pending invitations (as inviter).
   */
  async list(): Promise<InvitationListItem[]> {
    const response = await restClient.get<InvitationApiResponse[]>(PATHS.invitations);
    return response.map(mapApiToInvitationListItem);
  },

  /**
   * Create an invitation.
   */
  async create(request: CreateInvitationRequest): Promise<Invitation> {
    const response = await restClient.post<InvitationApiResponse>(
      PATHS.invitations,
      request
    );
    return mapApiToInvitation(response);
  },

  /**
   * Get invitation by token (public endpoint).
   */
  async getByToken(token: string): Promise<Invitation | null> {
    const response = await restClient.get<InvitationApiResponse | null>(
      PATHS.invitation(token)
    );
    if (!response) return null;
    return mapApiToInvitation(response);
  },

  /**
   * Accept an invitation.
   */
  async accept(
    token: string,
    request?: InvitationResponseRequest
  ): Promise<Invitation> {
    const response = await restClient.post<InvitationApiResponse>(
      PATHS.invitationApprove(token),
      request ?? {}
    );
    return mapApiToInvitation(response);
  },

  /**
   * Reject an invitation.
   */
  async reject(
    token: string,
    request?: InvitationResponseRequest
  ): Promise<Invitation> {
    const response = await restClient.post<InvitationApiResponse>(
      PATHS.invitationReject(token),
      request ?? {}
    );
    return mapApiToInvitation(response);
  },

  /**
   * Cancel an invitation (as inviter).
   */
  async cancel(invitationId: string): Promise<{ success: boolean }> {
    return restClient.delete<{ success: boolean }>(
      `${PATHS.invitations}/${invitationId}`
    );
  },
};
