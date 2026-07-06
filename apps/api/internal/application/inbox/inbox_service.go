package inbox

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/inbox"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/metrics"
	"github.com/gin-gonic/gin"
)

// FCMNotifier defines the interface for sending FCM notifications.
type FCMNotifier interface {
	SendSilentWake(ctx context.Context, wake fcm.SilentWake) error
}

// DeviceCreator defines the interface for creating devices from inbox entries.
type DeviceCreator interface {
	CreateFromInbox(ctx context.Context, entry *inbox.InboxEntry, commandSecret string) (*device.Device, error)
}

// DeviceLookup defines the interface for looking up devices.
type DeviceLookup interface {
	GetDeviceByIMEI(ctx context.Context, imei string) (*device.Device, error)
}

// Service handles inbox operations.
type Service struct {
	repo         inbox.Repository
	logRepo      inbox.RegistrationLogRepository
	deviceSvc    DeviceCreator
	deviceLookup DeviceLookup
	fcmNotifier  FCMNotifier
	txManager    transaction.TxManager
	logger       *slog.Logger
}

// NewService creates a new InboxService.
func NewService(
	repo inbox.Repository,
	logRepo inbox.RegistrationLogRepository,
	deviceSvc DeviceCreator,
	deviceLookup DeviceLookup,
	fcmNotifier FCMNotifier,
	logger *slog.Logger,
) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		repo:         repo,
		logRepo:      logRepo,
		deviceSvc:    deviceSvc,
		deviceLookup: deviceLookup,
		fcmNotifier:  fcmNotifier,
		logger:       logger,
	}
}

// WithTxManager sets the transaction manager for ACID transactions.
func (s *Service) WithTxManager(txManager transaction.TxManager) *Service {
	s.txManager = txManager
	return s
}

