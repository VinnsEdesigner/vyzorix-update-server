// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"errors"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/inbox"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/graphql-go/graphql"
)

// ============================================================.
// Inbox Query Resolvers.
// ============================================================.

// GetInbox resolves the inbox query.
func (r *Resolver) GetInbox(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}

	if r.InboxService == nil {
		return nil, r.Presenter.InternalError("inbox service not available")
	}

	status, _ := p.Args["status"].(string)
	if status == "" {
		status = "pending"
	}

	page, _ := p.Args["page"].(int)
	if page <= 0 {
		page = 1
	}

	limit, _ := p.Args["limit"].(int)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	result, err := r.InboxService.GetInbox(ctx, op.ID, orgID, status, page, limit)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get inbox")
	}

	requests := make([]map[string]interface{}, 0, len(result.Requests))
	for _, req := range result.Requests {
		requests = append(requests, map[string]interface{}{
			"id":                req.ID,
			"imei":              req.IMEI,
			"model":             req.Model,
			"manufacturer":      req.Manufacturer,
			"osVersion":         req.OSVersion,
			"appVersion":        req.AppVersion,
			"firebaseInstallId": req.FirebaseInstallID,
			"status":            req.Status,
			"createdAt":         req.CreatedAt,
			"approvedAt":        req.ApprovedAt,
			"rejectedAt":        req.RejectedAt,
			"notes":             req.Notes,
			"operatorId":        req.OperatorID,
		})
	}

	return map[string]interface{}{
		"requests": requests,
		"pagination": map[string]interface{}{
			"page":       result.Pagination.Page,
			"limit":      result.Pagination.Limit,
			"total":      result.Pagination.Total,
			"totalPages": result.Pagination.TotalPages,
		},
	}, nil
}

// GetInboxEntry resolves the inboxEntry query.
func (r *Resolver) GetInboxEntry(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	imei, ok := p.Args["imei"].(string)
	if !ok || imei == "" {
		return nil, r.Presenter.BadRequestError("IMEI is required")
	}

	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	if r.InboxService == nil {
		return nil, r.Presenter.InternalError("inbox service not available")
	}

	entry, err := r.InboxService.GetInboxEntry(ctx, imei, orgID)
	if err != nil {
		if err == inbox.ErrInboxNotFound {
			return nil, r.Presenter.NotFoundError("inbox entry not found")
		}
		return nil, r.Presenter.InternalError("failed to get inbox entry")
	}

	return map[string]interface{}{
		"id":                entry.ID,
		"imei":              entry.IMEI,
		"model":             entry.Model,
		"manufacturer":      entry.Manufacturer,
		"osVersion":         entry.OSVersion,
		"appVersion":        entry.AppVersion,
		"firebaseInstallId": entry.FirebaseInstallID,
		"status":            entry.Status,
		"createdAt":         entry.CreatedAt,
		"approvedAt":        entry.ApprovedAt,
		"rejectedAt":        entry.RejectedAt,
		"notes":             entry.Notes,
		"operatorId":        entry.OperatorID,
	}, nil
}

// ============================================================.
// Inbox Mutation Resolvers.
// ============================================================.

// AckInbox resolves the ackInbox mutation.
func (r *Resolver) AckInbox(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	imei, ok := p.Args["imei"].(string)
	if !ok || imei == "" {
		return nil, r.Presenter.BadRequestError("IMEI is required")
	}

	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}

	action, _ := p.Args["action"].(string)
	if action == "" {
		return nil, r.Presenter.BadRequestError("action is required")
	}

	notes, _ := p.Args["notes"].(string)

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	if r.InboxService == nil {
		return nil, r.Presenter.InternalError("inbox service not available")
	}

	result, err := r.InboxService.AckInbox(ctx, imei, action, op.ID, orgID, notes)
	if err != nil {
		se := inbox.ToServiceError(err)
		switch se.Code {
		case "not_found":
			return nil, r.Presenter.NotFoundError(se.Message)
		case "bad_request":
			return nil, r.Presenter.BadRequestError(se.Message)
		case "forbidden":
			return nil, r.Presenter.ForbiddenError(se.Message)
		default:
			return nil, r.Presenter.InternalError(se.Message)
		}
	}

	response := map[string]interface{}{
		"id":          result.ID,
		"imei":        result.IMEI,
		"status":      string(result.Status),
		"fcmPushSent": result.FCMPushSent,
		"notes":       result.Notes,
	}

	if result.ApprovedAt != nil {
		response["approvedAt"] = *result.ApprovedAt
	}
	if result.RejectedAt != nil {
		response["rejectedAt"] = *result.RejectedAt
	}

	return response, nil
}

// DeregisterDeviceGraphQL resolves the deregisterDevice mutation for GraphQL.
func (r *Resolver) DeregisterDeviceGraphQL(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}

	imei, ok := p.Args["imei"].(string)
	if !ok || imei == "" {
		return nil, r.Presenter.BadRequestError("IMEI is required")
	}

	hard, _ := p.Args["hard"].(bool)

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	result, err := r.DeviceService.DeregisterDeviceByOperator(ctx, imei, op.ID, orgID, hard)
	if err != nil {
		if errors.Is(err, device.ErrNotFound) {
			return nil, r.Presenter.NotFoundError("device not found or not owned by operator")
		}
		return nil, r.Presenter.InternalError("failed to deregister device")
	}

	r.Presenter.DeviceDelete(ctx, op.ID, imei)

	return map[string]interface{}{
		"imei":           result.IMEI,
		"status":         result.Status,
		"deregisteredAt": result.DeregisteredAt,
		"retentionUntil": result.RetentionUntil,
	}, nil
}
