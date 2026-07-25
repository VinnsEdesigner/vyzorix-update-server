

export const INVITATION_FRAGMENT = `
  fragment InvitationFields on Invitation {
    id
    email
    role
    status
    token
    notes
    invitedAt
    respondedAt
    expiresAt
    organization {
      id
      name
    }
    invitedBy {
      id
      email
      name
    }
  }
`;

export interface GQLInvitation {
  id: string;
  email: string;
  role: string;
  status: 'pending' | 'approved' | 'rejected' | 'expired';
  token?: string;
  notes?: string;
  invitedAt: string;
  respondedAt?: string;
  expiresAt: string;
  organization: {
    id: string;
    name: string;
  };
  invitedBy: {
    id: string;
    email: string;
    name: string;
  };
}
