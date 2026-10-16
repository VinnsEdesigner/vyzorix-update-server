// Package apierrors maps domain and application errors to HTTP responses.
//
// It is the single place that knows which sentinel error means which
// ErrorCode, so handlers stop hand-writing an errors.Is chain
// (if errors.Is(err, ErrX) { presenter.NotFound(...); return }) at every call
// site.
//
// The layering is deliberate: this adapter may import domain, application and
// infrastructure packages, but none of them import it, so the dependency arrow
// only points one way (guarded by a test).
package apierrors

import (
	"errors"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	apporganization "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	alertdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/alert"
	annotationdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/annotation"
	commanddomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	configversiondomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/configversion"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	orgdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	sadomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/serviceaccount"

	"github.com/gin-gonic/gin"
)

// Mapped is the canonical HTTP rendering of an error: status, machine code and
// client-safe message. It is a value rather than a write, so callers can inspect
// or transform it before it reaches the wire, and it is testable without a
// gin.Context.
type Mapped struct {
	Details any
	Code    apperrors.ErrorCode
	Message string
	Status  int
}

// MapError classifies err and returns its canonical HTTP mapping.
//
// Order matters and is intentional:
//  1. a *ServerError already carries its own code and status - honour it;
//  2. a ValidationError renders 400 with field-level details;
//  3. a known domain/application sentinel maps through sentinelRules;
//  4. anything else is unexpected: 500 with a generic message, never echoing the
//     raw error to the client.
//
// MapError never returns a nil-ish mapping; an unclassified error maps to 500.
func MapError(err error) Mapped {
	if err == nil {
		return Mapped{Status: http.StatusInternalServerError, Code: apperrors.CodeInternalServerError,
			Message: "an unexpected error occurred"}
	}

	if se := apperrors.AsServerError(err); se != nil {
		return Mapped{Status: se.Code.HTTPStatusCode(), Code: se.Code, Message: se.Message, Details: se.Details}
	}

	if details, ok := apperrors.ValidationDetailsOf(err); ok {
		return Mapped{
			Status:  http.StatusBadRequest,
			Code:    apperrors.CodeValidationFailed,
			Message: "One or more fields have validation errors",
			Details: details,
		}
	}

	if mapped, ok := mapSentinel(err); ok {
		return mapped
	}

	return Mapped{Status: http.StatusInternalServerError, Code: apperrors.CodeInternalServerError,
		Message: "an unexpected error occurred"}
}

// sentinelRule pairs an error with the code it should produce.
type sentinelRule struct {
	err  error
	code apperrors.ErrorCode
}

