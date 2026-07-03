package api

import (
	"net/http"

	gqladapters "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/adapters"
	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	gqlmiddleware "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/resolver"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/schema"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/subscription"
	appsvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dashboard"
	diagnosticsapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/logs"
	appmetrics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"

	"github.com/gin-gonic/gin"
	gql "github.com/graphql-go/graphql"
)

// RegisterGraphQL initializes and registers the GraphQL server with the API server.
func (s *Server) RegisterGraphQL(
	deviceService *device.Service,
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
) error {
	// Get auth services from server config
	authService := s.getAuthService()
	sessionManager := s.getSessionManager()

	// Create auth middleware for GraphQL
	authMw := gqlmiddleware.NewAuthMiddleware(sessionManager, authService, s.log)

	// Create GraphQL presenter for audit logging
	gqlPresenter := gqladapters.NewPresenter(s.AuditLogger)

	// Create resolver with presenter
	res := resolver.NewResolver(
		deviceService,
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
		nil, // FCM notifier
		authMw,
		gqlPresenter,
	)

	// Build schema
	gqlSchema, err := schema.BuildSchema(res)
	if err != nil {
		return err
	}

	// Create handler
	h := &gqlHandler{
		schema:         gqlSchema,
		authMiddleware: authMw,
	}

	// Register routes
	s.engine.POST("/graphql", h.Handle)
	s.engine.GET("/graphql", h.Handle)
	s.engine.GET("/playground", h.Playground)

	// Create subscription handler
	subsHandler := subscription.NewHandler(&subscription.Config{
		Hub:         wsHub,
		Resolver:    res,
		AuthMw:      authMw,
		Logger:      s.log,
		AuditLogger: subscription.NewAuditLoggerAdapter(s.AuditLogger),
		Config:      s.config,
	})
	s.engine.GET("/graphql/ws", subsHandler.HandleWebSocket)

	s.log.Info("GraphQL server registered", "path", "/graphql", "playground", "/playground", "subscriptions", "/graphql/ws")

	return nil
}

// gqlHandler is the GraphQL HTTP handler.
type gqlHandler struct {
	authMiddleware *gqlmiddleware.AuthMiddleware
	schema         gql.Schema
}

// gqlRequest represents a GraphQL request.
type gqlRequest struct {
	Variables     map[string]interface{} `json:"variables"`
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
}

// Handle processes GraphQL requests.
func (h *gqlHandler) Handle(c *gin.Context) {
	var req gqlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": []gin.H{{"message": "invalid request body"}},
		})

		return
	}

	// Authenticate
	headers := map[string]string{
		"Cookie":        c.GetHeader("Cookie"),
		"Authorization": c.GetHeader("Authorization"),
	}

	op, err := h.authMiddleware.Authenticate(c.Request.Context(), headers)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errors": []gin.H{{"message": "authentication required"}},
		})

		return
	}

	// Add operator to context
	ctx := gqlcontext.WithOperator(c.Request.Context(), op)

	// Execute query
	result := gql.Do(gql.Params{
		Schema:         h.schema,
		RequestString:  req.Query,
		VariableValues: req.Variables,
		OperationName:  req.OperationName,
		Context:        ctx,
	})

	// Convert errors
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
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, playgroundHTML)
}

const playgroundHTML = `<!DOCTYPE html>
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
        endpoint: '/graphql',
        settings: { 'request.credentials': 'include' },
        headers: { 'Authorization': 'Bearer YOUR_TOKEN_HERE' }
      })
    })
  </script>
</body>
</html>`

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
