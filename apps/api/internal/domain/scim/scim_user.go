package scim

import "time"

// SCIMUser represents a SCIM 2.0 user resource.
type SCIMUser struct {
	ID           string         `json:"id"`
	ExternalID   string         `json:"externalId,omitempty"`
	UserName     string         `json:"userName"`
	Name         SCIMName       `json:"name"`
	DisplayName  string         `json:"displayName"`
	Emails       []SCIMEmail    `json:"emails"`
	PhoneNumbers []SCIMPhone     `json:"phoneNumbers,omitempty"`
	Active       bool           `json:"active"`
	Groups       []SCIMGroupRef `json:"groups,omitempty"`
	Roles        []SCIMRole     `json:"roles,omitempty"`
	Meta         SCIMMeta       `json:"meta"`
}

// SCIMName represents the name components of a SCIM user.
type SCIMName struct {
	Formatted     string `json:"formatted"`
	FamilyName    string `json:"familyName"`
	GivenName     string `json:"givenName"`
	MiddleName    string `json:"middleName,omitempty"`
	HonorificPrefix string `json:"honorificPrefix,omitempty"`
	HonorificSuffix string `json:"honorificSuffix,omitempty"`
}

// SCIMEmail represents an email address.
type SCIMEmail struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// SCIMPhone represents a phone number.
type SCIMPhone struct {
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

// SCIMGroupRef represents a group membership.
type SCIMGroupRef struct {
	Value   string `json:"value"`
	Ref     string `json:"$ref,omitempty"`
	Type    string `json:"type,omitempty"`
	Display string `json:"display,omitempty"`
}

// SCIMRole represents a role assignment.
type SCIMRole struct {
	Value   string `json:"value"`
	Display string `json:"display,omitempty"`
}

// SCIMMeta represents metadata for a SCIM resource.
type SCIMMeta struct {
	ResourceType string    `json:"resourceType"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"lastModified"`
	Location     string    `json:"location"`
	Version      string    `json:"version,omitempty"`
}

// SCIMListResponse represents a paginated list response.
type SCIMListResponse struct {
	TotalResults int         `json:"totalResults"`
	StartIndex   int         `json:"startIndex"`
	ItemsPerPage int         `json:"itemsPerPage"`
	Resources    []SCIMUser  `json:"schemas,omitempty"`
}

// SCIMError represents a SCIM error response.
type SCIMError struct {
	SCIMType string `json:"scimType,omitempty"`
	Detail   string `json:"detail"`
	Status   int    `json:"status"`
}
