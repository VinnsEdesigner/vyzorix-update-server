// Package handler provides the HTTP handler for GraphQL.
package handler

import (
	"net/http"
	"time"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/resolver"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/schema"
	"github.com/gin-gonic/gin"
	gql "github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	gqlerrors "github.com/graphql-go/graphql/gqlerrors"
)

// Config holds handler configuration.
type Config struct {
	Logger         Logger
	Resolver       *resolver.Resolver
	AuthMiddleware *middleware.AuthMiddleware
	PlaygroundPath string
	Env            string // Environment: "production" disables Playground
	MaxDepth       int    // Maximum query depth (default 15)
	MaxComplexity  int    // Maximum query complexity score (default 1000)
}

// Handler is the GraphQL HTTP handler.
type Handler struct {
	logger         Logger
	resolver       *resolver.Resolver
	authMiddleware *middleware.AuthMiddleware
	presenter      *ResponsePresenter
	playgroundPath string
	env            string
	schema         gql.Schema
	maxDepth       int
	maxComplexity  int
}

// NewHandler creates a new GraphQL handler.
func NewHandler(cfg *Config) (*Handler, error) {
	gqlSchema, err := schema.BuildSchema(cfg.Resolver)
	if err != nil {
		return nil, err
	}

	path := cfg.PlaygroundPath
	if path == "" {
		path = "/playground"
	}

	maxDepth := cfg.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 15 // Default max depth
	}

	maxComplexity := cfg.MaxComplexity
	if maxComplexity <= 0 {
		maxComplexity = 1000 // Default max complexity
	}

	h := &Handler{
		schema:         gqlSchema,
		resolver:       cfg.Resolver,
		authMiddleware: cfg.AuthMiddleware,
		playgroundPath: path,
		env:            cfg.Env,
		logger:         cfg.Logger,
		presenter:      NewResponsePresenter(),
		maxDepth:       maxDepth,
		maxComplexity:  maxComplexity,
	}

	return h, nil
}

// Request represents a GraphQL request.
type Request struct {
	Variables     map[string]interface{} `json:"variables"`
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
}

// Response represents a GraphQL response.
type Response struct {
	Data   interface{} `json:"data,omitempty"`
	Errors interface{} `json:"errors,omitempty"`
}


func (h *Handler) validateQuery(query string) (int, int, error) {
	// Parse the query to get the AST
	doc, err := parser.Parse(parser.ParseParams{
		Source: query,
	})
	if err != nil {
		return 0, 0, err
	}

	// Calculate depth and complexity
	depth := 0
	complexity := 0

	for _, definition := range doc.Definitions {
		switch def := definition.(type) {
		case *ast.OperationDefinition:
			d, c := h.calculateDepthAndComplexity(def.SelectionSet)
			if d > depth {
				depth = d
			}
			complexity += c
		case *ast.FragmentDefinition:
			d, c := h.calculateDepthAndComplexity(def.SelectionSet)
			if d > depth {
				depth = d
			}
			complexity += c
		}
	}

	return depth, complexity, nil
}

// calculateDepthAndComplexity recursively calculates depth and complexity of a selection set.
func (h *Handler) calculateDepthAndComplexity(selectionSet *ast.SelectionSet) (int, int) {
	if selectionSet == nil {
		return 0, 0
	}

	maxChildDepth := 0
	complexity := 1 // Each selection adds at least 1 to complexity

	for _, selection := range selectionSet.Selections {
		switch sel := selection.(type) {
		case *ast.Field:
			if sel.SelectionSet != nil {
				childDepth, childComplexity := h.calculateDepthAndComplexity(sel.SelectionSet)
				if childDepth+1 > maxChildDepth {
					maxChildDepth = childDepth + 1
				}
				complexity += childComplexity
			}
		case *ast.InlineFragment:
			if sel.SelectionSet != nil {
				childDepth, childComplexity := h.calculateDepthAndComplexity(sel.SelectionSet)
				if childDepth+1 > maxChildDepth {
					maxChildDepth = childDepth + 1
				}
				complexity += childComplexity
			}
		case *ast.FragmentSpread:
			// Fragment spreads add to complexity due to potential nested selections
			complexity += 2
		}
	}

	if maxChildDepth == 0 {
		maxChildDepth = 1
	}

	return maxChildDepth, complexity
}

