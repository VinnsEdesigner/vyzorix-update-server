import { gql } from "@apollo/client";

export const ORGANIZATION_FRAGMENT = gql`
  fragment OrganizationFields on Organization {
    id
    name
    lifecycle
    maxMembers
    memberCount
    createdAt
    updatedAt
    deletedAt
    createdBy
  }
`;

export const MEMBERSHIP_FRAGMENT = gql`
  fragment MembershipFields on Membership {
    id
    organizationId
    operatorId
    role
    lifecycle
    invitedAt
    joinedAt
    removedAt
    suspendedAt
    operator {
      id
      email
      name
    }
    organization {
      id
      name
    }
  }
`;

export const INVITATION_FRAGMENT = gql`
  fragment InvitationFields on Invitation {
    id
    organizationId
    email
    role
    status
    token
    notes
    invitedAt
    respondedAt
    expiresAt
    inviter {
      id
      name
    }
    organization {
      id
      name
    }
  }
`;

export interface Organization {
  id: string;
  name: string;
  lifecycle: "active" | "inactive" | "archived";
  maxMembers: number;
  memberCount: number;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string;
  createdBy: string;
}

export interface Membership {
  id: string;
  organizationId: string;
  operatorId: string;
  role: "super_admin" | "admin" | "operator" | "viewer";
  lifecycle: "invited" | "active" | "suspended" | "removed";
  invitedAt?: string;
  joinedAt?: string;
  removedAt?: string;
  suspendedAt?: string;
  operator?: {
    id: string;
    email: string;
    name: string;
  };
  organization?: {
    id: string;
    name: string;
  };
}

export interface Invitation {
  id: string;
  organizationId: string;
  email: string;
  role: "super_admin" | "admin" | "operator" | "viewer";
  status: "pending" | "approved" | "rejected" | "expired";
  token?: string;
  notes?: string;
  invitedAt?: string;
  respondedAt?: string;
  expiresAt?: string;
  inviter?: {
    id: string;
    name: string;
  };
  organization?: {
    id: string;
    name: string;
  };
}

export interface OrganizationListResponse {
  items: Organization[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasMore: boolean;
  };
}

export interface MemberListResponse {
  items: Membership[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasMore: boolean;
  };
}

export interface InvitationListResponse {
  items: Invitation[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasMore: boolean;
  };
}