// sentinelRules is the classification table. Add a sentinel here once and every
// handler that returns it is mapped correctly, replacing the per-handler
// errors.Is chain.
var sentinelRules = []sentinelRule{
	// Not found.
	{devicedomain.ErrNotFound, apperrors.CodeResourceNotFound},
	{devicedomain.ErrSettingsNotFound, apperrors.CodeResourceNotFound},
	{orgdomain.ErrNotFound, apperrors.CodeOrgNotFound},
	{orgdomain.ErrMemberNotFound, apperrors.CodeResourceNotFound},
	{orgdomain.ErrInvitationNotFound, apperrors.CodeResourceNotFound},
	{orgdomain.ErrSettingsNotFound, apperrors.CodeResourceNotFound},
	{commanddomain.ErrNotFound, apperrors.CodeResourceNotFound},
	{alertdomain.ErrNotFound, apperrors.CodeResourceNotFound},
	{annotationdomain.ErrNotFound, apperrors.CodeResourceNotFound},
	{configversiondomain.ErrNotFound, apperrors.CodeResourceNotFound},
	{sadomain.ErrNotFound, apperrors.CodeResourceNotFound},
	{application.ErrOperatorNotFound, apperrors.CodeResourceNotFound},
	{application.ErrDeviceNotFound, apperrors.CodeResourceNotFound},
	{application.ErrCommandNotFound, apperrors.CodeResourceNotFound},
	{application.ErrClientNotFound, apperrors.CodeResourceNotFound},

	// Already exists / conflict.
	{devicedomain.ErrInvalidTransition, apperrors.CodeResourceConflict},
	{orgdomain.ErrAlreadyExists, apperrors.CodeResourceAlreadyExists},
	{orgdomain.ErrMemberExists, apperrors.CodeResourceAlreadyExists},
	{orgdomain.ErrInvitationExists, apperrors.CodeResourceAlreadyExists},
	{orgdomain.ErrOrganizationExists, apperrors.CodeResourceAlreadyExists},
	{orgdomain.ErrAlreadyResponded, apperrors.CodeResourceConflict},
	{application.ErrDeviceExists, apperrors.CodeResourceAlreadyExists},
	{application.ErrDeviceHijack, apperrors.CodeResourceConflict},
	{application.ErrEmailExists, apperrors.CodeResourceAlreadyExists},
	{application.ErrUserExists, apperrors.CodeResourceAlreadyExists},

	// Authentication (401).
	{application.ErrInvalidCredentials, apperrors.CodeAuthInvalidCredentials},
	{application.ErrAccountLocked, apperrors.CodeAuthAccountLocked},
	{application.ErrMFARequired, apperrors.CodeAuthMFARequired},
	{application.ErrInvalidMFACode, apperrors.CodeAuthMFAInvalid},
	{application.ErrTokenExpired, apperrors.CodeAuthTokenExpired},
	{application.ErrTokenUsed, apperrors.CodeAuthTokenInvalid},
	{application.ErrEmailNotVerified, apperrors.CodeAuthTokenInvalid},
	{application.ErrUnauthorized, apperrors.CodeAuthTokenInvalid},

	// Authorization (403).
	{application.ErrForbidden, apperrors.CodeAuthzInsufficientPermissions},
	{application.ErrForbiddenIP, apperrors.CodeSecurityIPBlocked},
	{application.ErrNotOrgMember, apperrors.CodeAuthzOrgMembershipRequired},
	{application.ErrMembershipInactive, apperrors.CodeAuthzOrgMembershipRequired},
	{application.ErrOrganizationInactive, apperrors.CodeAuthzOrgMembershipRequired},
	{application.ErrInvalidOrganization, apperrors.CodeAuthzOrgMembershipRequired},
	{orgdomain.ErrForbidden, apperrors.CodeAuthzInsufficientPermissions},
	{orgdomain.ErrOperatorNotInOrganization, apperrors.CodeAuthzOrgMembershipRequired},
	{apporganization.ErrCannotModifyHigherRole, apperrors.CodeAuthzInsufficientPermissions},
	{orgdomain.ErrCannotInviteSuperAdmin, apperrors.CodeAuthzInsufficientPermissions},
	{orgdomain.ErrCannotRemoveLastSuperAdmin, apperrors.CodeAuthzInsufficientPermissions},
	{orgdomain.ErrCannotDeleteLastSuperAdmin, apperrors.CodeAuthzInsufficientPermissions},
	{orgdomain.ErrTransferOwnershipNotAllowed, apperrors.CodeAuthzInsufficientPermissions},
	{orgdomain.ErrDeviceTransferNotAllowed, apperrors.CodeAuthzInsufficientPermissions},

	// Validation (400).
	{application.ErrInvalidInput, apperrors.CodeValidationFailed},
	{application.ErrPasswordPolicy, apperrors.CodeValidationInvalidPassword},
	{application.ErrPasswordBreached, apperrors.CodeValidationPasswordBreached},
	{application.ErrEmailAlreadyVerified, apperrors.CodeValidationFailed},
	{application.ErrOAuthFailed, apperrors.CodeAuthInvalidCredentials},
	{orgdomain.ErrInvalidRole, apperrors.CodeValidationFailed},
	{orgdomain.ErrInvalidOperatorID, apperrors.CodeValidationFailed},
	{orgdomain.ErrInvalidMemberID, apperrors.CodeValidationFailed},
	{orgdomain.ErrInvalidOrganizationID, apperrors.CodeValidationFailed},
	{orgdomain.ErrInvalidMemberTransition, apperrors.CodeValidationFailed},
	{orgdomain.ErrInvalidInvitationTransition, apperrors.CodeValidationFailed},
	{orgdomain.ErrInvalidOrganizationTransition, apperrors.CodeValidationFailed},
	{orgdomain.ErrInvalidThreshold, apperrors.CodeValidationFailed},
	{orgdomain.ErrEmailMismatch, apperrors.CodeValidationFailed},
	{orgdomain.ErrMaxMembersReached, apperrors.CodeOrgMemberLimit},
	{orgdomain.ErrInvitationExpired, apperrors.CodeOrgInvitationExpired},
	{devicedomain.ErrInvalidDeviceID, apperrors.CodeValidationFailed},
	{devicedomain.ErrInvalidThreshold, apperrors.CodeValidationFailed},
	{commanddomain.ErrInvalidCommandTransition, apperrors.CodeValidationFailed},

	{apporganization.ErrAlreadyOrgMember, apperrors.CodeResourceAlreadyExists},
	{apporganization.ErrCannotInviteSelf, apperrors.CodeValidationFailed},
	{apporganization.ErrCannotRemoveLastSuperAdmin, apperrors.CodeAuthzInsufficientPermissions},
	{apporganization.ErrMaxInvitationsReached, apperrors.CodeOrgMemberLimit},
	{apporganization.ErrMaxOrgsReached, apperrors.CodeOrgMemberLimit},
	{apporganization.ErrOrgAtCapacity, apperrors.CodeOrgMemberLimit},

	// Rate limiting (429).
	{application.ErrRateLimited, apperrors.CodeRateLimitExceeded},

	// Device/command state.
	{application.ErrDeviceOnline, apperrors.CodeDeviceNotOnline},
	{application.ErrCommandFailed, apperrors.CodeDeviceCommandFailed},
	{commanddomain.ErrSendFailed, apperrors.CodeDeviceCommandFailed},
	{commanddomain.ErrNoTransport, apperrors.CodeDeviceCommandFailed},
	{commanddomain.ErrSigningFailed, apperrors.CodeInternalServerError},
	{commanddomain.ErrAckTimeout, apperrors.CodeDeviceCommandTimeout},
}

// mapSentinel walks the classification table.
func mapSentinel(err error) (Mapped, bool) {
	for _, rule := range sentinelRules {
		if errors.Is(err, rule.err) {
			return Mapped{
				Status:  rule.code.HTTPStatusCode(),
				Code:    rule.code,
				Message: rule.err.Error(),
			}, true
		}
	}
	return Mapped{}, false
}

// RespondError maps err and writes it to the client, then aborts the chain:
//
//	if err := h.svc.Do(ctx); err != nil {
//	    apierrors.RespondError(c, err)
//	    return
//	}
//
// The envelope is identical to responses.RespondStructured, so the wire
// contract for clients is unchanged.
func RespondError(c *gin.Context, err error) {
	m := MapError(err)
	responses.RespondStructuredAbortDetails(c, m.Status, m.Code, m.Message, m.Details)
}
