package client

import (
	"context"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/client"
)

// Service handles client operations.
type Service struct {
	clientRepo client.Repository
}

// NewService creates a new ClientService.
func NewService(clientRepo client.Repository) *Service {
	return &Service{
		clientRepo: clientRepo,
	}
}

// CreateClientRequest is the request to create a new client.
type CreateClientRequest struct {
	OperatorID     string
	Name           string
	Platform       string
	AllowedOrigins []string
	AllowedPaths   []string
	RateLimit      int
}

// Create creates a new API client.
func (s *Service) Create(ctx context.Context, req *CreateClientRequest) (*dto.ClientResponse, string, error) {
	c := &client.Client{
		OperatorID:     req.OperatorID,
		Name:           req.Name,
		Platform:       req.Platform,
		AllowedOrigins: req.AllowedOrigins,
		AllowedPaths:   req.AllowedPaths,
		RateLimit:      req.RateLimit,
		IsActive:       true,
	}

	created, secret, err := s.clientRepo.Create(ctx, c, "")
	if err != nil {
		return nil, "", err
	}

	return &dto.ClientResponse{
		ID:             created.ID,
		OperatorID:     created.OperatorID,
		Name:           created.Name,
		Platform:       created.Platform,
		AllowedOrigins: created.AllowedOrigins,
		AllowedPaths:   created.AllowedPaths,
		RateLimit:      created.RateLimit,
		IsActive:       created.IsActive,
	}, secret, nil
}

// ListAll lists all clients (admin use).
func (s *Service) ListAll(ctx context.Context, limit, offset int) ([]dto.ClientResponse, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	clients, total, err := s.clientRepo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	response := make([]dto.ClientResponse, len(clients))
	for i, c := range clients {
		response[i] = toClientResponse(c)
	}

	return response, total, nil
}

// ListByOperatorID lists all clients for a specific operator.
func (s *Service) ListByOperatorID(ctx context.Context, operatorID string) ([]dto.ClientResponse, error) {
	clients, err := s.clientRepo.FindByOperatorID(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.ClientResponse, len(clients))
	for i, c := range clients {
		response[i] = toClientResponse(c)
	}

	return response, nil
}

// Get retrieves a client by ID.
func (s *Service) Get(ctx context.Context, clientID string) (*dto.ClientResponse, error) {
	c, err := s.clientRepo.FindByID(ctx, clientID)
	if err != nil {
		if err == client.ErrNotFound {
			return nil, application.ErrClientNotFound
		}
		return nil, err
	}

	resp := toClientResponse(c)
	return &resp, nil
}

// Update updates a client.
func (s *Service) Update(ctx context.Context, clientID string, req *dto.UpdateClientRequest) (*dto.ClientResponse, error) {
	c, err := s.clientRepo.FindByID(ctx, clientID)
	if err != nil {
		if err == client.ErrNotFound {
			return nil, application.ErrClientNotFound
		}
		return nil, err
	}

	// Update fields.
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.AllowedOrigins != nil {
		c.AllowedOrigins = req.AllowedOrigins
	}
	if req.AllowedPaths != nil {
		c.AllowedPaths = req.AllowedPaths
	}
	if req.RateLimit != nil {
		c.RateLimit = *req.RateLimit
	}
	if req.IsActive != nil {
		c.IsActive = *req.IsActive
	}

	if err := s.clientRepo.Update(ctx, c); err != nil {
		return nil, err
	}

	resp := toClientResponse(c)
	return &resp, nil
}

// Delete deletes a client.
func (s *Service) Delete(ctx context.Context, clientID string) error {
	return s.clientRepo.Delete(ctx, clientID)
}

// RotateKey rotates the signing key for a client with 24-hour grace period.
// Returns the new key version.
func (s *Service) RotateKey(ctx context.Context, clientID string) (int, error) {
	signingKey, _, err := s.clientRepo.RotateSigningKey(ctx, clientID, 24*time.Hour)
	if err != nil {
		return 0, err
	}
	return signingKey.Version, nil
}

// Deactivate deactivates a client (revokes credentials).
func (s *Service) Deactivate(ctx context.Context, clientID string) error {
	c, err := s.clientRepo.FindByID(ctx, clientID)
	if err != nil {
		return err
	}
	c.IsActive = false
	return s.clientRepo.Update(ctx, c)
}

// GetByOperatorID retrieves a client by ID and verifies it belongs to the operator.
func (s *Service) GetByOperatorID(ctx context.Context, clientID, operatorID string) (*dto.ClientResponse, error) {
	c, err := s.clientRepo.FindByID(ctx, clientID)
	if err != nil {
		if err == client.ErrNotFound {
			return nil, application.ErrClientNotFound
		}
		return nil, err
	}

	// Verify ownership
	if c.OperatorID != operatorID {
		return nil, application.ErrClientNotFound
	}

	resp := toClientResponse(c)
	return &resp, nil
}

// GetHmacKey retrieves the HMAC signing key for a client.
// Returns the key and ok=false if client not found or inactive.
func (s *Service) GetHmacKey(ctx context.Context, clientID string) (string, bool) {
	return s.clientRepo.GetHmacKey(ctx, clientID)
}

// toClientResponse converts a client to a response DTO.
func toClientResponse(c *client.Client) dto.ClientResponse {
	return dto.ClientResponse{
		ID:             c.ID,
		OperatorID:     c.OperatorID,
		Name:           c.Name,
		Platform:       c.Platform,
		AllowedOrigins: c.AllowedOrigins,
		AllowedPaths:   c.AllowedPaths,
		RateLimit:      c.RateLimit,
		IsActive:       c.IsActive,
		RequestCount:   c.RequestCount,
		LastRequestAt:  c.LastRequestAt,
		CreatedAt:      c.CreatedAt.UnixMilli(),
		UpdatedAt:      c.UpdatedAt.UnixMilli(),
	}
}
