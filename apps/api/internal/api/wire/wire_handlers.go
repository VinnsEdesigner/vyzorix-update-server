// Package wire provides dependency injection utilities for the API server.
package wire

import (
	"context"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/admin"
	authhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/auth"
	cmdhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/command"
	confirmationhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/confirmation"
	devicehandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/device"
	organizationhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/organization"
	updateshandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/updates"
	websockethandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/websocket"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/confirmation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	orgapplication "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	updatesapplication "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	domaincommand "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/appcheck"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/github"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"log/slog"
)

// HandlerDependencies contains all dependencies needed by handlers.
type HandlerDependencies struct {
	OAuthStateRepo        authhandlers.OAuthStateProvider
	FCMNotifier           fcm.Notifier
	OperatorRepo          operator.Repository
	IPIntelligence        *middleware.IPIntelligence
	CommandService        *command.Service
	Hub                   *hub.Hub
	RiskEvaluator         *domaincommand.RiskEvaluator
	ConfirmationService   *confirmation.Service
	EmailService          *emailService.Service
	EmailVerificationRepo *storage.EmailVerificationRepository
	Lockout               *middleware.Lockout
	DB                    *storage.SQLite
	AuditLogger           *audit.Logger
	GoogleVerifier        *infraauth.GoogleTokenVerifier
	DeviceService         *device.Service
	DeviceRepo            *storage.DeviceRepository
	AppCheckVerifier      *appcheck.Verifier
	ClientService         *client.Service
	Presenter             *response.Presenter
	SessionManager        *infraauth.SessionManager
	Log                   *slog.Logger
	HmacVerifier          *cryptohmac.Verifier
	UpdatesStorage        *storage.UpdatesStorage
	AuthService           *auth.AuthService
	DeviceSettingsService *device.DeviceSettingsService
	OrgService            *orgapplication.OrganizationService
	MemberService         *orgapplication.MemberService
	InvitationService     *orgapplication.InvitationService
	OrgSettingsService    *orgapplication.OrganizationSettingsService
	Config                config.Config
}

// HandlerSet contains all handler instances.
type HandlerSet struct {
	Auth *authhandlers.AllHandlers
	// DEPRECATED: DeviceRegister - /v1/device/register endpoint removed.
	// DeviceRegister     *devicehandlers.RegisterHandler.
	DeviceStatus          *devicehandlers.StatusHandler
	DeviceUpdater         *devicehandlers.UpdaterHandler
	DeviceList            *devicehandlers.ListHandler
	Devices               *devicehandlers.DevicesHandler
	DeviceSettings        *devicehandlers.SettingsHandler
	DeviceService         *device.Service
	Command               *cmdhandlers.ExecuteHandler
	Confirmation          *confirmationhandlers.Handler
	Stream                *websockethandlers.StreamHandler
	TelemetryHistory      *handlers.TelemetryHistoryHandler
	ConnectionStatus      *handlers.ConnectionStatusHandler
	AdminClients          *admin.ClientsHandler
	Updates               *updateshandlers.UpdatesHandler
	UpdatesService        *updatesapplication.Service
	Organization          *organizationhandlers.OrganizationHandler
	Invitation            *organizationhandlers.InvitationHandler
	Member                *organizationhandlers.MemberHandler
	Transfer              *devicehandlers.TransferHandler
	OrgService            *orgapplication.OrganizationService
	OrgSettingsService    *orgapplication.OrganizationSettingsService
	MemberService         *orgapplication.MemberService
	InvitationService     *orgapplication.InvitationService
	OrgSettings           *organizationhandlers.SettingsHandler
	DeviceSettingsService *device.DeviceSettingsService
	FCMNotifier           fcm.Notifier
	AppCheckVerifier      *appcheck.Verifier
}

