package apierrors

import (
	"encoding/json"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"

	"github.com/gin-gonic/gin"
)

// Response is the result of handling a request as a value rather than a write.
// A handler can build and return it, so its logic is unit-testable without a
// gin.Context, and middleware can inspect or rewrite it before it reaches the
// wire.
//
// The zero value is not useful; construct one with the helpers below.
type Response struct {
	// Headers are applied before the status and body. Nil means none.
	Headers http.Header
	// Body is the encoded response body. Nil means no body.
	Body []byte
	// Status is the HTTP status code. 0 is treated as 200 when written.
	Status int
}

// JSON returns a Response whose body is data encoded as JSON.
func JSON(status int, data any) Response {
	body, err := json.Marshal(data)
	if err != nil {
		// A value that cannot be encoded is a server bug; surface it as a 500 in
		// the standard envelope rather than emitting an empty 2xx.
		return Error(err)
	}
	return Response{Status: status, Body: body}
}

// OK is JSON(200, data).
func OK(data any) Response { return JSON(http.StatusOK, data) }

// Created is JSON(201, data).
func Created(data any) Response { return JSON(http.StatusCreated, data) }

// NoContent is a 204 with no body.
func NoContent() Response { return Response{Status: http.StatusNoContent} }

// Message returns a structured error Response for status with a client message.
// The envelope matches responses.RespondStructured, so the wire contract is
// identical to the existing helpers.
func Message(status int, message string) Response {
	return fromMapped(Mapped{Status: status, Code: responses.StatusToErrorCode(status), Message: message})
}

// Error maps err and returns the corresponding error Response. It never fails:
// an unclassified error becomes a 500 with a generic message.
func Error(err error) Response { return fromMapped(MapError(err)) }

// FromMapped encodes a Mapped value as a structured error Response.
func FromMapped(m Mapped) Response { return fromMapped(m) }

func fromMapped(m Mapped) Response {
	env := responses.StructuredErrorResponse{
		Error: responses.ErrorDetail{
			Code:    string(m.Code),
			Message: m.Message,
			Details: m.Details,
		},
	}
	body, err := json.Marshal(env)
	if err != nil {
		body = []byte(`{"error":{"code":"INTERNAL_SERVER_ERROR"}}`)
	}
	return Response{Status: m.Status, Body: body}
}

// WriteTo writes the Response to the gin context and aborts the handler chain.
//
// The trace id and docs URL are filled in here rather than at construction, so a
// Response carries no request-scoped state and stays reusable in tests.
func (v Response) WriteTo(c *gin.Context) {
	for k, vals := range v.Headers {
		for _, val := range vals {
			c.Header(k, val)
		}
	}

	status := v.Status
	if status == 0 {
		status = http.StatusOK
	}

	if len(v.Body) == 0 {
		c.Status(status)
		c.Abort()
		return
	}

	// A structured error body carries no trace id yet; fill it from the request
	// so the envelope matches responses.RespondStructured exactly.
	if status >= http.StatusBadRequest {
		if enriched, ok := responses.WithTrace(v.Body, responses.GetTraceID(c)); ok {
			v.Body = enriched
		}
	}

	c.Data(status, "application/json; charset=utf-8", v.Body)
	c.Abort()
}