// Handle processes GraphQL requests.
func (h *Handler) Handle(c *gin.Context) {
	startTime := time.Now()

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, h.presenter.BadRequest("invalid request body"))
		return
	}

	
	if req.Query != "" {
		depth, complexity, err := h.validateQuery(req.Query)
		if err != nil {
			h.sendError(c, http.StatusBadRequest, h.presenter.BadRequest("invalid query syntax"))
			return
		}

		if depth > h.maxDepth {
			h.sendError(c, http.StatusBadRequest, h.presenter.Forbidden("query depth exceeds maximum allowed"))
			return
		}

		if complexity > h.maxComplexity {
			h.sendError(c, http.StatusBadRequest, h.presenter.Forbidden("query complexity exceeds maximum allowed"))
			return
		}
	}

	headers := map[string]string{
		"Cookie":        c.GetHeader("Cookie"),
		"Authorization": c.GetHeader("Authorization"),
	}

	op, err := h.authMiddleware.Authenticate(c.Request.Context(), headers)
	if err != nil {
		h.sendError(c, http.StatusUnauthorized, h.presenter.Unauthorized())
		return
	}

	ctx := gqlcontext.WithOperator(c.Request.Context(), op)
	ctx = gqlcontext.WithRequestMetadata(ctx, c.ClientIP(), c.GetHeader("User-Agent"))

	// Extract organizationId from URL parameter and add to context
	if orgID := c.Param("organizationId"); orgID != "" {
		ctx = gqlcontext.WithOrganizationID(ctx, orgID)
	}

	result := gql.Do(gql.Params{
		Schema:         h.Schema(),
		RequestString:  req.Query,
		VariableValues: req.Variables,
		OperationName:  req.OperationName,
		Context:        ctx,
	})

	h.logRequest(c, req, op.ID, len(result.Errors), startTime)

	if len(result.Errors) > 0 {
		h.sendError(c, http.StatusOK, h.formatErrors(result.Errors))
		return
	}

	c.JSON(http.StatusOK, Response{
		Data: result.Data,
	})
}

// logRequest logs the GraphQL request.
func (h *Handler) logRequest(c *gin.Context, req Request, operatorID string, errorCount int, start time.Time) {
	if h.logger == nil {
		return
	}

	h.logger.LogRequest(RequestLog{
		Timestamp:  start,
		Operation:  req.OperationName,
		Variables:  req.Variables,
		Duration:   time.Since(start),
		StatusCode: http.StatusOK,
		ErrorCount: errorCount,
		OperatorID: operatorID,
		ClientIP:   c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	})
}

// formatErrors converts GraphQL errors to response format.
func (h *Handler) formatErrors(errs []gqlerrors.FormattedError) Response {
	details := make([]ErrorDetail, 0, len(errs))

	for _, err := range errs {
		detail := ErrorDetail{Message: err.Message}

		if err.Extensions != nil {
			if code, ok := err.Extensions["code"].(string); ok {
				detail.Code = code
			}
		}

		details = append(details, detail)
	}

	return Response{
		Data:   nil,
		Errors: details,
	}
}

// sendError sends an error response.
func (h *Handler) sendError(c *gin.Context, status int, resp Response) {
	c.JSON(status, resp)
}

// Schema returns the GraphQL schema.
func (h *Handler) Schema() gql.Schema {
	return h.schema
}

// Playground serves the GraphQL playground.
func (h *Handler) Playground(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, playgroundHTML)
}

// Routes registers GraphQL routes with the Gin engine.
// Routes are scoped under /v1/orgs/:organizationId/graphql for multi-tenant isolation.
func (h *Handler) Routes(r *gin.Engine) {
	// Organization-scoped GraphQL routes for multi-tenant isolation
	orgGraphQL := r.Group("/v1/orgs/:organizationId/graphql")
	orgGraphQL.POST("", h.Handle)
	orgGraphQL.GET("", h.Handle)

	// Playground is only enabled in non-production environments
	if h.env != "production" {
		orgGraphQL.GET(h.playgroundPath, h.Playground)
	}

	// Legacy routes have been removed - they lacked proper org membership checks
	// Use /v1/orgs/:organizationId/graphql for multi-tenant isolation
}

// RegisterSubscriptions registers the WebSocket endpoint for subscriptions.
func (h *Handler) RegisterSubscriptions(r *gin.Engine, wsHandler func(*gin.Context)) {
	// Organization-scoped WebSocket endpoint for subscriptions
	orgGraphQL := r.Group("/v1/orgs/:organizationId/graphql")
	orgGraphQL.GET("/ws", wsHandler)

	// Legacy WebSocket route removed - use /v1/orgs/:organizationId/graphql/ws instead
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
