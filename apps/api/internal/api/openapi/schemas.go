// Package openapi holds the request/response structs that back the
// swaggo annotations. swag init parses these into swagger.json definitions,
// which build_openapi3.py lifts into OpenAPI3 components.schemas so orval
// generates real TypeScript types instead of placeholder `object` shapes.
//
// Each struct here mirrors the exact JSON a handler emits (its gin.H keys),
// NOT the internal domain entity. Field names are snake_case to match the
// server's wire format. Keep these in sync with the handlers' c.JSON calls;
// the pre-commit swagger drift gate will fail if annotations and this file
// diverge from the generated swagger.json.
package openapi

import "time"

// ErrorResponse is the standard structured error body emitted by the error
// middleware for every 4xx/5xx.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	TraceID string            `json:"trace_id,omitempty"`
	Docs    string            `json:"docs,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Pagination is the shared paging envelope appended to list responses.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// DeletedResult is the generic `{ "deleted": true }` confirmation.
type DeletedResult struct {
	Deleted bool `json:"deleted"`
}

// ---- Alerts ---------------------------------------------------------------

type AlertRuleRequest struct {
	Name                  string  `json:"name"`
	Metric                string  `json:"metric"`
	Condition             string  `json:"condition"`
	WebhookURL            string  `json:"webhook_url"`
	OnNoData              string  `json:"on_no_data"`
	OnError               string  `json:"on_error"`
	Threshold             float64 `json:"threshold"`
	ForSeconds            int     `json:"for_seconds"`
	NotifyIntervalSeconds int     `json:"notify_interval_seconds"`
	Enabled               bool    `json:"enabled"`
}

type AlertInstance struct {
	Labels      map[string]string `json:"labels"`
	State       string            `json:"state"`
	Value       float64           `json:"value"`
	EvaluatedAt time.Time         `json:"evaluated_at"`
}

type AlertRule struct {
	ID                    string          `json:"id"`
	OrgID                 string          `json:"org_id"`
	Name                  string          `json:"name"`
	Metric                string          `json:"metric"`
	Condition             string          `json:"condition"`
	Threshold             float64         `json:"threshold"`
	ForSeconds            int             `json:"for_seconds"`
	NotifyIntervalSeconds int             `json:"notify_interval_seconds"`
	Enabled               bool            `json:"enabled"`
	WebhookURL            string          `json:"webhook_url"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	OnNoData              string          `json:"on_no_data"`
	OnError               string          `json:"on_error"`
	Instances             []AlertInstance `json:"instances"`
}

type AlertRuleListResult struct {
	Rules []AlertRule `json:"rules"`
}

type AlertHistoryEvent struct {
	ID        string    `json:"id"`
	RuleID    string    `json:"rule_id"`
	FromState string    `json:"from_state"`
	ToState   string    `json:"to_state"`
	Value     float64   `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

type AlertHistoryResult struct {
	Events []AlertHistoryEvent `json:"events"`
}

type AlertEvaluateResult struct {
	RuleID       string `json:"rule_id"`
	Transitioned int    `json:"transitioned"`
}

// ---- Annotations ----------------------------------------------------------

type AnnotationRequest struct {
	Title     string   `json:"title"`
	Text      string   `json:"text"`
	Source    string   `json:"source"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	Tags      []string `json:"tags"`
}

