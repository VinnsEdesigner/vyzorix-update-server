package inbox

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	deviceapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/inbox"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/password"
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

// GetInbox returns paginated inbox entries for a specific operator within an organization.
func (s *Service) GetInbox(ctx context.Context, operatorID, orgID, status string, page, limit int) (*InboxListResponse, error) {
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

	entries, total, err := s.repo.ListByOperator(ctx, operatorID, orgID, status, limit, offset)
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

// GetInboxEntry returns a single inbox entry by IMEI within an organization.
func (s *Service) GetInboxEntry(ctx context.Context, imei, orgID string) (*InboxEntryResponse, error) {
	entry, err := s.repo.GetByIMEI(ctx, imei)
	if err != nil {
		if err == inbox.ErrInboxNotFound {
			return nil, ErrInboxNotFound
		}
		return nil, fmt.Errorf("failed to get inbox entry: %w", err)
	}

	if orgID != "" && entry.OrganizationID != "" && entry.OrganizationID != orgID {
		return nil, ErrInboxNotFound
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
func (s *Service) AckInbox(ctx context.Context, imei, action, operatorID, orgID, notes string) (*AckResponse, error) {
	if action == string(inbox.DeviceAckActionAcknowledge) {
		return s.DeviceAcknowledge(ctx, imei)
	} else if action == string(inbox.OperatorActionApprove) {
		return s.ApproveDevice(ctx, imei, operatorID, orgID, notes)
	} else if action == string(inbox.OperatorActionReject) {
		return s.RejectDevice(ctx, imei, operatorID, orgID, notes)
	} else if action == string(inbox.AckActionApprove) {
		return s.ApproveDevice(ctx, imei, operatorID, orgID, notes)
	} else if action == string(inbox.AckActionReject) {
		return s.RejectDevice(ctx, imei, operatorID, orgID, notes)
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

	s.logRegistrationAction(ctx, entry, "acknowledged", "", "")

	return &AckResponse{
		ID:             entry.ID,
		IMEI:           entry.IMEI,
		Status:         string(entry.Status),
		AcknowledgedAt: entry.AcknowledgedAt,
	}, nil
}

// ApproveDevice handles operator approval of a device registration.
// Transitions from ACKNOWLEDGED -> APPROVING -> APPROVED.
func (s *Service) ApproveDevice(ctx context.Context, imei string, operatorID, orgID, notes string) (*AckResponse, error) {
	entry, err := s.repo.GetByIMEI(ctx, imei)
	if err != nil {
		if err == inbox.ErrInboxNotFound {
			return nil, ErrInboxNotFound
		}
		return nil, fmt.Errorf("failed to get inbox entry: %w", err)
	}

	if orgID != "" && entry.OrganizationID != "" && entry.OrganizationID != orgID {
		return nil, ErrInboxNotFound
	}

	if !entry.CanBeApproved() {
		return nil, ErrInboxCannotBeApproved
	}

	now := time.Now()
	entry.OperatorID = operatorID

	// Claim the entry for the operator's org on first action (public inbox.
	// entries have no org until an operator approves/rejects them).
	if entry.OrganizationID == "" && orgID != "" {
		entry.OrganizationID = orgID
	}

	// Transition to APPROVING (intermediate state).
	entry.Status = inbox.StatusApproving
	entry.ApprovingAt = PtrToInt64(now.UnixMilli())

	// Generate command secret.
	secret, err := generateSecret(32)
	if err != nil {
		return nil, ErrSecretGeneration
	}
	// Store plaintext temporarily for FCM notification (never persisted).
	entry.CommandSecret = secret

	// Hash the secret for secure storage in DB.
	secretHash, err := password.HashSecret(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to hash command secret: %w", err)
	}
	entry.CommandSecretHash = secretHash

	// Create device in devices table.
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

	// Transition to APPROVED.
	entry.Status = inbox.StatusApproved
	entry.ApprovedAt = PtrToInt64(now.UnixMilli())
	entry.UpdatedAt = now.UnixMilli()
	entry.Notes = notes

	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to update inbox entry: %w", err)
	}

	// Send FCM notification (best effort).
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

	s.logRegistrationAction(ctx, entry, "approved", operatorID, "")

	s.logger.Info("device approved by operator",
		"imei", entry.IMEI,
		"operator_id", operatorID,
		"fcm_push_sent", fcmPushSent,
	)

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

// RejectDevice handles operator rejection of a device registration.
// Transitions from PENDING/ACKNOWLEDGED -> REJECTED.
func (s *Service) RejectDevice(ctx context.Context, imei string, operatorID, orgID, notes string) (*AckResponse, error) {
	entry, err := s.repo.GetByIMEI(ctx, imei)
	if err != nil {
		if err == inbox.ErrInboxNotFound {
			return nil, ErrInboxNotFound
		}
		return nil, fmt.Errorf("failed to get inbox entry: %w", err)
	}

	if orgID != "" && entry.OrganizationID != "" && entry.OrganizationID != orgID {
		return nil, ErrInboxNotFound
	}

	if !entry.CanBeRejected() {
		return nil, ErrInboxCannotBeRejected
	}

	now := time.Now()
	entry.Status = inbox.StatusRejected
	entry.RejectedAt = PtrToInt64(now.UnixMilli())
	entry.UpdatedAt = now.UnixMilli()
	entry.OperatorID = operatorID
	entry.Notes = notes

	// Claim the entry for the operator's org on first action (public inbox.
	// entries have no org until an operator approves/rejects them).
	if entry.OrganizationID == "" && orgID != "" {
		entry.OrganizationID = orgID
	}

	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to update inbox entry: %w", err)
	}

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

// CreateInboxRequest creates a new inbox entry for device registration.
func (s *Service) CreateInboxRequest(ctx context.Context, req *InboxRequest) (*InboxEntryResponse, error) {
	if err := s.validateInboxRequest(req); err != nil {
		return nil, err
	}

	// Use transaction to prevent TOCTOU race on IMEI uniqueness.
	// The cleanup + create must be atomic to avoid duplicate entries.
	if s.txManager != nil {
		var response *InboxEntryResponse
		err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
			// Check if device already exists or has pending registration.
			existingEntry, err := s.checkDeviceAndInboxStatus(txCtx, req.IMEI)
			if err != nil {
				return err
			}

			if existingEntry != nil {
				response = s.buildInboxEntryResponse(existingEntry)
				return nil
			}

			// Create new inbox entry within transaction.
			entry, err := s.createInboxEntry(txCtx, req)
			if err != nil {
				return err
			}

			s.logAndRecordMetrics(txCtx, req.IMEI, entry, existingEntry)
			response = s.buildInboxEntryResponse(entry)
			return nil
		})
		if err != nil {
			return nil, err
		}
		return response, nil
	}

	// Fallback without transaction (should not happen in production).
	existingEntry, err := s.checkDeviceAndInboxStatus(ctx, req.IMEI)
	if err != nil {
		return nil, err
	}

	if existingEntry != nil {
		return s.buildInboxEntryResponse(existingEntry), nil
	}

	entry, err := s.createInboxEntry(ctx, req)
	if err != nil {
		return nil, err
	}

	s.logAndRecordMetrics(ctx, req.IMEI, entry, existingEntry)

	return s.buildInboxEntryResponse(entry), nil
}

func (s *Service) validateInboxRequest(req *InboxRequest) error {
	if req.IMEI == "" {
		return ErrInvalidIMEI
	}
	if !isValidIMEI(req.IMEI) {
		return ErrInvalidIMEI
	}
	if req.FCMToken != "" && !isValidFCMToken(req.FCMToken) {
		return ErrInvalidFCMToken
	}
	return nil
}

func (s *Service) checkDeviceAndInboxStatus(ctx context.Context, imei string) (*inbox.InboxEntry, error) {
	if s.deviceLookup != nil {
		if err := s.checkDeviceStatus(ctx, imei); err != nil {
			return nil, err
		}
	}
	return s.cleanupStaleInboxEntry(ctx, imei)
}

func (s *Service) checkDeviceStatus(ctx context.Context, imei string) error {
	existingDevice, err := s.deviceLookup.GetDeviceByIMEI(ctx, imei)
	// A "not found" device is the expected, happy path for new registrations.
	// The device lookup is backed by the device application service, which returns.
	// application.ErrDeviceNotFound; the domain repository may surface.
	// device.ErrNotFound. Accept either sentinel so the contract is robust to the.
	// concrete implementation behind the DeviceLookup interface.
	if err != nil && !errors.Is(err, device.ErrNotFound) && !errors.Is(err, deviceapp.ErrDeviceNotFound) {
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

func (s *Service) cleanupStaleInboxEntry(ctx context.Context, imei string) (*inbox.InboxEntry, error) {
	existingInboxEntry, err := s.repo.GetByIMEI(ctx, imei)
	if err != nil && err != inbox.ErrInboxNotFound {
		return nil, fmt.Errorf("failed to check existing inbox entry: %w", err)
	}
	if existingInboxEntry != nil {
		if existingInboxEntry.IsPending() {
			return nil, ErrAlreadyExists
		}
		s.logger.Info("will replace stale inbox entry for re-registration",
			"imei", imei,
			"old_status", existingInboxEntry.Status,
		)
	}
	return existingInboxEntry, nil
}

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

	// Use CreateOrReplace to atomically handle the case where a stale entry exists.
	// This avoids TOCTOU races between delete and create.
	if err := s.repo.CreateOrReplace(ctx, entry); err != nil {
		metrics.Get().RecordDeviceRegistrationFailure()
		return nil, fmt.Errorf("failed to create inbox entry: %w", err)
	}
	return entry, nil
}

func (s *Service) logAndRecordMetrics(ctx context.Context, imei string, entry *inbox.InboxEntry, existingInboxEntry *inbox.InboxEntry) {
	s.logRegistrationAction(ctx, entry, "pending", "", "")
	metrics.Get().RecordDeviceRegistrationSuccess()

	if existingInboxEntry != nil || s.deviceLookup != nil {
		if d, err := s.deviceLookup.GetDeviceByIMEI(ctx, imei); err == nil && d != nil && d.IsDeregistered() {
			metrics.Get().RecordDeviceReRegistration()
		}
	}
}

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
		ApprovingAt:       entry.ApprovingAt,
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

	if err := s.logRepo.Create(ctx, log); err != nil {
		s.logger.Error("failed to create audit log entry",
			"action", action,
			"imei", entry.IMEI,
			"operatorId", operatorID,
			"error", err,
		)
	}
}

// extractClientIP extracts the client IP from the context.
func extractClientIP(ctx context.Context) string {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if xff := ginCtx.GetHeader("X-Forwarded-For"); xff != "" {
			if idx := strings.Index(xff, ","); idx != -1 {
				return strings.TrimSpace(xff[:idx])
			}
			return strings.TrimSpace(xff)
		}
		if xri := ginCtx.GetHeader("X-Real-IP"); xri != "" {
			return xri
		}
		return ginCtx.ClientIP()
	}
	return ""
}

// extractUserAgent extracts the User-Agent from the context.
func extractUserAgent(ctx context.Context) string {
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
func isValidIMEI(imei string) bool {
	if len(imei) != 15 {
		return false
	}
	for _, c := range imei {
		if c < '0' || c > '9' {
			return false
		}
	}
	return isValidLuhn(imei)
}

// isValidLuhn validates a number using the Luhn algorithm.
func isValidLuhn(number string) bool {
	sum := 0
	isSecond := false
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
func isValidFCMToken(token string) bool {
	if len(token) < 100 {
		return false
	}
	for _, c := range token {
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}

// UpdateInboxEntry updates an inbox entry (e.g., operator notes).
func (s *Service) UpdateInboxEntry(ctx context.Context, imei, operatorID, orgID, notes string) (*InboxEntryResponse, error) {
	entry, err := s.repo.GetByIMEI(ctx, imei)
	if err != nil {
		if err == inbox.ErrInboxNotFound {
			return nil, ErrInboxNotFound
		}
		return nil, fmt.Errorf("failed to get inbox entry: %w", err)
	}

	if orgID != "" && entry.OrganizationID != "" && entry.OrganizationID != orgID {
		return nil, ErrInboxNotFound
	}

	if !entry.CanBeAcknowledged() {
		return nil, ErrInboxNotPending
	}

	entry.Notes = notes
	entry.UpdatedAt = time.Now().UnixMilli()

	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to update inbox entry: %w", err)
	}

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
func (s *Service) ResendApproval(ctx context.Context, imei, operatorID, orgID string) (*ResendResponse, error) {
	entry, err := s.repo.GetByIMEI(ctx, imei)
	if err != nil {
		if err == inbox.ErrInboxNotFound {
			return nil, ErrInboxNotFound
		}
		return nil, fmt.Errorf("failed to get inbox entry: %w", err)
	}

	if orgID != "" && entry.OrganizationID != "" && entry.OrganizationID != orgID {
		return nil, ErrInboxNotFound
	}

	if entry.Status != inbox.StatusApproved {
		return nil, ErrInboxNotApproved
	}

	if entry.CommandSecret == "" {
		return nil, fmt.Errorf("command secret not found for approved entry")
	}

	fcmPushSent := false
	if s.fcmNotifier != nil && entry.FCMToken != "" {
		wake := fcm.SilentWake{
			Token:         entry.FCMToken,
			Command:       "REGISTRATION_APPROVED",
			CommandSecret: entry.CommandSecret,
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
