package scim

import "time"

// SCIMUser represents a SCIM-compliant user resource.
type SCIMUser struct {
	ID           string            `json:"id"`
	ExternalID   string            `json:"externalId,omitempty"`
	UserName     string            `json:"userName"`
	Name         SCIMName         `json:"name"`
	DisplayName  string            `json:"displayName,omitempty"`
	Emails       []SCIMEmail       `json:"emails,omitempty"`
	PhoneNumbers []SCIMPhoneNumber `json:"phoneNumbers,omitempty"`
	Active       bool              `json:"active"`
	Groups       []SCIMMember      `json:"groups,omitempty"`
	Meta         SCIMMeta          `json:"meta"`
}

// SCIMName represents the name components.
type SCIMName struct {
	Formatted     string `json:"formatted,omitempty"`
	FamilyName   string `json:"familyName,omitempty"`
	GivenName    string `json:"givenName,omitempty"`
	MiddleName   string `json:"middleName,omitempty"`
	HonorificPrefix string `json:"honorificPrefix,omitempty"`
	HonorificSuffix string `json:"honorificSuffix,omitempty"`
}

// SCIMEmail represents an email address.
type SCIMEmail struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// SCIMPhoneNumber represents a phone number.
type SCIMPhoneNumber struct {
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

// SCIMMember represents group membership.
type SCIMMember struct {
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
	Ref   string `json:"$ref,omitempty"`
}

// SCIMMeta contains metadata about the resource.
type SCIMMeta struct {
	ResourceType string    `json:"resourceType"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"lastModified"`
	Location    string    `json:"location"`
	Version     string    `json:"version,omitempty"`
}

// SCIMListResponse represents a paginated list response.
type SCIMListResponse struct {
	TotalResults int         `json:"totalResults"`
	StartIndex   int         `json:"startIndex"`
	ItemsPerPage int         `json:"itemsPerPage"`
	Schemas     []string    `json:"schemas"`
	Resources   []SCIMUser  `json:"Resources"`
}

// NewSCIMUserFromOperator creates a SCIM user from an operator.
func NewSCIMUserFromOperator(id, externalID, userName, displayName string) *SCIMUser {
	now := time.Now()
	return &SCIMUser{
		ID:         id,
		ExternalID: externalID,
		UserName:   userName,
		Name: SCIMName{
			Formatted: displayName,
		},
		DisplayName: displayName,
		Active:      true,
		Meta: SCIMMeta{
			ResourceType: "User",
			Created:      now,
			Modified:     now,
		},
	}
}
