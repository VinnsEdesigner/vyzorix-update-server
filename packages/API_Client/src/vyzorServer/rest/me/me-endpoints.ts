/**
 * Current user ("me") REST API endpoints.
 */

import { restClient } from "../_shared/rest-client";
import type { MemberApiResponse } from "@/domain/organization";
import type { OrganizationRole, MemberLifecycle } from "@/domain/organization";

const PATHS = {
  memberships: "/v1/me/memberships",
} as const;

export interface MembershipInfo {
  id: string;
  organizationId: string;
  organizationName: string;
  operatorId: string;
  role: OrganizationRole;
  lifecycle: MemberLifecycle;
  joinedAt: Date;
  invitedBy?: string;
  invitedByName?: string;
}

interface RawMembershipResponse {
  id: string;
  organization_id: string;
  organization_name?: string;
  operator_id: string;
  role: string;
  lifecycle: string;
  joined_at: string;
  invited_by?: string;
  invited_by_name?: string;
}

function mapRawToMembership(raw: RawMembershipResponse): MembershipInfo {
  return {
    id: raw.id,
    organizationId: raw.organization_id,
    organizationName: raw.organization_name ?? "Unknown",
    operatorId: raw.operator_id,
    role: raw.role as OrganizationRole,
    lifecycle: raw.lifecycle as MemberLifecycle,
    joinedAt: new Date(raw.joined_at),
    invitedBy: raw.invited_by,
    invitedByName: raw.invited_by_name,
  };
}

export const me = {
  /**
   * Get current user's organization memberships.
   */
  async memberships(): Promise<MembershipInfo[]> {
    const response = await restClient.get<RawMembershipResponse[]>(PATHS.memberships);
    return response.map(mapRawToMembership);
  },
};
