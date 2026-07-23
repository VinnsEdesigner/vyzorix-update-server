package operator

// AuthResponse is returned on successful login or registration.
type AuthResponse struct {
    Token     string           `json:"token"`
    Operator  OperatorResponse `json:"operator"`
    ExpiresAt int64            `json:"expiresAt"`
}

// OperatorResponse is the safe JSON representation returned to clients.
type OperatorResponse struct {
    Thresholds    *Thresholds     `json:"thresholds,omitempty"`
    Client        *ClientSettings `json:"client,omitempty"`
    ID            string          `json:"id"`
    Email         string          `json:"email"`
    Name          string          `json:"name"`
    NeedsOrganization     bool    `json:"needs_organization"`
    Organizations         []OrganizationInfo `json:"organizations,omitempty"`
    LastOrganizationID    string           `json:"last_organization_id,omitempty"`
    SelectedOrganization  *OrganizationInfo `json:"selected_organization,omitempty"`
    CreatedAt     int64           `json:"createdAt"`
    EmailVerified bool            `json:"emailVerified,omitempty"`
}

// OrganizationInfo represents an organization with the operator's role in it.
type OrganizationInfo struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    Role string `json:"role"`
}

// AuthErrorResponse is the standard error envelope for auth endpoints.
type AuthErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message,omitempty"`
}

// ToResponse converts an Operator to its safe JSON representation.
// Note: Role is not included as it's now org-scoped only.
func (o *Operator) ToResponse() OperatorResponse {
    // Build organization info from memberships.
    var orgs []OrganizationInfo
    for _, m := range o.Memberships {
        if m.IsActive() {
            orgs = append(orgs, OrganizationInfo{
                ID:   m.OrganizationID,
                Role: string(m.Role),
            })
        }
    }

    // Find selected org.
    var selectedOrg *OrganizationInfo
    if o.LastOrganizationID != "" {
        for _, org := range o.Memberships {
            if org.OrganizationID == o.LastOrganizationID && org.IsActive() {
                selectedOrg = &OrganizationInfo{
                    ID:   org.OrganizationID,
                    Role: string(org.Role),
                }
                break
            }
        }
    }

    return OperatorResponse{
        ID:            o.ID,
        Email:         o.Email,
        Name:          o.Name,
        NeedsOrganization: len(orgs) == 0,
        Organizations:       orgs,
        LastOrganizationID:   o.LastOrganizationID,
        SelectedOrganization: selectedOrg,
        EmailVerified:  o.EmailVerified,
        Thresholds:    &o.Thresholds,
        Client:        &o.ClientSettings,
        CreatedAt:      o.CreatedAt.UnixMilli(),
    }
}
