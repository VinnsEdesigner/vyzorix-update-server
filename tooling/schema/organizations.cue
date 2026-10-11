package vyzorix

// CreateOrganizationRequest.
#CreateOrganizationRequest: {
        name: string
        description: string
        role: string
        maxMembers: int
}

// SelectOrganizationRequest.
#SelectOrganizationRequest: {
        organization_id: string
}

// Organization.
#Organization: {
        id: string
        name: string
        description?: string
        created_by: string
        created_at: int64
        updated_at?: int64
        max_members: int
        is_active: bool
        member_count?: int
}

// OrganizationListResult.
#OrganizationListResult: {
        organizations: [...#Organization]
}

// SelectOrganizationResult.
#SelectOrganizationResult: {
        organization_id: string
        organization_name: string
        role: string
}

// OrganizationInfo.
#OrganizationInfo: {
        id: string
        name: string
        role: string
}

// OrganizationSettingsResult.
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

// UpdateOrganizationSettingsRequest.
#UpdateOrganizationSettingsRequest: {
        timezone?: string
        dateFormat?: string
        alertCooldownMinutes?: int
        defaultThresholds?: #OperatorThresholds
}

// UpdateOrganizationRequest.
#UpdateOrganizationRequest: {
        name?: string
        maxMembers?: int
        isActive?: bool
}
