package vyzorix

// CreateInvitationRequest — generated from openapi/schemas.go.
#CreateInvitationRequest: {
        email: string
        role: string
        org_id?: string
}

// Invitation — generated from openapi/schemas.go.
#Invitation: {
        id: string
        organization_id: string
        email: string
        role: string
        status: string
        token?: string
        invited_by?: string
        invited_at?: string
        expires_at?: string
}

// InvitationListResult — generated from openapi/schemas.go.
#InvitationListResult: {
        invitations: [...#Invitation]
}

// InvitationByTokenResult — generated from openapi/schemas.go.
#InvitationByTokenResult: {
        id: string
        organization_id: string
        organization_name?: string
        email: string
        role: string
        status: string
        invited_at?: string
        inviter_name?: string
        expires_at?: string
}