type Annotation struct {
	ID        string     `json:"id"`
	OrgID     string     `json:"org_id"`
	Title     string     `json:"title"`
	Text      string     `json:"text"`
	Tags      []string   `json:"tags"`
	Source    string     `json:"source"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type AnnotationListResult struct {
	Annotations []Annotation `json:"annotations"`
}

// ---- Contact points -------------------------------------------------------

type ContactPointRequest struct {
	Name       string            `json:"name"`
	Channel    string            `json:"channel"`
	Secret     string            `json:"secret"`
	Config     map[string]string `json:"config"`
	TemplateID string            `json:"template_id"`
	Enabled    bool              `json:"enabled"`
}

type ContactPoint struct {
	ID         string            `json:"id"`
	OrgID      string            `json:"org_id"`
	Name       string            `json:"name"`
	Channel    string            `json:"channel"`
	Secret     bool              `json:"secret"`
	Config     map[string]string `json:"config"`
	TemplateID string            `json:"template_id"`
	Enabled    bool              `json:"enabled"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type ContactPointListResult struct {
	ContactPoints []ContactPoint `json:"contact_points"`
}

type ContactPointTestResult struct {
	Sent     bool      `json:"sent"`
	TestedAt time.Time `json:"tested_at"`
}

// ---- Service accounts -----------------------------------------------------

type CreateServiceAccountRequest struct {
	Name string `json:"name"`
}

type CreateServiceAccountTokenRequest struct {
	ExpiresAt *string  `json:"expires_at"`
	ServiceID string   `json:"service_id"`
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
}

type RotateServiceAccountTokenRequest struct {
	ExpiresAt *string  `json:"expires_at"`
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
}

type ServiceAccount struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Name      string    `json:"name"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ServiceAccountToken struct {
	ID        string     `json:"id"`
	ServiceID string     `json:"service_id"`
	Name      string     `json:"name"`
	KeyPrefix string     `json:"key_prefix"`
	Scopes    []string   `json:"scopes"`
	Valid     bool       `json:"valid"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at"`
}

type ServiceAccountListResult struct {
	ServiceAccounts []ServiceAccount `json:"service_accounts"`
}

type ServiceAccountTokenListResult struct {
	Tokens []ServiceAccountToken `json:"tokens"`
}

type ServiceAccountTokenCreated struct {
	ServiceAccountToken
	Secret string `json:"secret"`
}

type ServiceAccountTokenRotated struct {
	ServiceAccountToken
	Secret string `json:"secret"`
}

type RevokedResult struct {
	Revoked bool `json:"revoked"`
}

// ---- Config versions ------------------------------------------------------

type ConfigVersion struct {
	ID           string    `json:"id"`
	OrgID        string    `json:"org_id"`
	ResourceType string    `json:"resource_type"`
	Version      int       `json:"version"`
	Snapshot     any       `json:"snapshot"`
	ChangedBy    string    `json:"changed_by"`
	CreatedAt    time.Time `json:"created_at"`
}

type ConfigVersionListResult struct {
	Versions []ConfigVersion `json:"versions"`
}

type ConfigVersionRestoreResult struct {
	RestoredToVersion int `json:"restored_to_version"`
	Settings          any `json:"settings"`
}

// ---- API keys -------------------------------------------------------------

type CreateAPIKeyRequest struct {
	Name          string `json:"name"`
	Scope         string `json:"scope"`
	ExpiresInDays *int   `json:"expires_in_days,omitempty"`
}

type UpdateAPIKeyRequest struct {
	Name  *string `json:"name,omitempty"`
	Scope *string `json:"scope,omitempty"`
}

type APIKey struct {
	ID           string     `json:"id"`
	OperatorID   string     `json:"operator_id,omitempty"`
	Name         string     `json:"name"`
	KeyPrefix    string     `json:"key_prefix"`
	Scope        string     `json:"scope"`
	ExpiresAt    *time.Time `json:"expires_at"`
	IsActive     bool       `json:"is_active"`
	RequestCount int64      `json:"request_count"`
	LastRequest  *time.Time `json:"last_request_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	RevokedAt    *time.Time `json:"revoked_at"`
}

type APIKeyWithSecret struct {
	ID           string     `json:"id"`
	OperatorID   string     `json:"operator_id,omitempty"`
	Name         string     `json:"name"`
	KeyPrefix    string     `json:"key_prefix"`
	Scope        string     `json:"scope"`
	ExpiresAt    *time.Time `json:"expires_at"`
	IsActive     bool       `json:"is_active"`
	RequestCount int64      `json:"request_count"`
	LastRequest  *time.Time `json:"last_request_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	RevokedAt    *time.Time `json:"revoked_at"`
	APIKey       string     `json:"api_key"`
}