// GetInbox returns paginated inbox entries for a specific operator.
func (s *Service) GetInbox(ctx context.Context, operatorID, status string, page, limit int) (*InboxListResponse, error) {
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

	entries, total, err := s.repo.ListByOperator(ctx, operatorID, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list inbox entries: %w", err)
	}

	responses := make([]InboxEntryResponse, 0, len(entries))
	for _, e := range entries {
		responses = append(responses, InboxEntryResponse{
			ID:                e.ID,
			IMEI:              e.IMEI,
			DeviceName:        e.DeviceName,
			DeviceClass:       e.DeviceClass,
			Model:             e.Model,
			Manufacturer:      e.Manufacturer,
			OSVersion:         e.OSVersion,
			AppVersion:        e.AppVersion,
			FCMToken:          e.FCMToken,
			FirebaseInstallID: e.FirebaseInstallID,
			Status:            string(e.Status),
			CreatedAt:         e.CreatedAt,
			ApprovedAt:        e.ApprovedAt,
			RejectedAt:        e.RejectedAt,
			OperatorID:        e.OperatorID,
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
		DeviceName:        entry.DeviceName,
		DeviceClass:       entry.DeviceClass,
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
		Notes:             entry.Notes,
		OperatorID:        entry.OperatorID,
	}, nil
}

// AckInbox handles the inbox acknowledgement based on action type.
// Implements the 5-state model from SPEC:
// - action=acknowledge: Device acknowledges receipt (PENDING -> ACKNOWLEDGED)
// - action=approve: Operator approves registration (ACKNOWLEDGED -> APPROVING -> APPROVED)
// - action=reject: Operator rejects registration (PENDING/ACKNOWLEDGED -> REJECTED).
func (s *Service) AckInbox(ctx context.Context, imei string, action string, operatorID, notes string) (*AckResponse, error) {
	// Validate action
	if action == string(inbox.DeviceAckActionAcknowledge) {
		// Device acknowledges receipt
		return s.DeviceAcknowledge(ctx, imei)
	} else if action == string(inbox.OperatorActionApprove) {
		// Operator approves
		return s.ApproveDevice(ctx, imei, operatorID, notes)
	} else if action == string(inbox.OperatorActionReject) {
		// Operator rejects
		return s.RejectDevice(ctx, imei, operatorID, notes)
	} else if action == string(inbox.AckActionApprove) || action == string(inbox.AckActionReject) {
		// Legacy action names for backward compatibility
		if action == string(inbox.AckActionApprove) {
			return s.ApproveDevice(ctx, imei, operatorID, notes)
		}
		return s.RejectDevice(ctx, imei, operatorID, notes)
	}

	return nil, ErrInvalidAckAction
}

// DeviceAcknowledge handles device acknowledgement of a registration request.
// Transitions from PENDING -> ACKNOWLEDGED.
func (s *Service) DeviceAcknowledge(ctx context.Context, imei string) (*AckResponse, error) {
	entry, err := s.repo.GetByIMEI(ctx, imei)
	if err != nil {
		if err == inbox.ErrInboxNotFound {
			return nil, ErrInboxNotFound
		}
		return nil, fmt.Errorf("failed to get inbox entry: %w", err)
	}

	// Validate state transition
	if !entry.CanBeAcknowledged() {
		return nil, ErrInboxCannotBeAcknowledged
	}

	now := time.Now()
	entry.Status = inbox.StatusAcknowledged
	entry.AcknowledgedAt = PtrToInt64(now.UnixMilli())
	entry.UpdatedAt = now.UnixMilli()

	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to update inbox entry: %w", err)
	}

	// Log the acknowledgement
	s.logRegistrationAction(ctx, entry, "acknowledged", "", "")

	s.logger.Info("device acknowledged registration request",
		"imei", entry.IMEI,
	)

	return &AckResponse{
		ID:             entry.ID,
		IMEI:           entry.IMEI,
		Status:         string(entry.Status),
		AcknowledgedAt: entry.AcknowledgedAt,
	}, nil
}

// ApproveDevice handles operator approval of a device registration.
// Transitions from ACKNOWLEDGED -> APPROVING -> APPROVED.
// Generates commandSecret and sends FCM push to device.
func (s *Service) ApproveDevice(ctx context.Context, imei string, operatorID, notes string) (*AckResponse, error) {
	entry, err := s.repo.GetByIMEI(ctx, imei)
	if err != nil {
		if err == inbox.ErrInboxNotFound {
			return nil, ErrInboxNotFound
		}
		return nil, fmt.Errorf("failed to get inbox entry: %w", err)
	}

	// Validate state transition - must be acknowledged first
	if !entry.CanBeApproved() {
		return nil, ErrInboxCannotBeApproved
	}

	now := time.Now()
	entry.OperatorID = operatorID

	// Transition to APPROVING (intermediate state while processing)
	entry.Status = inbox.StatusApproving
	entry.ApprovingAt = PtrToInt64(now.UnixMilli())

	// Generate command secret
	secret, err := generateSecret(32)
	if err != nil {
		return nil, ErrSecretGeneration
	}
	entry.CommandSecret = secret

	// Create device in devices table
	if s.deviceSvc != nil {
		createdDevice, err := s.deviceSvc.CreateFromInbox(ctx, entry, secret)
		if err != nil {
			s.logger.Error("failed to create device from inbox entry",
				"imei", entry.IMEI,
				"error", err,
			)
			return nil, fmt.Errorf("failed to create device: %w", err)
		}
		s.logger.Info("device created from inbox entry",
			"imei", entry.IMEI,
			"device_id", createdDevice.ID,
		)
	}

	// Transition to APPROVED (final approved state)
	entry.Status = inbox.StatusApproved
	entry.ApprovedAt = PtrToInt64(now.UnixMilli())
	entry.UpdatedAt = now.UnixMilli()
	entry.Notes = notes

	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to update inbox entry: %w", err)
	}

	// Send FCM notification (best effort)
	fcmPushSent := false
	if s.fcmNotifier != nil && entry.FCMToken != "" {
		wake := fcm.SilentWake{
			Token:         entry.FCMToken,
			Command:       "REGISTRATION_APPROVED",
			CommandSecret: secret,
			DispatchID:    entry.ID,
			DeviceID:      entry.IMEI,
			Priority:      "high",
		}
		if err := s.fcmNotifier.SendSilentWake(ctx, wake); err != nil {
			s.logger.Warn("failed to send FCM notification on approval",
				"imei", entry.IMEI,
				"error", err,
			)
		} else {
			fcmPushSent = true
		}
	}

	// Log the approval
	s.logRegistrationAction(ctx, entry, "approved", operatorID, "")

	s.logger.Info("device approved by operator",
		"imei", entry.IMEI,
		"operator_id", operatorID,
		"fcm_push_sent", fcmPushSent,
	)

	return &AckResponse{
		ID:             entry.ID,
		IMEI:           entry.IMEI,
		Status:         string(entry.Status),
		ApprovedAt:     entry.ApprovedAt,
		CommandSecret:  secret,
		FCMPushSent:    fcmPushSent,
		Notes:          notes,
	}, nil
}

