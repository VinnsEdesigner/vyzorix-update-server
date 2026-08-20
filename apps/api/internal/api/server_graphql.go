package api

import (
	"net/http"

	gqladapters "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/adapters"
	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	gqlmw "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/middleware"
	gqlresolver "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/resolver"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/schema"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/subscription"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	appsvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dashboard"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	diagnosticsapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/logs"
	appmetrics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
	appoperator "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/operator"
	orgapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	infrawebhook "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/webhook"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"

	"github.com/gin-gonic/gin"
	gql "github.com/graphql-go/graphql"
)

// RegisterGraphQL initializes and registers the GraphQL server with the API server.
// GraphQL routes use the same system middleware chain as REST tenant routes:
// rate limiting, cookie/session auth, token revocation, API key auth + scope.
// enforcement, organization context (from URL :org param), and organization.
// membership validation.
func (s *Server) RegisterGraphQL(
	deviceService *device.Service,
	deviceSettingsService *device.DeviceSettingsService,
	commandService *command.Service,
	historyService *command.HistoryService,
	dashboardSvc *dashboard.Service,
	logsSvc *logs.Service,
	metricsSvc *appmetrics.Service,
	telemetryRepo *storage.TelemetryRepository,
	logsRepo *storage.LogsRepository,
	metricsRepo *storage.MetricsRepository,
	wsHub *hub.Hub,
	updatesSvc *updates.Service,
	diagnosticsSvc *diagnosticsapp.Service,
	operatorRepo operator.Repository,
	settingsService *appsvc.ClientSettingsService,
	notificationSvc *appoperator.NotificationService,
	webhookClient *infrawebhook.Client,
	orgService *orgapp.OrganizationService,
	orgSettingsService *orgapp.OrganizationSettingsService,
	memberService *orgapp.MemberService,
	invitationService *orgapp.InvitationService,
) error {
	// Store InvitationService for graceful shutdown.
	s.InvitationService = invitationService

	// Get auth services from server config.
	authService := s.getAuthService()
	sessionManager := s.getSessionManager()

	// Create GraphQL presenter for audit logging.
	gqlPresenter := gqladapters.NewPresenter(s.AuditLogger)

	// The auth middleware is retained for the subscription handler's internal.
	// use. HTTP authentication is handled by the system middleware chain.
	authMw := gqlmw.NewAuthMiddleware(sessionManager, authService, s.log)

	// Create resolver.
	res := gqlresolver.NewResolver(
		deviceService,
		deviceSettingsService,
		commandService,
		historyService,
		dashboardSvc,
		logsSvc,
		metricsSvc,
		updatesSvc,
		diagnosticsSvc,
		wsHub,
		telemetryRepo,
		logsRepo,
		metricsRepo,
		nil, // FCM notifier.
		authMw,
		gqlPresenter,
		operatorRepo,
		settingsService,
		notificationSvc,
		webhookClient,
		orgService,
		orgSettingsService,
		memberService,
		invitationService,
		s.inboxService,
		s.commandAuthorizer,
	)

	// Wire the custom-permission-grants repository so the scoped permission
	// evaluator can union custom per-resource scopes onto role defaults (Issue 4).
	// Also wire the device-groups repository for team/device-group scoping (Issue 5).
	if s.DB != nil {
		res.GrantRepo = storage.NewGrantRepository(s.DB.DB())
		res.GroupRepo = storage.NewDeviceGroupRepository(s.DB.DB())
	}

	// Build schema.
	gqlSchema, err := schema.BuildSchema(res)
	if err != nil {
		return err
	}

	// Create handler.
	h := &gqlHandler{
		schema: gqlSchema,
	}

	// Create the GraphQL middleware group with the same chain as tenant routes.
	gqlGroup := s.engine.Group("/:org")
	gqlGroup.Use(s.authLimiter.Middleware())
	gqlGroup.Use(s.cookieAuth.Middleware())
	if s.revocationList != nil {
		gqlGroup.Use(middleware.AuthRevocationMiddleware(s.revocationList))
	}
	if s.tenantAPIKeyAuth != nil {
		gqlGroup.Use(s.tenantAPIKeyAuth.Middleware())
		gqlGroup.Use(s.tenantAPIKeyAuth.ScopeEnforcement(middleware.MethodToScope))
		if s.apiKeyRateLimiter != nil {
			gqlGroup.Use(middleware.APIKeyRateLimitMiddleware(s.apiKeyRateLimiter))
		}
	}
	// Set organization ID from the URL :org param into gin context so that.
	// the OrganizationMembership middleware can validate it.
	gqlGroup.Use(orgFromURLParamMiddleware)
	// Validate that the authenticated operator is a member of the organization.
	gqlGroup.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())

	// HTTP endpoints carry per-session HMAC request signatures (same scheme as.
	// REST tenant routes). WebSocket upgrades cannot set custom headers in the.
	// browser, so the WS route is registered separately on the parent group.
	// without signing.
	gqlHTTPGroup := gqlGroup.Group("")
	gqlHTTPGroup.Use(middleware.SessionSignatureMiddleware(s.sessionSignatureVerifier))

	// Register GraphQL routes through the protected group.
	gqlHTTPGroup.POST("/graphql", h.Handle)
	gqlHTTPGroup.GET("/graphql", h.Handle)

	// Register batch endpoint.
	batchHandler := &gqlBatchHandler{schema: gqlSchema}
	gqlHTTPGroup.POST("/graphql/batch", batchHandler.Handle)

	// Playground is only enabled in non-production environments.
	// It is served without middleware (it's just a static HTML page; actual.
	// query auth happens when requests are sent to /:org/graphql).
	if s.config.Env != "production" {
		s.engine.GET("/:org/playground", h.Playground)
	}

	// Create subscription handler. The WS endpoint goes through the same.
	// middleware group, so the operator is already authenticated by the time.
	// HandleWebSocket runs.
	subsHandler := subscription.NewHandler(&subscription.Config{
		Hub:         wsHub,
		Resolver:    res,
		AuthMw:      authMw,
		Logger:      s.log,
		AuditLogger: subscription.NewAuditLoggerAdapter(s.AuditLogger),
		Config:      s.config,
	})
	gqlGroup.GET("/graphql/ws", subsHandler.HandleWebSocket)

	s.log.Info("GraphQL server registered", "path", "/:org/graphql", "playground", "/:org/playground", "subscriptions", "/:org/graphql/ws", "middleware", "system-chain+signature")

	return nil
}