type APIKeyListResult struct {
	Keys                 []APIKey   `json:"keys"`
	Pagination           Pagination `json:"pagination"`
	MonthlyLimit         int        `json:"monthly_limit"`
	KeysCreatedThisMonth int        `json:"keys_created_this_month"`
}

// ---- Devices --------------------------------------------------------------

type DeviceStatus struct {
	DeviceID    string `json:"device_id"`
	Online      bool   `json:"online"`
	LastSeen    int64  `json:"last_seen"`
	AppVersion  string `json:"app_version"`
	DeviceClass string `json:"device_class"`
}

type DeviceEvent struct {
	ID        string    `json:"id"`
	DeviceID  string    `json:"device_id"`
	Type      string    `json:"type"`
	Data      any       `json:"data"`
	CreatedAt time.Time `json:"created_at"`
}

type DeviceEventListResult struct {
	Events []DeviceEvent `json:"events"`
}

type DeviceLog struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source,omitempty"`
}

type DeviceLogListResult struct {
	Logs       []DeviceLog `json:"logs"`
	Pagination Pagination  `json:"pagination"`
}

// ---- Commands -------------------------------------------------------------

type CommandRequest struct {
	Command           string                 `json:"command"`
	ConfirmationToken string                 `json:"confirmation_token,omitempty"`
	DispatchID        string                 `json:"dispatch_id,omitempty"`
	Nonce             string                 `json:"nonce"`
	Signature         string                 `json:"signature,omitempty"`
	Timestamp         int64                  `json:"timestamp"`
	Args              map[string]interface{} `json:"args,omitempty"`
}

type CommandDispatchResult struct {
	Status       string `json:"status"`
	DeviceOnline bool   `json:"device_online"`
	DispatchID   string `json:"dispatchId"`
	CommandID    string `json:"command_id"`
	ServerTime   int64  `json:"serverTime"`
}

type CommandStatus struct {
	DispatchID string `json:"dispatchId"`
	CommandID  string `json:"command_id"`
	DeviceID   string `json:"device_id"`
	Command    string `json:"command"`
	Status     string `json:"status"`
	ServerTime int64  `json:"serverTime"`
}

type CommandRetryResult struct {
	DispatchID string `json:"dispatchId"`
	CommandID  string `json:"command_id"`
	Retried    bool   `json:"retried"`
	ServerTime int64  `json:"serverTime"`
}

type CommandCancelResult struct {
	DispatchID string `json:"dispatchId"`
	Cancelled  bool   `json:"cancelled"`
	ServerTime int64  `json:"serverTime"`
}

type CommandPendingResult struct {
	Commands []any `json:"commands"`
}

// ---- Channels -------------------------------------------------------------

type ChannelSubscribeRequest struct {
	Scope string `json:"scope"`
}

type ChannelStatusResult struct {
	Org           string `json:"org"`
	ActiveStreams int    `json:"active_streams"`
}

type ChannelSubscribeResult struct {
	Subscribed string `json:"subscribed"`
}

type ChannelUnsubscribeResult struct {
	Unsubscribed string `json:"unsubscribed"`
}

// ---- Organizations --------------------------------------------------------

type CreateOrganizationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Role        string `json:"role"`
	MaxMembers  int    `json:"maxMembers"`
}

type SelectOrganizationRequest struct {
	OrganizationID string `json:"organization_id"`
}

// ---- Auth -----------------------------------------------------------------

type LoginRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	Remember          bool   `json:"remember"`
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`
}

type LoginResult struct {
	Operator          any    `json:"operator"`
	RequiresMFA       bool   `json:"requires_mfa"`
	MFASession        string `json:"mfa_session,omitempty"`
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`
}

// ---- Inbox ----------------------------------------------------------------

