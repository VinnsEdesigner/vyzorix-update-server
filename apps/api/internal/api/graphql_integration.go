// Package api provides GraphQL integration helpers.
package api

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	gqlerrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/resolver"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/schema"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
)

// RegisterGraphQL registers the GraphQL endpoint with the Gin engine.
// Call this after NewServer() to enable GraphQL API support.
func (s *Server) RegisterGraphQL(
	deviceService *device.Service,
	commandService *command.Service,
	telemetryRepo *storage.TelemetryRepository,
	hub *ws.Hub,
) error {
	// Create auth middleware
	authMw := middleware.NewAuthMiddleware(
		s.authHandlers.Dependencies.SessionManager,
		s.authHandlers.Dependencies.AuthService,
		s.log,
	)

	// Create resolver
	res := resolver.NewResolver(
		deviceService,
		commandService,
		hub,
		telemetryRepo,
		hub, // FCM notifier from hub
		authMw,
		s.log,
	)

	// Build schema
	gqlSchema, err := schema.BuildSchema(res)
	if err != nil {
		return err
	}

	// Create handler
	h := &gqlHandler{
		schema:        gqlSchema,
		authMiddleware: authMw,
	}

	// Register routes
	s.engine.POST("/graphql", h.Handle)
	s.engine.GET("/graphql", h.Handle)
	s.engine.GET("/playground", h.Playground)

	s.log.Info("GraphQL endpoint registered at /graphql and /playground")
	return nil
}

// gqlHandler is the GraphQL HTTP handler.
type gqlHandler struct {
	schema        graphql.Schema
	authMiddleware *middleware.AuthMiddleware
}

// gqlRequest represents a GraphQL request.
type gqlRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
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
	ctx := context.WithOperator(c.Request.Context(), op)

	// Execute query
	result := graphql.Do(graphql.Params{
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
			"data": result.Data,
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