// RejectDevice handles operator rejection of a device registration.
// Transitions from PENDING/ACKNOWLEDGED -> REJECTED.
func (s *Service) RejectDevice(ctx context.Context, imei string, operatorID, notes string) (*AckResponse, error) {
	entry, err := s.repo.GetByIMEI(ctx, imei)
	if err != nil {
		if err == inbox.ErrInboxNotFound {
			return nil, ErrInboxNotFound
		}
		return nil, fmt.Errorf("failed to get inbox entry: %w", err)
	}

	// Validate state transition - can reject from pending or acknowledged
	if !entry.CanBeRejected() {
		return nil, ErrInboxCannotBeRejected
	}

	now := time.Now()
	entry.Status = inbox.StatusRejected
	entry.RejectedAt = PtrToInt64(now.UnixMilli())
	entry.UpdatedAt = now.UnixMilli()
	entry.OperatorID = operatorID
	entry.Notes = notes

	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to update inbox entry: %w", err)
	}

	// Log the rejection
	s.logRegistrationAction(ctx, entry, "rejected", operatorID, notes)

	s.logger.Info("device rejected by operator",
		"imei", entry.IMEI,
		"operator_id", operatorID,
	)

	return &AckResponse{
		ID:         entry.ID,
		IMEI:       entry.IMEI,
		Status:     string(entry.Status),
		RejectedAt: entry.RejectedAt,
		Notes:      notes,
	}, nil
}