type InboxEntry struct {
	ID        string    `json:"id"`
	DeviceID  string    `json:"device_id"`
	Type      string    `json:"type"`
	Payload   any       `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type InboxListResult struct {
	Entries    []InboxEntry `json:"entries"`
	Pagination Pagination   `json:"pagination"`
}

type UpdateInboxEntryRequest struct {
	Status string `json:"status"`
}

// ---- Telemetry ------------------------------------------------------------

type TelemetryEntry struct {
	Timestamp time.Time          `json:"timestamp"`
	Metrics   map[string]float64 `json:"metrics"`
}

type TelemetryHistoryRequest struct {
	DeviceID  string `form:"device_id"`
	StartTime int64  `form:"start_time"`
	EndTime   int64  `form:"end_time"`
	Limit     int    `form:"limit"`
}

type TelemetryHistoryResponse struct {
	DeviceID   string           `json:"device_id"`
	Entries    []TelemetryEntry `json:"entries"`
	Pagination Pagination       `json:"pagination"`
}

// ---- Updates --------------------------------------------------------------

type UpdateVersionManifest struct {
	Version      string `json:"version"`
	APKFilename  string `json:"apk_filename"`
	APKSize      int64  `json:"apk_size"`
	SHA256       string `json:"sha256"`
	ReleaseType  string `json:"release_type"`
	ReleaseNotes string `json:"release_notes"`
	ReleasedAt   int64  `json:"released_at"`
	IsLatest     bool   `json:"is_latest"`
}

type UpdateChangelogEntry struct {
	Version string    `json:"version"`
	Date    time.Time `json:"date"`
	Type    string    `json:"type"`
	Notes   string    `json:"notes"`
}

type UpdateCheckRequest struct {
	CurrentVersion string `json:"current_version"`
}

type UpdateCheckResult struct {
	UpdateAvailable bool   `json:"update_available"`
	LatestVersion   string `json:"latest_version,omitempty"`
	CurrentVersion  string `json:"current_version"`
	ReleaseNotes    string `json:"release_notes,omitempty"`
	DownloadURL     string `json:"download_url,omitempty"`
	SHA256          string `json:"sha256,omitempty"`
	APKSize         int64  `json:"apk_size,omitempty"`
}

type DownloadProgressRequest struct {
	DeviceID string `json:"device_id"`
	Version  string `json:"version"`
	Progress int    `json:"progress"`
}

type DownloadProgressResult struct {
	Recorded bool `json:"recorded"`
}

// ---- Dashboard ------------------------------------------------------------

type DashboardStats struct {
	TotalDevices   int `json:"total_devices"`
	OnlineDevices  int `json:"online_devices"`
	OfflineDevices int `json:"offline_devices"`
	PendingDevices int `json:"pending_devices,omitempty"`
}

// ---- Diagnostics ----------------------------------------------------------

type DeviceInspection struct {
	DeviceID      string   `json:"device_id"`
	Online        bool     `json:"online"`
	LastSeen      int64    `json:"last_seen"`
	AppVersion    string   `json:"app_version"`
	FCMTokenValid bool     `json:"fcm_token_valid"`
	Battery       *float64 `json:"battery,omitempty"`
}

// ---- Usage stats / admin --------------------------------------------------

type UsageStatsSnapshot struct {
	CollectedAt time.Time        `json:"collected_at"`
	Toggles     map[string]bool  `json:"toggles"`
	Counts      UsageStatsCounts `json:"counts"`
}

type UsageStatsCounts struct {
	Devices         int `json:"devices"`
	Operators       int `json:"operators"`
	Organizations   int `json:"organizations"`
	ServiceAccounts int `json:"service_accounts"`
	AlertRules      int `json:"alert_rules"`
	ContactPoints   int `json:"contact_points"`
	Annotations     int `json:"annotations"`
}

type UpdateCheckerResult struct {
	UpdateAvailable bool   `json:"update_available"`
	LatestVersion   string `json:"latest_version,omitempty"`
	CurrentVersion  string `json:"current_version"`
	ReleaseNotes    string `json:"release_notes,omitempty"`
	Repo            string `json:"repo,omitempty"`
}