// orgFromURLParamMiddleware extracts the organization ID from the URL :org.
// parameter and stores it in the gin context under ContextKeyOrganizationID.
// This allows the OrganizationMembership middleware to validate the operator's.
// membership in the organization specified in the URL.
func orgFromURLParamMiddleware(c *gin.Context) {
	orgID := c.Param("org")
	if orgID == "" {
		responses.RespondStructuredAbort(c, http.StatusBadRequest,

			"organization ID required in URL path",
		)
		return
	}
	c.Set(middleware.ContextKeyOrganizationID, orgID)
	c.Next()
}

// gqlHandler is the GraphQL HTTP handler.
type gqlHandler struct {
	schema gql.Schema
}

// gqlRequest represents a GraphQL request.
type gqlRequest struct {
	Variables     map[string]interface{} `json:"variables"`
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
}

// Handle processes GraphQL requests. Authentication and organization membership.
// are enforced by the system middleware chain before this handler runs.
func (h *gqlHandler) Handle(c *gin.Context) {
	var req gqlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": []gin.H{{"message": "invalid request body"}},
		})

		return
	}

	// Extract operator from gin context (set by cookieAuth or tenantAPIKeyAuth).
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errors": []gin.H{{"message": "authentication required"}},
		})

		return
	}

	// Extract org from gin context (set by orgFromURLParamMiddleware).
	orgID := middleware.GetOrganizationID(c)

	// Add operator, organization, and MFA state to GraphQL context so the
	// command risk gate sees the same actor attributes as the REST path.
	ctx := gqlcontext.WithOperator(c.Request.Context(), op)
	ctx = gqlcontext.WithOrganizationID(ctx, orgID)
	if sess := middleware.GetSession(c); sess != nil && sess.MFAVerifiedAt != nil {
		ctx = gqlcontext.WithMFAVerified(ctx, true)
	}

	// Execute query.
	result := gql.Do(gql.Params{
		Schema:         h.schema,
		RequestString:  req.Query,
		VariableValues: req.Variables,
		OperationName:  req.OperationName,
		Context:        ctx,
	})

	// Convert errors.
	if len(result.Errors) > 0 {
		gqlErrs := make([]map[string]interface{}, 0, len(result.Errors))

		for _, err := range result.Errors {
			ext := make(map[string]interface{})

			if err.Extensions != nil {
				if code, ok := err.Extensions["code"].(string); ok {
					ext["code"] = code
				}
			}

			gqlErrs = append(gqlErrs, map[string]interface{}{
				"message":    err.Message,
				"extensions": ext,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"data":   result.Data,
			"errors": gqlErrs,
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result.Data,
	})
}

// Playground serves the GraphQL playground.
func (h *gqlHandler) Playground(c *gin.Context) {
	orgID := c.Param("org")
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, playgroundHTML(orgID))
}

