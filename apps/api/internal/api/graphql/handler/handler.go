// Package handler provides the HTTP handler for GraphQL.
package handler

import (
	"net/http"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/schema"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/resolver"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
)

// Config holds handler configuration.
type Config struct {
	Resolver       *resolver.Resolver
	AuthMiddleware *middleware.AuthMiddleware
	PlaygroundPath string
}

// Handler is the GraphQL HTTP handler.
type Handler struct {
	schema        graphql.Schema
	resolver      *resolver.Resolver
	authMiddleware *middleware.AuthMiddleware
	playgroundPath string
}

// NewHandler creates a new GraphQL handler.
func NewHandler(cfg *Config) (*Handler, error) {
	// Build the GraphQL schema
	gqlSchema, err := schema.BuildSchema(cfg.Resolver)
	if err != nil {
		return nil, err
	}

	path := cfg.PlaygroundPath
	if path == "" {
		path = "/playground"
	}

	return &Handler{
		schema:         gqlSchema,
		resolver:       cfg.Resolver,
		authMiddleware: cfg.AuthMiddleware,
		playgroundPath: path,
	}, nil
}

// Request represents a GraphQL request.
type Request struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

// Response represents a GraphQL response.
type Response struct {
	Data   interface{} `json:"data,omitempty"`
	Errors interface{} `json:"errors,omitempty"`
}

// Handle processes GraphQL requests.
func (h *Handler) Handle(c *gin.Context) {
	// Parse request
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": []gin.H{
				{"message": "invalid request body"},
			},
		})
		return
	}

	// Authenticate - extract operator from session cookie or Authorization header
	headers := map[string]string{
		"Cookie":        c.GetHeader("Cookie"),
		"Authorization": c.GetHeader("Authorization"),
	}

	op, err := h.authMiddleware.Authenticate(c.Request.Context(), headers)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errors": []gin.H{
				{"message": "authentication required"},
			},
		})
		return
	}

	// Add operator to context
	ctx := gqlcontext.WithOperator(c.Request.Context(), op)

	// Execute query
	result := graphql.Do(graphql.Params{
		Schema:         h.Schema(),
		RequestString:  req.Query,
		VariableValues: req.Variables,
		OperationName:  req.OperationName,
		Context:        ctx,
	})

	// Convert errors
	if len(result.Errors) > 0 {
		gqlErrs := make([]gin.H, 0, len(result.Errors))
		for _, err := range result.Errors {
			ext := make(map[string]interface{})
			if err.Extensions != nil {
				if code, ok := err.Extensions["code"].(string); ok {
					ext["code"] = code
				}
			}
			gqlErrs = append(gqlErrs, gin.H{
				"message":    err.Message,
				"extensions": ext,
			})
		}
		c.JSON(http.StatusOK, Response{
			Data:   result.Data,
			Errors: gqlErrs,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Data: result.Data,
	})
}

// Schema returns the GraphQL schema.
func (h *Handler) Schema() graphql.Schema {
	return h.schema
}

// Playground serves the GraphQL playground.
func (h *Handler) Playground(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, playgroundHTML)
}

// Routes registers GraphQL routes with the Gin engine.
func (h *Handler) Routes(r *gin.Engine) {
	r.POST("/graphql", h.Handle)
	r.GET("/graphql", h.Handle)
	r.GET(h.playgroundPath, h.Playground)
}

// RegisterSubscriptions registers the WebSocket endpoint for subscriptions.
// Call this separately after creating the subscription handler.
func (h *Handler) RegisterSubscriptions(r *gin.Engine, wsHandler func(*gin.Context)) {
	r.GET("/graphql/ws", wsHandler)
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
        settings: {
          'request.credentials': 'include',
        },
        headers: {
          // Example headers - uncomment and configure as needed
          // 'Authorization': 'Bearer YOUR_TOKEN_HERE',
        }
      })
    })
  </script>
</body>
</html>`