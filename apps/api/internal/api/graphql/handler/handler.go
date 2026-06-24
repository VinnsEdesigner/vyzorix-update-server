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
	gqlerrors "github.com/graphql-go/graphql/gqlerrors"
)

// Config holds handler configuration.
type Config struct {
	Resolver       *resolver.Resolver
	AuthMiddleware *middleware.AuthMiddleware
	PlaygroundPath string
	Logger         Logger
}

// Handler is the GraphQL HTTP handler.
type Handler struct {
	schema         gql.Schema
	resolver       *resolver.Resolver
	authMiddleware *middleware.AuthMiddleware
	playgroundPath string
	logger         Logger
	presenter      *ResponsePresenter
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

	h := &Handler{
		schema:         gqlSchema,
		resolver:       cfg.Resolver,
		authMiddleware: cfg.AuthMiddleware,
		playgroundPath: path,
		logger:         cfg.Logger,
		presenter:      NewResponsePresenter(),
	}

	return h, nil
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
	startTime := time.Now()

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, h.presenter.BadRequest("invalid request body"))
		return
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
func (h *Handler) Routes(r *gin.Engine) {
	r.POST("/graphql", h.Handle)
	r.GET("/graphql", h.Handle)
	r.GET(h.playgroundPath, h.Playground)
}

// RegisterSubscriptions registers the WebSocket endpoint for subscriptions.
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