func playgroundHTML(orgID string) string {
	return `<!DOCTYPE html>
<html>
<head>
  <title>Vyzorix GraphQL Playground</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
  <link rel="shortcut icon" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
  <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>
<body>
  <div id="root">
    <style>
      body { background-color: rgb(23, 42, 58); font-family: Open Sans, sans-serif; height: 90vh; }
      #root { height: 100%; width: 100%; display: flex; align-items: center; justify-content: center; }
      .loading { font-size: 32px; font-weight: 200; color: rgba(255, 255, 255, .6); margin-left: 28px; }
      img { width: 78px; height: 78px; }
      .title { font-weight: 400; }
    </style>
    <img src='https://cdn.jsdelivr.net/npm/graphql-playground-react/build/logo.png' alt=''>
    <div class="loading">Loading <span class="title">Vyzorix GraphQL Playground</span></div>
  </div>
  <script>
    window.addEventListener('load', function (event) {
      GraphQLPlayground.init(document.getElementById('root'), {
        endpoint: '/` + orgID + `/graphql',
        settings: { 'request.credentials': 'include' },
        headers: { 'Authorization': 'Bearer YOUR_TOKEN_HERE' }
      })
    })
  </script>
</body>
</html>`
}

// getAuthService returns the AuthService from auth handlers.
func (s *Server) getAuthService() *appsvc.AuthService {
	if s.authHandlers != nil {
		return s.authHandlers.AuthService
	}

	return nil
}

// getSessionManager returns the SessionManager from the server.
func (s *Server) getSessionManager() *infraauth.SessionManager {
	return s.sessionManager
}

// gqlBatchHandler handles GraphQL batch requests.
type gqlBatchHandler struct {
	schema gql.Schema
}

// gqlBatchRequest represents a batch of GraphQL requests.
type gqlBatchRequest struct {
	Variables     map[string]interface{} `json:"variables"`
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
}

// Handle processes GraphQL batch requests. Authentication and organization.
// membership are enforced by the system middleware chain before this handler runs.
func (h *gqlBatchHandler) Handle(c *gin.Context) {
	var requests []gqlBatchRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": []gin.H{{"message": "invalid request body"}},
		})
		return
	}

	if len(requests) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": []gin.H{{"message": "empty batch"}},
		})
		return
	}

	// Extract operator from gin context (set by cookieAuth or tenantAPIKeyAuth).
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errors": []gin.H{{"message": "authentication required"}},
		})
		return
	}

	// Extract org from gin context (set by orgFromURLParamMiddleware).
	orgID := middleware.GetOrganizationID(c)

	// Add operator, organization, and MFA state to GraphQL context so the
	// command risk gate sees the same actor attributes as the REST path.
	ctx := gqlcontext.WithOperator(c.Request.Context(), op)
	ctx = gqlcontext.WithOrganizationID(ctx, orgID)
	if sess := middleware.GetSession(c); sess != nil && sess.MFAVerifiedAt != nil {
		ctx = gqlcontext.WithMFAVerified(ctx, true)
	}

	// Execute all queries.
	results := make([]map[string]interface{}, len(requests))
	for i, req := range requests {
		result := gql.Do(gql.Params{
			Schema:         h.schema,
			RequestString:  req.Query,
			VariableValues: req.Variables,
			OperationName:  req.OperationName,
			Context:        ctx,
		})

		if len(result.Errors) > 0 {
			gqlErrs := make([]map[string]interface{}, 0, len(result.Errors))
			for _, err := range result.Errors {
				ext := make(map[string]interface{})
				if err.Extensions != nil {
					if code, ok := err.Extensions["code"].(string); ok {
						ext["code"] = code
					}
				}
				gqlErrs = append(gqlErrs, map[string]interface{}{
					"message":    err.Message,
					"extensions": ext,
				})
			}
			results[i] = map[string]interface{}{
				"data":   result.Data,
				"errors": gqlErrs,
			}
		} else {
			results[i] = map[string]interface{}{
				"data": result.Data,
			}
		}
	}

	c.JSON(http.StatusOK, results)
}
