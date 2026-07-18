// Package wire provides dependency injection utilities for the API server.
package wire

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/admin"
	authhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/auth"
	cmdhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/command"
	devicehandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/device"
	organizationhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/organization"
	updateshandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/updates"
	websockethandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/websocket"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	orgapplication "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	updatesapplication "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/github"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"log/slog"
)

// HandlerDependencies contains all dependencies needed by handlers.
type HandlerDependencies struct {
	OperatorRepo              operator.Repository
	FCMNotifier              fcm.Notifier
	OAuthStateRepo           authhandlers.OAuthStateProvider
	Presenter                *response.Presenter
	Hub                      *ws.Hub
	EmailService             *emailService.Service
	Lockout                  *middleware.Lockout
	DB                       *storage.SQLite
	AuditLogger              *audit.Logger
	GoogleVerifier           *infraauth.GoogleTokenVerifier
	DeviceService            *device.Service
	DeviceRepo               *storage.DeviceRepository
	IPIntelligence           *middleware.IPIntelligence
	ClientService            *client.Service
	CommandService           *command.Service
	SessionManager           *infraauth.SessionManager
	Log                      *slog.Logger
	HmacVerifier             *cryptohmac.Verifier
	UpdatesStorage           *storage.UpdatesStorage
	AuthService              *auth.AuthService
	Config                   config.Config
	OrgService               *orgapplication.OrganizationService
	MemberService            *orgapplication.MemberService
	InvitationService        *orgapplication.InvitationService
	OrgSettingsService       *orgapplication.OrganizationSettingsService
	OrganizationRepo         orgapplication.OrganizationRepository
	MemberRepo               orgapplication.MemberRepository
	DeviceSettingsService    *device.DeviceSettingsService
}

// HandlerSet contains all handler instances.
type HandlerSet struct {
	Auth              *authhandlers.AllHandlers
	// DEPRECATED: DeviceRegister - /v1/device/register endpoint removed
	// DeviceRegister     *devicehandlers.RegisterHandler
	DeviceStatus      *devicehandlers.StatusHandler
	DeviceUpdater     *devicehandlers.UpdaterHandler
	DeviceList        *devicehandlers.ListHandler
	Devices           *devicehandlers.DevicesHandler
	DeviceSettings    *devicehandlers.SettingsHandler
	Command           *cmdhandlers.ExecuteHandler
	Stream            *websockethandlers.StreamHandler
	TelemetryHistory  *handlers.TelemetryHistoryHandler
	ConnectionStatus  *handlers.ConnectionStatusHandler
	AdminClients      *admin.ClientsHandler
	Updates           *updateshandlers.UpdatesHandler
	UpdatesService    *updatesapplication.Service
	Organization      *organizationhandlers.OrganizationHandler
	Invitation        *organizationhandlers.InvitationHandler
	Member            *organizationhandlers.MemberHandler
	Transfer          *devicehandlers.TransferHandler
	OrgService        *orgapplication.OrganizationService
	MemberService     *orgapplication.MemberService
	InvitationService *orgapplication.InvitationService
	OrgSettings       *organizationhandlers.SettingsHandler
}

