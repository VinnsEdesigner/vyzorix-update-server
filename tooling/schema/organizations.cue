package vyzorix

// CreateOrganizationRequest — generated from openapi/schemas.go.
#CreateOrganizationRequest: {
        name: string
        description: string
        role: string
        maxMembers: int
}

// SelectOrganizationRequest — generated from openapi/schemas.go.
#SelectOrganizationRequest: {
        organization_id: string
}

// Organization — generated from openapi/schemas.go.
#Organization: {
        id: string
        name: string
        description?: string
        created_by: string
        created_at: string
        updated_at?: string
        max_members: int
        is_active: bool
        member_count?: int
}

// OrganizationListResult — generated from openapi/schemas.go.
#OrganizationListResult: {
        organizations: [...#Organization]
}

// SelectOrganizationResult — generated from openapi/schemas.go.
#SelectOrganizationResult: {
        organization_id: string
        organization_name: string
        role: string
}

// OrganizationInfo — generated from openapi/schemas.go.
#OrganizationInfo: {
        id: string
        name: string
        role: string
}

// OrganizationSettingsResult — generated from openapi/schemas.go.
#OrganizationSettingsResult: {
        defaultThresholds?: #OperatorThresholds
        id: string
        organizationId: string
        timezone: string
        dateFormat: string
        createdAt: string
        updatedAt: string
        alertCooldownMinutes: int
}

// UpdateOrganizationSettingsRequest — generated from openapi/schemas.go.
#UpdateOrganizationSettingsRequest: {
        timezone?: string
        dateFormat?: string
        alertCooldownMinutes?: int
        defaultThresholds?: #OperatorThresholds
}

// UpdateOrganizationRequest — generated from openapi/schemas.go.
#UpdateOrganizationRequest: {
        name?: string
        maxMembers?: int
        isActive?: bool
}
