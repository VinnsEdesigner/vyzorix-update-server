package inbox

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/inbox"
)

// Service handles inbox operations.
type Service struct {
	repo      inbox.Repository
	logRepo   inbox.RegistrationLogRepository
	deviceSvc interface {
		CreateFromInbox(ctx context.Context, entry *inbox.InboxEntry, commandSecret string) error
	}
}

// NewService creates a new InboxService.
func NewService(
	repo inbox.Repository,
	logRepo inbox.RegistrationLogRepository,
	deviceSvc interface {
		CreateFromInbox(ctx context.Context, entry *inbox.InboxEntry, commandSecret string) error
	},
) *Service {
	return &Service{
		repo:      repo,
		logRepo:   logRepo,
		deviceSvc: deviceSvc,
	}
}

// GetInbox returns paginated inbox entries.
func (s *Service) GetInbox(ctx context.Context, status string, page, limit int) (*InboxListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	entries, total, err := s.repo.List(ctx, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list inbox entries: %w", err)
	}

	responses := make([]InboxEntryResponse, 0, len(entries))
	for _, e := range entries {
		responses = append(responses, InboxEntryResponse{
			ID:                e.ID,
			IMEI:              e.IMEI,
			Model:             e.Model,
			Manufacturer:      e.Manufacturer,
			OSVersion:         e.OSVersion,
			AppVersion:        e.AppVersion,
			FCMToken:         e.FCMToken,
			FirebaseInstallID: e.FirebaseInstallID,
			Status:            string(e.Status),
			CreatedAt:         e.CreatedAt,
			ApprovedAt:        e.ApprovedAt,
			RejectedAt:        e.RejectedAt,
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return &InboxListResponse{
		Requests: responses,
		Pagination: PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetInboxEntry returns a single inbox entry by IMEI.
func (s *Service) GetInboxEntry(ctx context.Context, imei string) (*InboxEntryResponse, error) {
	entry, err := s.repo.GetByIMEI(ctx, imei)
	if err != nil {
		if err == inbox.ErrInboxNotFound {
			return nil, ErrInboxNotFound
		}
		return nil, fmt.Errorf("failed to get inbox entry: %w", err)
	}

	return &InboxEntryResponse{
		ID:                entry.ID,
		IMEI:              entry.IMEI,
		Model:             entry.Model,
		Manufacturer:      entry.Manufacturer,
		OSVersion:         entry.OSVersion,
		AppVersion:        entry.AppVersion,
		FCMToken:          entry.FCMToken,
		FirebaseInstallID: entry.FirebaseInstallID,
		Status:            string(entry.Status),
		CreatedAt:         entry.CreatedAt,
		ApprovedAt:        entry.ApprovedAt,
		RejectedAt:        entry.RejectedAt,
		CommandSecret:      entry.CommandSecret,
	}, nil
}

// AckInbox acknowledges (approves or rejects) an inbox entry.
func (s *Service) AckInbox(ctx context.Context, imei string, action string, operatorID, notes string) (*AckResponse, error) {
	entry, err := s.repo.GetByIMEI(ctx, imei)
	if err != nil {
		if err == inbox.ErrInboxNotFound {
			return nil, ErrInboxNotFound
		}
		return nil, fmt.Errorf("failed to get inbox entry: %w", err)
	}

	if !entry.CanBeAcknowledged() {
		return nil, ErrInboxNotPending
	}

	now := time.Now()
	entry.OperatorID = operatorID

	if action == string(inbox.AckActionApprove) {
		entry.Status = inbox.StatusApproved
		entry.ApprovedAt = PtrToInt64(now.UnixMilli())

		// Generate command secret
		secret, err := generateSecret(32)
		if err != nil {
			return nil, ErrSecretGeneration
		}
		entry.CommandSecret = secret

		// Update in database
		if err := s.repo.Update(ctx, entry); err != nil {
			return nil, fmt.Errorf("failed to update inbox entry: %w", err)
		}

		// Log the approval
		s.logRegistrationAction(ctx, entry, "approved", operatorID, "")

		return &AckResponse{
			ID:            entry.ID,
			IMEI:          entry.IMEI,
			Status:        string(entry.Status),
			ApprovedAt:    entry.ApprovedAt,
			CommandSecret: secret,
			FCMPushSent:   false, // FCM push handled by handler
			Notes:         notes,
		}, nil

	} else if action == string(inbox.AckActionReject) {
		entry.Status = inbox.StatusRejected
		entry.RejectedAt = PtrToInt64(now.UnixMilli())
		entry.Notes = notes

		if err := s.repo.Update(ctx, entry); err != nil {
			return nil, fmt.Errorf("failed to update inbox entry: %w", err)
		}

		// Log the rejection
		s.logRegistrationAction(ctx, entry, "rejected", operatorID, notes)

		return &AckResponse{
			ID:         entry.ID,
			IMEI:       entry.IMEI,
			Status:     string(entry.Status),
			RejectedAt: entry.RejectedAt,
			Notes:      notes,
		}, nil

	} else {
		return nil, ErrInvalidAckAction
	}
}

// CreateInboxRequest creates a new inbox entry from a device registration request.
func (s *Service) CreateInboxRequest(ctx context.Context, req *InboxRequest) (*InboxEntryResponse, error) {
	// Check if already exists
	exists, err := s.repo.ExistsByIMEI(ctx, req.IMEI)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing entry: %w", err)
	}
	if exists {
		return nil, ErrAlreadyExists
	}

	now := time.Now()
	entry := &inbox.InboxEntry{
		ID:                 generateID(),
		IMEI:               req.IMEI,
		Model:              req.Model,
		Manufacturer:       req.Manufacturer,
		OSVersion:          req.OSVersion,
		AppVersion:         req.AppVersion,
		FCMToken:           req.FCMToken,
		FirebaseInstallID:  req.FirebaseInstallID,
		Status:             inbox.StatusPending,
		CreatedAt:          now.UnixMilli(),
	}

	if err := s.repo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create inbox entry: %w", err)
	}

	// Log the registration request
	s.logRegistrationAction(ctx, entry, "pending", "", "")

	return &InboxEntryResponse{
		ID:                entry.ID,
		IMEI:              entry.IMEI,
		Model:             entry.Model,
		Manufacturer:      entry.Manufacturer,
		OSVersion:         entry.OSVersion,
		AppVersion:        entry.AppVersion,
		FCMToken:          entry.FCMToken,
		FirebaseInstallID: entry.FirebaseInstallID,
		Status:            string(entry.Status),
		CreatedAt:         entry.CreatedAt,
	}, nil
}

// logRegistrationAction logs a registration action to the audit log.
func (s *Service) logRegistrationAction(ctx context.Context, entry *inbox.InboxEntry, action, operatorID, details string) {
	if s.logRepo == nil {
		return
	}

	log := &inbox.RegistrationLog{
		ID:          generateID(),
		DeviceID:    "",
		IMEI:        entry.IMEI,
		Action:      action,
		OperatorID:  operatorID,
		Details:     details,
		Timestamp:   time.Now().UnixMilli(),
	}

	_ = s.logRepo.Create(ctx, log)
}

// generateSecret generates a cryptographically secure random hex string.
func generateSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateID generates a unique ID.
func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "inb_" + hex.EncodeToString(bytes)
}