// WireHandlers creates and wires all handler instances.
func WireHandlers(deps HandlerDependencies) *HandlerSet {
	hs := &HandlerSet{}

	// Auth handlers
	hs.Auth = authhandlers.NewAllHandlers(&authhandlers.Dependencies{
		AuthService:    deps.AuthService,
		SessionManager: deps.SessionManager,
		Config:         deps.Config,
		GoogleVerifier: deps.GoogleVerifier,
		ClientService:  deps.ClientService,
		EmailService:   deps.EmailService,
		Lockout:        deps.Lockout,
		OperatorRepo:   deps.OperatorRepo,
		AuditLogger:    deps.AuditLogger,
		IPIntelligence: deps.IPIntelligence,
		Presenter:      deps.Presenter,
		OAuthStateRepo: deps.OAuthStateRepo,
	})

	// Device handlers
	// DEPRECATED: hs.DeviceRegister = devicehandlers.NewRegisterHandler(deps.DeviceService) // /v1/device/register removed
	hs.DeviceStatus = devicehandlers.NewStatusHandler(deps.DeviceService)
	hs.DeviceUpdater = devicehandlers.NewUpdaterHandler(deps.DeviceService)
	hs.DeviceList = devicehandlers.NewListHandler(deps.DeviceService, deps.Hub)
	hs.Devices = devicehandlers.NewDevicesHandler(deps.DeviceService)

	// Command handler
	hs.Command = cmdhandlers.NewExecuteHandler(deps.CommandService, deps.DeviceService, deps.Hub, deps.FCMNotifier)

	// WebSocket handler
	hs.Stream = websockethandlers.NewStreamHandler(deps.Log, deps.Config, deps.Hub, *deps.HmacVerifier, deps.AuditLogger)
	// Telemetry history handler
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

	// Connection status handler
	hs.ConnectionStatus = handlers.NewConnectionStatusHandler(deps.Log, deps.Hub, deps.DeviceRepo)

	// Admin handlers
	hs.AdminClients = admin.NewClientsHandler(deps.ClientService)

	// Updates handlers
	if deps.UpdatesStorage != nil && deps.DeviceService != nil {
		// Create sub-services
		versionsStatusSvc := updatesapplication.NewVersionsStatusService(deps.UpdatesStorage)
		versionsListSvc := updatesapplication.NewVersionsListService(deps.UpdatesStorage)
		changelogSvc := updatesapplication.NewChangelogService(deps.UpdatesStorage)
		exportSvc := updatesapplication.NewExportService(deps.UpdatesStorage)
		historySvc := updatesapplication.NewHistoryService(deps.UpdatesStorage)

		// GitHub sync: create client from config if token/repo are configured.
		var githubSyncSvc *github.SyncService
		if deps.Config.GitHubReleaseToken != "" && deps.Config.GitHubReleaseRepo != "" {
			githubClient := github.NewClient(
				"VinnsEdesigner", // owner - could be made configurable
				deps.Config.GitHubReleaseRepo,
				deps.Config.GitHubReleaseToken,
			)
			githubSyncSvc = github.NewSyncService(githubClient, deps.UpdatesStorage, deps.Log)
		}
		syncSvc := updatesapplication.NewSyncService(deps.UpdatesStorage, githubSyncSvc)

		// PushService needs Hub (for WSS), FCM (for offline wake), CommandService (for persistence),
		// and DeviceService (for FCM token lookup). All wired from deps.
		pushSvc := updatesapplication.NewPushService(
			deps.UpdatesStorage,
			deps.DeviceService,
			deps.Hub,
			deps.FCMNotifier,
			deps.CommandService,
			deps.Log,
		)

		// Create main service with all sub-services
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

		// Create rate limiter middleware for updates endpoints
		updatesRateLimiters := middleware.NewUpdatesRateLimiterMiddleware(nil)

		hs.Updates = updateshandlers.NewUpdatesHandler(updatesService, pushSvc, updatesRateLimiters, deps.AuditLogger, deps.Config.GitHubWebhookSecret)
		hs.UpdatesService = updatesService
	}

	// Organization handlers
	if deps.OrgService != nil && deps.MemberService != nil {
		hs.Organization = organizationhandlers.NewOrganizationHandler(deps.OrgService, deps.MemberService, deps.Presenter)
		hs.Invitation = organizationhandlers.NewInvitationHandler(deps.InvitationService, deps.MemberService, deps.Presenter)
		hs.Member = organizationhandlers.NewMemberHandler(deps.MemberService, deps.OrgService, deps.Presenter)
		hs.OrgService = deps.OrgService
		hs.MemberService = deps.MemberService
		hs.InvitationService = deps.InvitationService

		// Organization settings handler
		if deps.OrgSettingsService != nil {
			hs.OrgSettings = organizationhandlers.NewSettingsHandler(deps.OrgSettingsService, deps.MemberService, deps.Presenter)
		}
	}

	// Device handlers
	hs.Devices = devicehandlers.NewDevicesHandler(deps.DeviceService)

	// Device settings handler
	if deps.DeviceSettingsService != nil && deps.MemberService != nil {
		// Create membership checker function
		membershipChecker := func(ctx context.Context, operatorID, orgID string) error {
			return deps.MemberService.CheckMembership(ctx, operatorID, orgID)
		}
		hs.DeviceSettings = devicehandlers.NewDeviceSettingsHandler(deps.DeviceSettingsService, membershipChecker, deps.Presenter)
	}

	// Device transfer handler
	if deps.DeviceService != nil {
		hs.Transfer = devicehandlers.NewTransferHandler(deps.DeviceService, deps.Presenter)
	}

	return hs
}
