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
    Role          OperatorRole    `json:"role"`
    CreatedAt     int64           `json:"createdAt"`
    EmailVerified bool            `json:"emailVerified,omitempty"`
}

// AuthErrorResponse is the standard error envelope for auth endpoints.
type AuthErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message,omitempty"`
}

// ToResponse converts an Operator to its safe JSON representation.
func (o *Operator) ToResponse() OperatorResponse {
    return OperatorResponse{
        ID:            o.ID,
        Email:         o.Email,
        Name:          o.Name,
        Role:          o.Role,
        EmailVerified: o.EmailVerified,
        Thresholds:    &o.Thresholds,
        Client:        &o.ClientSettings,
        CreatedAt:     o.CreatedAt.UnixMilli(),
    }
}
