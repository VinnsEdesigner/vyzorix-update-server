package vyzorix

// OrganizationMember — generated from openapi/schemas.go.
#OrganizationMember: {
        id: string
        organization_id: string
        operator_id: string
        role: string
        invited_by?: string
        joined_at?: string
        removed_at?: string
        status: string
        operator_name?: string
        operator_email?: string
}

// OrganizationMemberListResult — generated from openapi/schemas.go.
#OrganizationMemberListResult: {
        members: [...#OrganizationMember]
}

// UpdateMemberRoleRequest — generated from openapi/schemas.go.
#UpdateMemberRoleRequest: {
        role: string
}