// WireHandlers creates and wires all handler instances.
func WireHandlers(deps HandlerDependencies) *HandlerSet {
	hs := &HandlerSet{}

	// Auth handlers.
	hs.Auth = authhandlers.NewAllHandlers(&authhandlers.Dependencies{
		AuthService:     deps.AuthService,
		SessionManager:  deps.SessionManager,
		Config:          deps.Config,
		GoogleVerifier:  deps.GoogleVerifier,
		ClientService:   deps.ClientService,
		EmailService:    deps.EmailService,
		EmailVerifyRepo: deps.EmailVerificationRepo,
		Lockout:         deps.Lockout,
		OperatorRepo:    deps.OperatorRepo,
		AuditLogger:     deps.AuditLogger,
		IPIntelligence:  deps.IPIntelligence,
		Presenter:       deps.Presenter,
		OAuthStateRepo:  deps.OAuthStateRepo,
	})

	// Device handlers.
	// DEPRECATED: hs.DeviceRegister = devicehandlers.NewRegisterHandler(deps.DeviceService) // /v1/device/register removed.
	hs.DeviceStatus = devicehandlers.NewStatusHandler(deps.DeviceService)
	hs.DeviceUpdater = devicehandlers.NewUpdaterHandler(deps.DeviceService)
	hs.DeviceList = devicehandlers.NewListHandler(deps.DeviceService, deps.Hub)
	hs.Devices = devicehandlers.NewDevicesHandler(deps.DeviceService)
	hs.DeviceService = deps.DeviceService

	// Command handler. The risk evaluator gates dangerous commands; the audit.
	// logger records every execution attempt. Fall back to a no-op logger when.
	// no audit store is configured so the handler never holds a nil dependency.
	var aud cmdhandlers.AuditLogger = audit.NewNoOpLogger()
	if deps.AuditLogger != nil {
		aud = deps.AuditLogger
	}
	evaluator := deps.RiskEvaluator
	if evaluator == nil {
		evaluator = domaincommand.NewRiskEvaluator()
	}
	// Confirmation handler doubles as the confirmation consumer for the.
	// command handler. When no confirmation service is configured, pass nil so.
	// risky commands are blocked (the handler treats nil as "disabled").
	var confirmConsumer cmdhandlers.ConfirmationConsumer
	if deps.ConfirmationService != nil {
		confirmationHandler := confirmationhandlers.NewHandler(deps.ConfirmationService, deps.DeviceService, evaluator)
		hs.Confirmation = confirmationHandler
		confirmConsumer = confirmationHandler
	}
	hs.Command = cmdhandlers.NewExecuteHandler(deps.CommandService, deps.DeviceService, deps.Hub, deps.FCMNotifier, evaluator, aud, confirmConsumer)

	// WebSocket handler.
	hs.Stream = websockethandlers.NewStreamHandler(deps.Log, deps.Config, deps.Hub, *deps.HmacVerifier, deps.AuditLogger)
	// Telemetry history handler.
	var telemetryRepo *storage.TelemetryRepository
	if deps.DB != nil {
		telemetryRepo = storage.NewTelemetryRepository(deps.DB.DB())
	}

	hs.TelemetryHistory = handlers.NewTelemetryHistoryHandler(
		deps.Log,
		telemetryRepo,
		deps.DeviceRepo,
		nil,
	)

	// Connection status handler.
	hs.ConnectionStatus = handlers.NewConnectionStatusHandler(deps.Log, deps.Hub, deps.DeviceRepo)

	// Admin handlers.
	hs.AdminClients = admin.NewClientsHandler(deps.ClientService)

	// Updates handlers.
	if deps.UpdatesStorage != nil && deps.DeviceService != nil {
		// Create sub-services.
		versionsStatusSvc := updatesapplication.NewVersionsStatusService(deps.UpdatesStorage)
		versionsListSvc := updatesapplication.NewVersionsListService(deps.UpdatesStorage)
		changelogSvc := updatesapplication.NewChangelogService(deps.UpdatesStorage)
		exportSvc := updatesapplication.NewExportService(deps.UpdatesStorage)
		historySvc := updatesapplication.NewHistoryService(deps.UpdatesStorage)

		// GitHub sync: create client from config if token/repo are configured.
		var githubSyncSvc *github.SyncService
		if deps.Config.GitHubReleaseToken != "" && deps.Config.GitHubReleaseRepo != "" {
			githubClient := github.NewClient(
				"VinnsEdesigner", // owner - could be made configurable.
				deps.Config.GitHubReleaseRepo,
				deps.Config.GitHubReleaseToken,
			)
			githubSyncSvc = github.NewSyncService(githubClient, deps.UpdatesStorage, deps.Log)
		}
		syncSvc := updatesapplication.NewSyncService(deps.UpdatesStorage, githubSyncSvc)

		// PushService needs Hub (for WSS), FCM (for offline wake), CommandService (for persistence),.
		// and DeviceService (for FCM token lookup). All wired from deps.
		pushSvc := updatesapplication.NewPushService(
			deps.UpdatesStorage,
			deps.DeviceService,
			deps.Hub,
			deps.FCMNotifier,
			deps.CommandService,
			deps.Log,
		)

		// Create main service with all sub-services.
		updatesService := updatesapplication.NewService(
			deps.UpdatesStorage,
			versionsStatusSvc,
			versionsListSvc,
			changelogSvc,
			exportSvc,
			pushSvc,
			historySvc,
			syncSvc,
		)

		// Create rate limiter middleware for updates endpoints.
		updatesRateLimiters := middleware.NewUpdatesRateLimiterMiddleware(nil)

		hs.Updates = updateshandlers.NewUpdatesHandler(updatesService, pushSvc, updatesRateLimiters, deps.AuditLogger, deps.Config.GitHubWebhookSecret)
		hs.UpdatesService = updatesService
	}

	// Organization handlers.
	if deps.OrgService != nil && deps.MemberService != nil {
		hs.Organization = organizationhandlers.NewOrganizationHandler(deps.OrgService, deps.MemberService, deps.Presenter)
		hs.Invitation = organizationhandlers.NewInvitationHandler(deps.InvitationService, deps.MemberService, deps.Presenter)
		hs.Member = organizationhandlers.NewMemberHandler(deps.MemberService, deps.Presenter)
		hs.OrgService = deps.OrgService
		hs.OrgSettingsService = deps.OrgSettingsService
		hs.MemberService = deps.MemberService
		hs.InvitationService = deps.InvitationService

		// Organization settings handler.
		if deps.OrgSettingsService != nil {
			hs.OrgSettings = organizationhandlers.NewSettingsHandler(deps.OrgSettingsService, deps.MemberService, deps.Presenter)
		}
	}

	// Device handlers.
	hs.Devices = devicehandlers.NewDevicesHandler(deps.DeviceService)

	// Device settings handler.
	if deps.DeviceSettingsService != nil && deps.MemberService != nil {
		// Create membership checker function.
		membershipChecker := func(ctx context.Context, operatorID, orgID string) error {
			return deps.MemberService.CheckMembership(ctx, operatorID, orgID)
		}
		hs.DeviceSettings = devicehandlers.NewDeviceSettingsHandler(deps.DeviceSettingsService, membershipChecker, deps.Presenter)
		hs.DeviceSettingsService = deps.DeviceSettingsService
	}

	// Device transfer handler.
	if deps.DeviceService != nil {
		hs.Transfer = devicehandlers.NewTransferHandler(deps.DeviceService, deps.Presenter)
	}

	// Set AppCheck verifier for device attestation.
	hs.AppCheckVerifier = deps.AppCheckVerifier

	return hs
}
