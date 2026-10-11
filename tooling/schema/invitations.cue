package vyzorix

// CreateInvitationRequest.
#CreateInvitationRequest: {
        email: string
        role: string
        org_id?: string
}

// Invitation.
#Invitation: {
        id: string
        organization_id: string
        email: string
        role: string
        status: string
        token?: string
        invited_by?: string
        invited_at?: int64
        responded_at?: int64
        responder_id?: string
        organization_name?: string
        inviter_name?: string
        expires_at?: int64
}

// InvitationListResult.
#InvitationListResult: {
        invitations: [...#Invitation]
}

// InvitationByTokenResult.
#InvitationByTokenResult: {
        id: string
        organization_id: string
        organization_name?: string
        email: string
        role: string
        status: string
        invited_at?: int64
        inviter_name?: string
        expires_at?: int64
}
