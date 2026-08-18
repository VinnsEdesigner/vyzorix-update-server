// Package audit provides audit logging functionality for security events.
package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// Action represents the type of security event being logged.
//
//nolint:gosec // G101 false positive: these are audit-action identifiers, not hardcoded credentials.
type Action string

const (
	ActionLoginSuccess           Action = "login_success"
	ActionLoginFailed            Action = "login_failed"
	ActionLogout                 Action = "logout"
	ActionRegister               Action = "register"
	ActionPasswordChange         Action = "password_change"
	ActionPasswordReset          Action = "password_reset"
	ActionEmailVerify            Action = "email_verify"
	ActionMFAEnabled             Action = "mfa_enabled"
	ActionMFADisabled            Action = "mfa_disabled"
	ActionMFALogin               Action = "mfa_login"
	ActionSessionRevoked         Action = "session_revoked"
	ActionAccountLocked          Action = "account_locked"
	ActionAccountUnlocked        Action = "account_unlocked"
	ActionCSRFFailure            Action = "csrf_failure"
	ActionSigningFailure         Action = "signing_failure"
	ActionRateLimitExceeded      Action = "rate_limit_exceeded"
	ActionAPIClientCreated       Action = "api_client_created"
	ActionAPIClientRevoked       Action = "api_client_revoked"
	ActionAPIClientSecretRotated Action = "api_client_secret_rotated" //nolint:gosec // G101 false positive: audit action identifier, not a secret
	ActionSigningKeyRotated      Action = "signing_key_rotated"
	ActionAdminAction            Action = "admin_action"
	// Updates-specific audit actions.
	ActionUpdatePushed      Action = "update_pushed"
	ActionUpdateCancelled   Action = "update_cancelled"
	ActionUpdateSyncStarted Action = "update_sync_started"
	ActionUpdateSyncFailed  Action = "update_sync_failed"
	// Settings-specific audit actions.
	ActionSettingsChanged      Action = "settings_changed"
	ActionWebhookTest          Action = "webhook_test"
	ActionWebhookSecretRotated Action = "webhook_secret_rotated"
	// MFA-specific audit actions.
	ActionMFAVerifyAttempt Action = "mfa_verify_attempt"
	ActionMFAVerifySuccess Action = "mfa_verify_success"
	ActionMFAVerifyFailed  Action = "mfa_verify_failed"
	// API key-specific audit actions.
	ActionAPIKeyCreated Action = "api_key_created"
	ActionAPIKeyUpdated Action = "api_key_updated"
	ActionAPIKeyRevoked Action = "api_key_revoked"
	ActionAPIKeyRotated Action = "api_key_rotated" //nolint:gosec // G101 false positive: audit action identifier, not a secret
	ActionAPIKeyFailed  Action = "api_key_failed"
	// Command execution audit actions (Phase 2 risk/audit).
	ActionCommandExecuted Action = "command_executed"
)

// Result represents the outcome of an action.
type Result string

const (
	ResultSuccess Result = "success"
	ResultFailure Result = "failure"
	ResultBlocked Result = "blocked"
	ResultPending Result = "pending"
)

// Entry represents a single audit log entry.
type Entry struct {
	CreatedAt    time.Time         `json:"created_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	ID           string            `json:"id"`
	OperatorID   string            `json:"operator_id,omitempty"`
	Action       Action            `json:"action"`
	ResourceType string            `json:"resource_type,omitempty"`
	ResourceID   string            `json:"resource_id,omitempty"`
	IPAddress    string            `json:"ip_address,omitempty"`
	UserAgent    string            `json:"user_agent,omitempty"`
	Result       Result            `json:"result"`
	// TraceID correlates this audit entry with request logs. Populated from the
	// tracing middleware so a security event can be joined to its access log.
	TraceID string `json:"trace_id,omitempty"`
	// RiskTier records the risk classification of the audited operation, when
	// applicable (e.g. command execution). Empty for non-risky events.
	RiskTier string `json:"risk_tier,omitempty"`
	// ActorType classifies the actor: "operator", "api_key", or "system".
	ActorType string `json:"actor_type,omitempty"`
	// ActorEmail is the operator's email, when available, for human-readable
	// audit queries (the OperatorID is an opaque ID).
	ActorEmail string `json:"actor_email,omitempty"`
	// OldValue/NewValue capture the before/after state of a mutated resource
	// for change-tracking compliance (e.g. settings updates, command state
	// transitions). JSON-encoded strings; empty when not applicable.
	OldValue string `json:"old_value,omitempty"`
	NewValue string `json:"new_value,omitempty"`
}

// Repository handles audit log persistence.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new audit repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Log writes an audit entry to the database.
func (r *Repository) Log(ctx context.Context, entry *Entry) error {
	metadataJSON := ""

	if entry.Metadata != nil {
		data, err := json.Marshal(entry.Metadata)
		if err != nil {
			return err
		}

		metadataJSON = string(data)
	}

	query := `
		INSERT INTO audit_logs (id, operator_id, action, resource_type, resource_id, ip_address, user_agent, metadata, result, created_at, trace_id, risk_tier, actor_type, actor_email, old_value, new_value)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		entry.ID,
		nullableString(entry.OperatorID),
		string(entry.Action),
		nullableString(entry.ResourceType),
		nullableString(entry.ResourceID),
		nullableString(entry.IPAddress),
		nullableString(entry.UserAgent),
		nullableString(metadataJSON),
		string(entry.Result),
		entry.CreatedAt.UnixMilli(),
		nullableString(entry.TraceID),
		nullableString(entry.RiskTier),
		nullableString(entry.ActorType),
		nullableString(entry.ActorEmail),
		nullableString(entry.OldValue),
		nullableString(entry.NewValue),
	)

	return err
}