// ackInboxWithTx executes ack within a database transaction for ACID guarantees.
func (s *Service) ackInboxWithTx(ctx context.Context, imei string, action string, operatorID, notes string) (*AckResponse, error) {
	var result *AckResponse
	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		entry, err := s.repo.GetByIMEI(txCtx, imei)
		if err != nil {
			if err == inbox.ErrInboxNotFound {
				return ErrInboxNotFound
			}
			return fmt.Errorf("failed to get inbox entry: %w", err)
		}

		if !entry.CanBeAcknowledged() {
			return ErrInboxNotPending
		}

		now := time.Now()
		entry.OperatorID = operatorID

		if action == string(inbox.AckActionApprove) {
			entry.Status = inbox.StatusApproved
			entry.ApprovedAt = PtrToInt64(now.UnixMilli())

			// Generate command secret
			secret, err := generateSecret(32)
			if err != nil {
				return ErrSecretGeneration
			}
			entry.CommandSecret = secret

			// Create device in devices table from inbox entry (within same transaction)
			if s.deviceSvc != nil {
				createdDevice, err := s.deviceSvc.CreateFromInbox(txCtx, entry, secret)
				if err != nil {
					s.logger.Error("failed to create device from inbox entry",
						"imei", entry.IMEI,
						"error", err,
					)
					return fmt.Errorf("failed to create device: %w", err)
				}
				s.logger.Info("device created from inbox entry",
					"imei", entry.IMEI,
					"device_id", createdDevice.ID,
				)
			}

			// Update inbox entry in database (within same transaction)
			if err := s.repo.Update(txCtx, entry); err != nil {
				return fmt.Errorf("failed to update inbox entry: %w", err)
			}

			// Send FCM notification (outside transaction - best effort)
			fcmPushSent := false
			if s.fcmNotifier != nil && entry.FCMToken != "" {
				wake := fcm.SilentWake{
					Token:         entry.FCMToken,
					Command:       "REGISTRATION_APPROVED",
					CommandSecret: secret,
					DispatchID:    entry.ID,
					DeviceID:      entry.IMEI,
					Priority:      "high",
				}
				if err := s.fcmNotifier.SendSilentWake(ctx, wake); err != nil {
					s.logger.Warn("failed to send FCM notification on approval",
						"imei", entry.IMEI,
						"error", err,
					)
				} else {
					fcmPushSent = true
				}
			}

			// Log the approval
			s.logRegistrationAction(txCtx, entry, "approved", operatorID, "")

			result = &AckResponse{
				ID:            entry.ID,
				IMEI:          entry.IMEI,
				Status:        string(entry.Status),
				ApprovedAt:    entry.ApprovedAt,
				CommandSecret: secret,
				FCMPushSent:   fcmPushSent,
				Notes:         notes,
			}
			return nil
		}

		// Reject path
		entry.Status = inbox.StatusRejected
		entry.RejectedAt = PtrToInt64(now.UnixMilli())
		entry.Notes = notes

		if err := s.repo.Update(txCtx, entry); err != nil {
			return fmt.Errorf("failed to update inbox entry: %w", err)
		}

		// Log the rejection
		s.logRegistrationAction(txCtx, entry, "rejected", operatorID, notes)

		result = &AckResponse{
			ID:         entry.ID,
			IMEI:       entry.IMEI,
			Status:     string(entry.Status),
			RejectedAt: entry.RejectedAt,
			Notes:      notes,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// ackInboxWithoutTx executes ack without transaction (legacy behavior).
func (s *Service) ackInboxWithoutTx(ctx context.Context, imei string, action string, operatorID, notes string) (*AckResponse, error) {
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

		// Create device in devices table from inbox entry
		if s.deviceSvc != nil {
			createdDevice, err := s.deviceSvc.CreateFromInbox(ctx, entry, secret)
			if err != nil {
				s.logger.Error("failed to create device from inbox entry",
					"imei", entry.IMEI,
					"error", err,
				)
				return nil, fmt.Errorf("failed to create device: %w", err)
			}
			s.logger.Info("device created from inbox entry",
				"imei", entry.IMEI,
				"device_id", createdDevice.ID,
			)
		}

		// Update inbox entry in database
		if err := s.repo.Update(ctx, entry); err != nil {
			return nil, fmt.Errorf("failed to update inbox entry: %w", err)
		}

		// Send FCM notification (best effort)
		fcmPushSent := false
		if s.fcmNotifier != nil && entry.FCMToken != "" {
			wake := fcm.SilentWake{
				Token:         entry.FCMToken,
				Command:       "REGISTRATION_APPROVED",
				CommandSecret: secret,
				DispatchID:    entry.ID,
				DeviceID:      entry.IMEI,
				Priority:      "high",
			}
			if err := s.fcmNotifier.SendSilentWake(ctx, wake); err != nil {
				s.logger.Warn("failed to send FCM notification on approval",
					"imei", entry.IMEI,
					"error", err,
				)
			} else {
				fcmPushSent = true
				s.logger.Info("FCM notification sent on approval",
					"imei", entry.IMEI,
				)
			}
		}

		// Log the approval
		s.logRegistrationAction(ctx, entry, "approved", operatorID, "")

		return &AckResponse{
			ID:            entry.ID,
			IMEI:          entry.IMEI,
			Status:        string(entry.Status),
			ApprovedAt:    entry.ApprovedAt,
			CommandSecret: secret,
			FCMPushSent:   fcmPushSent,
			Notes:         notes,
		}, nil
	}

	// Reject action
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
}

// CreateInboxRequest creates a new inbox entry from a device registration request.
func (s *Service) CreateInboxRequest(ctx context.Context, req *InboxRequest) (*InboxEntryResponse, error) {
	if err := s.validateInboxRequest(req); err != nil {
		metrics.Get().RecordDeviceRegistrationFailure()
		return nil, err
	}

	existingInboxEntry, err := s.checkDeviceAndInboxStatus(ctx, req.IMEI)
	if err != nil {
		return nil, err
	}

	entry, err := s.createInboxEntry(ctx, req)
	if err != nil {
		return nil, err
	}

	s.logAndRecordMetrics(ctx, req.IMEI, entry, existingInboxEntry)

	return s.buildInboxEntryResponse(entry), nil
}

// validateInboxRequest validates the inbox request input.
func (s *Service) validateInboxRequest(req *InboxRequest) error {
	if !isValidIMEI(req.IMEI) {
		return ErrInvalidIMEI
	}
	if req.FCMToken != "" && !isValidFCMToken(req.FCMToken) {
		return ErrInvalidFCMToken
	}
	if req.FirebaseInstallID != "" && !isValidFirebaseInstallID(req.FirebaseInstallID) {
		return ErrInvalidFirebaseInstallID
	}
	return nil
}

// checkDeviceAndInboxStatus checks device and inbox status for the request.
func (s *Service) checkDeviceAndInboxStatus(ctx context.Context, imei string) (*inbox.InboxEntry, error) {
	if s.deviceLookup != nil {
		if err := s.checkDeviceStatus(ctx, imei); err != nil {
			return nil, err
		}
	}

	return s.cleanupStaleInboxEntry(ctx, imei)
}

// checkDeviceStatus checks if device exists and is not already registered.
func (s *Service) checkDeviceStatus(ctx context.Context, imei string) error {
	existingDevice, err := s.deviceLookup.GetDeviceByIMEI(ctx, imei)
	if err != nil && err != device.ErrNotFound {
		return fmt.Errorf("failed to check existing device: %w", err)
	}
	if existingDevice != nil && !existingDevice.IsDeregistered() {
		return ErrDeviceAlreadyExists
	}
	if existingDevice != nil {
		s.logger.Info("device re-registration allowed for deregistered device",
			"imei", imei,
			"deregistered_at", existingDevice.DeregisteredAt,
		)
	}
	return nil
}

// cleanupStaleInboxEntry cleans up stale inbox entries.
func (s *Service) cleanupStaleInboxEntry(ctx context.Context, imei string) (*inbox.InboxEntry, error) {
	existingInboxEntry, err := s.repo.GetByIMEI(ctx, imei)
	if err != nil && err != inbox.ErrInboxNotFound {
		return nil, fmt.Errorf("failed to check existing inbox entry: %w", err)
	}
	if existingInboxEntry != nil {
		if existingInboxEntry.IsPending() {
			return nil, ErrAlreadyExists
		}
		s.logger.Info("cleaning up stale inbox entry for re-registration",
			"imei", imei,
			"old_status", existingInboxEntry.Status,
		)
		if err := s.repo.DeleteByIMEI(ctx, imei); err != nil {
			return nil, fmt.Errorf("failed to clean up stale inbox entry: %w", err)
		}
	}
	return existingInboxEntry, nil
}

// createInboxEntry creates the inbox entry in the repository.
func (s *Service) createInboxEntry(ctx context.Context, req *InboxRequest) (*inbox.InboxEntry, error) {
	now := time.Now()
	entry := &inbox.InboxEntry{
		ID:                generateID(),
		IMEI:              req.IMEI,
		DeviceName:        req.DeviceName,
		DeviceClass:       req.DeviceClass,
		Model:             req.Model,
		Manufacturer:      req.Manufacturer,
		OSVersion:         req.OSVersion,
		AppVersion:        req.AppVersion,
		FCMToken:          req.FCMToken,
		FirebaseInstallID: req.FirebaseInstallID,
		Status:            inbox.StatusPending,
		CreatedAt:         now.UnixMilli(),
		UpdatedAt:         now.UnixMilli(),
	}

	if err := s.repo.Create(ctx, entry); err != nil {
		metrics.Get().RecordDeviceRegistrationFailure()
		return nil, fmt.Errorf("failed to create inbox entry: %w", err)
	}
	return entry, nil
}

// logAndRecordMetrics logs and records metrics for the registration.
func (s *Service) logAndRecordMetrics(ctx context.Context, imei string, entry *inbox.InboxEntry, existingInboxEntry *inbox.InboxEntry) {
	s.logRegistrationAction(ctx, entry, "pending", "", "")
	metrics.Get().RecordDeviceRegistrationSuccess()

	if existingInboxEntry != nil || s.deviceLookup != nil {
		if d, err := s.deviceLookup.GetDeviceByIMEI(ctx, imei); err == nil && d != nil && d.IsDeregistered() {
			metrics.Get().RecordDeviceReRegistration()
		}
	}
}

// buildInboxEntryResponse builds the response from the entry.
func (s *Service) buildInboxEntryResponse(entry *inbox.InboxEntry) *InboxEntryResponse {
	return &InboxEntryResponse{
		ID:                entry.ID,
		IMEI:              entry.IMEI,
		DeviceName:        entry.DeviceName,
		DeviceClass:       entry.DeviceClass,
		Model:             entry.Model,
		Manufacturer:      entry.Manufacturer,
		OSVersion:         entry.OSVersion,
		AppVersion:        entry.AppVersion,
		FCMToken:          entry.FCMToken,
		FirebaseInstallID: entry.FirebaseInstallID,
		Status:            string(entry.Status),
		AcknowledgedAt:    entry.AcknowledgedAt,
		ApprovingAt:      entry.ApprovingAt,
		ApprovedAt:        entry.ApprovedAt,
		RejectedAt:        entry.RejectedAt,
		Notes:             entry.Notes,
		OperatorID:        entry.OperatorID,
		CreatedAt:         entry.CreatedAt,
	}
}

// logRegistrationAction logs a registration action to the audit log.
func (s *Service) logRegistrationAction(ctx context.Context, entry *inbox.InboxEntry, action, operatorID, details string) {
	if s.logRepo == nil {
		return
	}

	log := &inbox.RegistrationLog{
		ID:         generateID(),
		DeviceID:   "",
		IMEI:       entry.IMEI,
		Action:     action,
		OperatorID: operatorID,
		Details:    details,
		Timestamp:  time.Now().UnixMilli(),
		ClientIP:   extractClientIP(ctx),
		UserAgent:  extractUserAgent(ctx),
	}

	_ = s.logRepo.Create(ctx, log)
}

// extractClientIP extracts the client IP from the context.
// Looks for X-Forwarded-For, X-Real-IP headers or the request context.
func extractClientIP(ctx context.Context) string {
	// Try to get from Gin context if available
	if ginCtx, ok := ctx.(*gin.Context); ok {
		// Check X-Forwarded-For header (proxy)
		if xff := ginCtx.GetHeader("X-Forwarded-For"); xff != "" {
			// Take the first IP (original client)
			if idx := strings.Index(xff, ","); idx != -1 {
				return strings.TrimSpace(xff[:idx])
			}
			return strings.TrimSpace(xff)
		}
		// Check X-Real-IP header
		if xri := ginCtx.GetHeader("X-Real-IP"); xri != "" {
			return xri
		}
		// Fall back to remote address
		return ginCtx.ClientIP()
	}
	return ""
}

// extractUserAgent extracts the User-Agent from the context.
func extractUserAgent(ctx context.Context) string {
	// Try to get from Gin context if available
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx.GetHeader("User-Agent")
	}
	return ""
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

// isValidIMEI validates IMEI format (15 digits) using Luhn checksum.
// The IMEI is a 15-digit number where the last digit is a Luhn check digit.
func isValidIMEI(imei string) bool {
	if len(imei) != 15 {
		return false
	}
	// Check all characters are digits
	for _, c := range imei {
		if c < '0' || c > '9' {
			return false
		}
	}
	// Validate using Luhn algorithm
	return isValidLuhn(imei)
}

// isValidLuhn validates a number using the Luhn algorithm.
// The Luhn algorithm is used to validate IMEI, credit card numbers, etc.
func isValidLuhn(number string) bool {
	sum := 0
	isSecond := false

	// Process digits from right to left
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if isSecond {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}

// isValidFCMToken validates basic FCM token format.
// FCM tokens are typically 152+ characters and alphanumeric with some special chars.
func isValidFCMToken(token string) bool {
	// FCM tokens are typically at least 100 characters ( from Firebase docs)
	if len(token) < 100 {
		return false
	}
	// FCM tokens should not contain spaces or control characters
	for _, c := range token {
		if c < 32 || c > 126 {
			return false
		}
	}
	// FCM tokens are generally alphanumeric with : and _ allowed
	// Pattern: alphanumeric, colon, underscore, hyphen
	for _, c := range token {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') &&
			c != ':' && c != '_' && c != '-' && c != '.' {
			return false
		}
	}
	return true
}

// isValidFirebaseInstallID validates Firebase Installation ID format.
// Firebase Installation IDs are 22 characters alphanumeric.
func isValidFirebaseInstallID(id string) bool {
	// Firebase Installation IDs are 22 characters
	if len(id) < 10 || len(id) > 50 {
		return false
	}
	// Must be alphanumeric with specific allowed characters
	for _, c := range id {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') &&
			c != '-' && c != '_' {
			return false
		}
	}
	return true
}

// UpdateInboxEntry updates an inbox entry (e.g., operator notes).
// Only pending entries can be updated.
func (s *Service) UpdateInboxEntry(ctx context.Context, imei, operatorID, notes string) (*InboxEntryResponse, error) {
	entry, err := s.repo.GetByIMEI(ctx, imei)
	if err != nil {
		if err == inbox.ErrInboxNotFound {
			return nil, ErrInboxNotFound
		}
		return nil, fmt.Errorf("failed to get inbox entry: %w", err)
	}

	// Only pending entries can be updated
	if !entry.CanBeAcknowledged() {
		return nil, ErrInboxNotPending
	}

	// Update notes
	entry.Notes = notes
	entry.UpdatedAt = time.Now().UnixMilli()

	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to update inbox entry: %w", err)
	}

	// Log the update action
	s.logRegistrationAction(ctx, entry, "updated", operatorID, notes)

	return &InboxEntryResponse{
		ID:                entry.ID,
		IMEI:              entry.IMEI,
		DeviceName:        entry.DeviceName,
		DeviceClass:       entry.DeviceClass,
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
		Notes:             entry.Notes,
		OperatorID:        entry.OperatorID,
	}, nil
}

// ResendApproval resends the FCM notification to a device that was approved.
// This handles the race condition where the original FCM notification may have failed.
func (s *Service) ResendApproval(ctx context.Context, imei, operatorID string) (*ResendResponse, error) {
	entry, err := s.repo.GetByIMEI(ctx, imei)
	if err != nil {
		if err == inbox.ErrInboxNotFound {
			return nil, ErrInboxNotFound
		}
		return nil, fmt.Errorf("failed to get inbox entry: %w", err)
	}

	// Only approved entries can have their notification resent
	if entry.Status != inbox.StatusApproved {
		return nil, ErrInboxNotApproved
	}

	// Validate we have a command secret to resend
	if entry.CommandSecret == "" {
		return nil, fmt.Errorf("command secret not found for approved entry")
	}

	// Try to send FCM notification
	fcmPushSent := false
	if s.fcmNotifier != nil && entry.FCMToken != "" {
		wake := fcm.SilentWake{
			Token:         entry.FCMToken,
			Command:       "REGISTRATION_APPROVED",
			CommandSecret: entry.CommandSecret, // Plaintext secret stored in inbox entry
			DispatchID:    entry.ID,
			DeviceID:      entry.IMEI,
			Priority:      "high",
		}
		if err := s.fcmNotifier.SendSilentWake(ctx, wake); err != nil {
			s.logger.Warn("failed to resend FCM notification",
				"imei", entry.IMEI,
				"error", err,
			)
		} else {
			fcmPushSent = true
		}
	}

	// Log the resend action
	s.logRegistrationAction(ctx, entry, "resend_approval", operatorID, "")

	s.logger.Info("resent approval notification",
		"imei", entry.IMEI,
		"fcm_push_sent", fcmPushSent,
	)

	return &ResendResponse{
		IMEI:        entry.IMEI,
		FCMPushSent: fcmPushSent,
		ResentAt:    time.Now().UnixMilli(),
	}, nil
}

// ResendResponse represents the response for resending approval notification.
type ResendResponse struct {
	IMEI        string `json:"imei"`
	FCMPushSent bool   `json:"fcmPushSent"`
	ResentAt    int64  `json:"resentAt"`
}
