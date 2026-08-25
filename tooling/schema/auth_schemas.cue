package vyzorix

// UpdatePushRequest is the request body for POST /v1/updates/push.
#UpdatePushRequest: {
        scheduledAt?: int64
        version:      string
        installType:  string
        deviceIds:    [...string]
}

// OrganizationInfo is a minimal org reference in auth responses.
#OrganizationInfo: {
        id:   string
        name: string
        role: string
}

// SelectOrganizationRequest is the body for POST /v1/auth/organizations/select.
#SelectOrganizationRequest: {
        organization_id: string
}

// SelectOrganizationResult is the response for org selection.
#SelectOrganizationResult: {
        organization_id:   string
        organization_name: string
        role:              string
}