// Query filters for searching audit logs.
type Query struct {
	StartTime    time.Time
	EndTime      time.Time
	OperatorID   string
	Action       Action
	ResourceType string
	ResourceID   string
	IPAddress    string
	Result       Result
	Limit        int
	Offset       int
}

// List retrieves audit entries matching the query.
func (r *Repository) List(ctx context.Context, q Query) ([]Entry, error) {
	if q.Limit <= 0 {
		q.Limit = 100
	}

	if q.Limit > 1000 {
		q.Limit = 1000
	}

	query := `
		SELECT id, operator_id, action, resource_type, resource_id, ip_address, user_agent, metadata, result, created_at
		FROM audit_logs WHERE 1=1
	`
	args := []any{}

	if q.OperatorID != "" {
		query += " AND operator_id = ?"

		args = append(args, q.OperatorID)
	}

	if q.Action != "" {
		query += " AND action = ?"

		args = append(args, string(q.Action))
	}

	if q.ResourceType != "" {
		query += " AND resource_type = ?"

		args = append(args, q.ResourceType)
	}

	if q.ResourceID != "" {
		query += " AND resource_id = ?"

		args = append(args, q.ResourceID)
	}

	if q.IPAddress != "" {
		query += " AND ip_address = ?"

		args = append(args, q.IPAddress)
	}

	if q.Result != "" {
		query += " AND result = ?"

		args = append(args, string(q.Result))
	}

	if !q.StartTime.IsZero() {
		query += " AND created_at >= ?"

		args = append(args, q.StartTime.UnixMilli())
	}

	if !q.EndTime.IsZero() {
		query += " AND created_at <= ?"

		args = append(args, q.EndTime.UnixMilli())
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"

	args = append(args, q.Limit, q.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var entries []Entry

	for rows.Next() {
		var e Entry

		var operatorID, resourceType, resourceID, ipAddress, userAgent, metadataJSON sql.NullString

		var action, result string

		var createdAt int64

		err := rows.Scan(&e.ID, &operatorID, &action, &resourceType, &resourceID, &ipAddress, &userAgent, &metadataJSON, &result, &createdAt)
		if err != nil {
			return nil, err
		}

		e.OperatorID = operatorID.String
		e.Action = Action(action)
		e.ResourceType = resourceType.String
		e.ResourceID = resourceID.String
		e.IPAddress = ipAddress.String
		e.UserAgent = userAgent.String
		e.Result = Result(result)
		e.CreatedAt = time.UnixMilli(createdAt)

		if metadataJSON.Valid && metadataJSON.String != "" {
			_ = json.Unmarshal([]byte(metadataJSON.String), &e.Metadata)
		}

		entries = append(entries, e)
	}

	return entries, rows.Err()
}

// Count returns the total number of entries matching the query.
func (r *Repository) Count(ctx context.Context, q Query) (int, error) {
	query := `SELECT COUNT(*) FROM audit_logs WHERE 1=1`
	args := []any{}

	if q.OperatorID != "" {
		query += " AND operator_id = ?"

		args = append(args, q.OperatorID)
	}

	if q.Action != "" {
		query += " AND action = ?"

		args = append(args, string(q.Action))
	}

	if q.Result != "" {
		query += " AND result = ?"

		args = append(args, string(q.Result))
	}

	if !q.StartTime.IsZero() {
		query += " AND created_at >= ?"

		args = append(args, q.StartTime.UnixMilli())
	}

	if !q.EndTime.IsZero() {
		query += " AND created_at <= ?"

		args = append(args, q.EndTime.UnixMilli())
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)

	return count, err
}

// Cleanup removes old audit entries beyond the retention period.
func (r *Repository) Cleanup(ctx context.Context, retentionDays int) (int, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays).UnixMilli()

	result, err := r.db.ExecContext(ctx, "DELETE FROM audit_logs WHERE created_at < ?", cutoff)
	if err != nil {
		return 0, err
	}

	n, _ := result.RowsAffected()

	return int(n), nil
}

func nullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}

	return sql.NullString{String: s, Valid: true}
}
