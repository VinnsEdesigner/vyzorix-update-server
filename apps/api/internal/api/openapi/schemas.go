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
	Fields  map[string]string `json:"fields,omitempty"`
	Error   string            `json:"error"`
	Message string            `json:"message"`
	TraceID string            `json:"trace_id,omitempty"`
	Docs    string            `json:"docs,omitempty"`
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

// ---- Alerts ---------------------------------------------------------------.

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
	EvaluatedAt time.Time         `json:"evaluated_at"`
	Labels      map[string]string `json:"labels"`
	State       string            `json:"state"`
	Value       float64           `json:"value"`
}

type AlertRule struct {
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	WebhookURL            string          `json:"webhook_url"`
	Metric                string          `json:"metric"`
	Condition             string          `json:"condition"`
	ID                    string          `json:"id"`
	Name                  string          `json:"name"`
	OrgID                 string          `json:"org_id"`
	OnNoData              string          `json:"on_no_data"`
	OnError               string          `json:"on_error"`
	Instances             []AlertInstance `json:"instances"`
	Threshold             float64         `json:"threshold"`
	ForSeconds            int             `json:"for_seconds"`
	NotifyIntervalSeconds int             `json:"notify_interval_seconds"`
	Enabled               bool            `json:"enabled"`
}

type AlertRuleListResult struct {
	Rules []AlertRule `json:"rules"`
}

type AlertHistoryEvent struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	RuleID    string    `json:"rule_id"`
	FromState string    `json:"from_state"`
	ToState   string    `json:"to_state"`
	Value     float64   `json:"value"`
}

type AlertHistoryResult struct {
	Events []AlertHistoryEvent `json:"events"`
}

type AlertEvaluateResult struct {
	RuleID       string `json:"rule_id"`
	Transitioned int    `json:"transitioned"`
}

// ---- Annotations ----------------------------------------------------------.

type AnnotationRequest struct {
	Title     string   `json:"title"`
	Text      string   `json:"text"`
	Source    string   `json:"source"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	Tags      []string `json:"tags"`
}

type Annotation struct {
	StartTime time.Time  `json:"start_time"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	EndTime   *time.Time `json:"end_time"`
	ID        string     `json:"id"`
	OrgID     string     `json:"org_id"`
	Title     string     `json:"title"`
	Text      string     `json:"text"`
	Source    string     `json:"source"`
	Tags      []string   `json:"tags"`
}

type AnnotationListResult struct {
	Annotations []Annotation `json:"annotations"`
}

// ---- Contact points -------------------------------------------------------.

type ContactPointRequest struct {
	Name       string            `json:"name"`
	Channel    string            `json:"channel"`
	Secret     string            `json:"secret"`
	Config     map[string]string `json:"config"`
	TemplateID string            `json:"template_id"`
	Enabled    bool              `json:"enabled"`
}

type ContactPoint struct {
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Config     map[string]string `json:"config"`
	ID         string            `json:"id"`
	OrgID      string            `json:"org_id"`
	Name       string            `json:"name"`
	Channel    string            `json:"channel"`
	TemplateID string            `json:"template_id"`
	Secret     bool              `json:"secret"`
	Enabled    bool              `json:"enabled"`
}

type ContactPointListResult struct {
	ContactPoints []ContactPoint `json:"contact_points"`
}

type ContactPointTestResult struct {
	TestedAt time.Time `json:"tested_at"`
	Sent     bool      `json:"sent"`
}

// ---- Service accounts -----------------------------------------------------.

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
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Name      string    `json:"name"`
	Enabled   bool      `json:"enabled"`
}

type ServiceAccountToken struct {
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
	ID        string     `json:"id"`
	ServiceID string     `json:"service_id"`
	Name      string     `json:"name"`
	KeyPrefix string     `json:"key_prefix"`
	Scopes    []string   `json:"scopes"`
	Valid     bool       `json:"valid"`
}

type ServiceAccountListResult struct {
	ServiceAccounts []ServiceAccount `json:"service_accounts"`
}

type ServiceAccountTokenListResult struct {
	Tokens []ServiceAccountToken `json:"tokens"`
}

type ServiceAccountTokenCreated struct {
	Secret string `json:"secret"`
	ServiceAccountToken
}

type ServiceAccountTokenRotated struct {
	Secret string `json:"secret"`
	ServiceAccountToken
}

type RevokedResult struct {
	Revoked bool `json:"revoked"`
}

// ---- Config versions ------------------------------------------------------.

type ConfigVersion struct {
	CreatedAt    time.Time      `json:"created_at"`
	Snapshot     map[string]any `json:"snapshot"`
	ID           string         `json:"id"`
	OrgID        string         `json:"org_id"`
	ResourceType string         `json:"resource_type"`
	ChangedBy    string         `json:"changed_by"`
	Version      int            `json:"version"`
}

type ConfigVersionListResult struct {
	Versions []ConfigVersion `json:"versions"`
}

type ConfigVersionRestoreResult struct {
	Settings          map[string]any `json:"settings"`
	RestoredToVersion int            `json:"restored_to_version"`
}

// ---- API keys -------------------------------------------------------------.

type CreateAPIKeyRequest struct {
	ExpiresInDays *int   `json:"expires_in_days,omitempty"`
	Name          string `json:"name"`
	Scope         string `json:"scope"`
}

type UpdateAPIKeyRequest struct {
	Name  *string `json:"name,omitempty"`
	Scope *string `json:"scope,omitempty"`
}

type APIKey struct {
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
	LastRequest  *time.Time `json:"last_request_at"`
	RevokedAt    *time.Time `json:"revoked_at"`
	ID           string     `json:"id"`
	OperatorID   string     `json:"operator_id,omitempty"`
	Name         string     `json:"name"`
	KeyPrefix    string     `json:"key_prefix"`
	Scope        string     `json:"scope"`
	RequestCount int64      `json:"request_count"`
	IsActive     bool       `json:"is_active"`
}

type APIKeyWithSecret struct {
	UpdatedAt    time.Time  `json:"updated_at"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
	RevokedAt    *time.Time `json:"revoked_at"`
	LastRequest  *time.Time `json:"last_request_at"`
	KeyPrefix    string     `json:"key_prefix"`
	Scope        string     `json:"scope"`
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	OperatorID   string     `json:"operator_id,omitempty"`
	APIKey       string     `json:"api_key"`
	RequestCount int64      `json:"request_count"`
	IsActive     bool       `json:"is_active"`
}

type APIKeyListResult struct {
	Keys                 []APIKey   `json:"keys"`
	Pagination           Pagination `json:"pagination"`
	MonthlyLimit         int        `json:"monthly_limit"`
	KeysCreatedThisMonth int        `json:"keys_created_this_month"`
}

// ---- Devices --------------------------------------------------------------.

type DeviceStatus struct {
	DeviceID    string `json:"device_id"`
	AppVersion  string `json:"app_version"`
	DeviceClass string `json:"device_class"`
	LastSeen    int64  `json:"last_seen"`
	Online      bool   `json:"online"`
}

// ---- Device list / detail -------------------------------------------------.

type DeviceListItem struct {
	RegisteredAt *int64 `json:"registered_at,omitempty"`
	ID           string `json:"id"`
	IMEI         string `json:"imei,omitempty"`
	DeviceName   string `json:"device_name,omitempty"`
	Model        string `json:"model,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
	AppVersion   string `json:"app_version,omitempty"`
	Status       string `json:"status"`
	LastSeen     int64  `json:"last_seen,omitempty"`
	Online       bool   `json:"online,omitempty"`
}

type DeviceListResult struct {
	NextCursor *int             `json:"nextCursor,omitempty"`
	Devices    []DeviceListItem `json:"devices"`
	Total      int64            `json:"total"`
}

type DeviceDetailResult struct {
	RegisteredAt *int64 `json:"registered_at,omitempty"`
	ID           string `json:"id"`
	IMEI         string `json:"imei,omitempty"`
	DeviceName   string `json:"device_name,omitempty"`
	Model        string `json:"model,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
	AppVersion   string `json:"app_version,omitempty"`
	Status       string `json:"status"`
	LastSeen     int64  `json:"last_seen,omitempty"`
}

type DeviceCountResult struct {
	Count      int   `json:"count"`
	ServerTime int64 `json:"serverTime,omitempty"`
}

type DeviceTagsResult struct {
	Tags []string `json:"tags"`
}

type SetDeviceTagsRequest struct {
	Tags []string `json:"tags"`
}

type DeviceConfirmResult struct {
	RegisteredAt *int64 `json:"registered_at,omitempty"`
	DeviceID     string `json:"device_id"`
	IMEI         string `json:"imei"`
	ServerTime   int64  `json:"server_time"`
	Confirmed    bool   `json:"confirmed"`
	Online       bool   `json:"online"`
}

type DeviceConfirmRequest struct {
	IMEI          string `json:"imei"`
	CommandSecret string `json:"commandSecret"`
}

type DeviceSettingsResult struct {
	ID         string              `json:"id"`
	DeviceIMEI string              `json:"deviceImei"`
	CustomName string              `json:"customName,omitempty"`
	Location   string              `json:"location,omitempty"`
	Metadata   map[string]string   `json:"metadata,omitempty"`
	Thresholds *OperatorThresholds `json:"thresholds,omitempty"`
	CreatedAt  string              `json:"createdAt"`
	UpdatedAt  string              `json:"updatedAt"`
}

type UpdateDeviceSettingsRequest struct {
	CustomName *string             `json:"customName,omitempty"`
	Location   *string             `json:"location,omitempty"`
	Metadata   map[string]string   `json:"metadata,omitempty"`
	Thresholds *OperatorThresholds `json:"thresholds,omitempty"`
}

// ---- Metrics / telemetry --------------------------------------------------.

type MetricStatsDTO struct {
	Current float64 `json:"current"`
	Avg     float64 `json:"avg"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
}

type TelemetryFrameDTO struct {
	Timestamp   int64   `json:"timestamp"`
	Uptime      int64   `json:"uptime"`
	RiskScore   float64 `json:"riskScore"`
	ThermalTemp float64 `json:"thermalTemp"`
	BufferLevel float64 `json:"bufferLevel"`
}

type TelemetryStatsDTO struct {
	RiskScore   MetricStatsDTO `json:"riskScore"`
	ThermalTemp MetricStatsDTO `json:"thermalTemp"`
	BufferLevel MetricStatsDTO `json:"bufferLevel"`
}

type GetTelemetryResponse struct {
	Frames []TelemetryFrameDTO `json:"frames"`
	Stats  TelemetryStatsDTO   `json:"stats"`
}

// ---- Admin stats ----------------------------------------------------------.

type TopOperatorStat struct {
	OperatorID     string `json:"operator_id"`
	OperatorName   string `json:"operator_name"`
	TotalRequests  int64  `json:"total_requests"`
	ActiveKeyCount int    `json:"active_key_count"`
}

type GlobalAPIKeyStatsResult struct {
	RequestsByScope map[string]int64  `json:"requests_by_scope"`
	TopOperators    []TopOperatorStat `json:"top_operators"`
	TotalKeys       int               `json:"total_keys"`
	ActiveKeys      int               `json:"active_keys"`
	RevokedKeys     int               `json:"revoked_keys"`
	TotalOperators  int               `json:"total_operators"`
	MaxPerMonth     int               `json:"max_per_month"`
	TotalRequests   int64             `json:"total_requests"`
}

type OperatorAPIKeyStatsResult struct {
	OperatorID           string `json:"operator_id"`
	TotalKeys            int64  `json:"total_keys"`
	ActiveKeys           int64  `json:"active_keys"`
	RevokedKeys          int64  `json:"revoked_keys"`
	KeysCreatedThisMonth int    `json:"keys_created_this_month"`
	MonthlyLimit         int    `json:"monthly_limit"`
}

// ---- Tags -----------------------------------------------------------------.

type DeviceTagAddedResult struct {
	Added string `json:"added"`
}

type DeviceTagRemovedResult struct {
	Removed string `json:"removed"`
}

// ---- Admin client update --------------------------------------------------.

type UpdateAdminClientRequest struct {
	Name           *string  `json:"name,omitempty"`
	RateLimit      *int     `json:"rate_limit,omitempty"`
	IsActive       *bool    `json:"is_active,omitempty"`
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
	AllowedPaths   []string `json:"allowed_paths,omitempty"`
}

type DeviceTransferRequest struct {
	TargetOrganizationID string `json:"target_organization_id"`
}

type DeviceTransferResult struct {
	Message   string `json:"message,omitempty"`
	DeviceID  string `json:"device_id,omitempty"`
	FromOrgID string `json:"from_org_id,omitempty"`
	ToOrgID   string `json:"to_org_id,omitempty"`
	Success   bool   `json:"success"`
}

type DeviceFCMTokenRequest struct {
	FCMToken string `json:"fcmToken"`
}

type ConnectionStatusResult struct {
	DeviceID string `json:"device_id"`
	Status   string `json:"status,omitempty"`
	Online   bool   `json:"online"`
}

type ConnectionListResult struct {
	Connections []ConnectionStatusResult `json:"connections"`
}

type ConnectionMetricsResult struct {
	TotalConnections  int `json:"total_connections"`
	OnlineConnections int `json:"online_connections"`
}

type DeviceDisconnectResult struct {
	DeviceID     string `json:"deviceId"`
	OperatorID   string `json:"operatorId,omitempty"`
	Disconnected bool   `json:"disconnected"`
}

// ---- Admin clients --------------------------------------------------------.

type AdminClient struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ClientID  string `json:"clientId"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type AdminClientListResult struct {
	Clients []AdminClient `json:"clients"`
	Total   int           `json:"total"`
}

type AdminClientResult struct {
	Client AdminClient `json:"client"`
}

type SupportBundleResult struct {
	GeneratedAt   string `json:"generated_at"`
	Hostname      string `json:"hostname"`
	GoVersion     string `json:"go_version"`
	Goroutines    int    `json:"goroutines"`
	GoMaxProcs    int    `json:"go_max_procs"`
	GoNumCPU      int    `json:"go_num_cpu"`
	SchemaVersion int    `json:"schema_version,omitempty"`
	DeviceCount   int    `json:"device_count,omitempty"`
	OrgCount      int    `json:"org_count,omitempty"`
	OperatorCount int    `json:"operator_count,omitempty"`
}

type DeviceEvent struct {
	CreatedAt time.Time      `json:"created_at"`
	Data      map[string]any `json:"data"`
	ID        string         `json:"id"`
	DeviceID  string         `json:"device_id"`
	Type      string         `json:"type"`
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

// ---- Commands -------------------------------------------------------------.

type CommandRequest struct {
	Args              map[string]interface{} `json:"args,omitempty"`
	Command           string                 `json:"command"`
	ConfirmationToken string                 `json:"confirmation_token,omitempty"`
	DispatchID        string                 `json:"dispatch_id,omitempty"`
	Nonce             string                 `json:"nonce"`
	Signature         string                 `json:"signature,omitempty"`
	Timestamp         int64                  `json:"timestamp"`
}

type CommandDispatchResult struct {
	Status       string `json:"status"`
	DispatchID   string `json:"dispatchId"`
	CommandID    string `json:"command_id"`
	ServerTime   int64  `json:"serverTime"`
	DeviceOnline bool   `json:"device_online"`
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
	Commands []CommandResponse `json:"commands"`
}

type CommandResponse struct {
	ID         string `json:"id,omitempty"`
	DeviceID   string `json:"deviceId,omitempty"`
	Command    string `json:"command,omitempty"`
	DispatchID string `json:"dispatchId,omitempty"`
	Status     string `json:"status,omitempty"`
	Delivery   string `json:"delivery,omitempty"`
	Args       []byte `json:"args,omitempty"`
	ServerTime int64  `json:"serverTime,omitempty"`
}

// ---- Command history ------------------------------------------------------.

type CommandHistoryEntry struct {
	ID            string `json:"id,omitempty"`
	DispatchID    string `json:"dispatchId"`
	DeviceID      string `json:"deviceId"`
	Command       string `json:"command"`
	Status        string `json:"status"`
	FailureReason string `json:"failureReason,omitempty"`
	CreatedAt     int64  `json:"createdAt"`
	SentAt        int64  `json:"sentAt"`
	DeliveredAt   int64  `json:"deliveredAt,omitempty"`
	CompletedAt   int64  `json:"completedAt,omitempty"`
	LatencyMs     int64  `json:"latencyMs,omitempty"`
}

type CommandHistoryResult struct {
	Commands   []CommandHistoryEntry `json:"commands"`
	Pagination Pagination            `json:"pagination"`
}

// ---- Channels -------------------------------------------------------------.

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

// ---- Organizations --------------------------------------------------------.

type CreateOrganizationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Role        string `json:"role"`
	MaxMembers  int    `json:"maxMembers"`
}

type SelectOrganizationRequest struct {
	OrganizationID string `json:"organization_id"`
}

// Organization is the wire shape returned by organization endpoints.
type Organization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	MaxMembers  int    `json:"max_members"`
	IsActive    bool   `json:"is_active"`
	MemberCount int    `json:"member_count,omitempty"`
}

type OrganizationListResult struct {
	Organizations []Organization `json:"organizations"`
}

type SelectOrganizationResult struct {
	OrganizationID   string `json:"organization_id"`
	OrganizationName string `json:"organization_name"`
	Role             string `json:"role"`
}

type OrganizationMember struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	OperatorID     string `json:"operator_id"`
	Role           string `json:"role"`
	InvitedBy      string `json:"invited_by,omitempty"`
	JoinedAt       string `json:"joined_at,omitempty"`
	RemovedAt      string `json:"removed_at,omitempty"`
	Status         string `json:"status"`
	OperatorName   string `json:"operator_name,omitempty"`
	OperatorEmail  string `json:"operator_email,omitempty"`
}

type OrganizationMemberListResult struct {
	Members []OrganizationMember `json:"members"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}

type MessageResult struct {
	Message string `json:"message"`
}

// ---- Invitations ----------------------------------------------------------.

type CreateInvitationRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	OrgID string `json:"org_id,omitempty"`
}

type Invitation struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	Status         string `json:"status"`
	Token          string `json:"token,omitempty"`
	InvitedBy      string `json:"invited_by,omitempty"`
	InvitedAt      string `json:"invited_at,omitempty"`
	ExpiresAt      string `json:"expires_at,omitempty"`
}

type InvitationListResult struct {
	Invitations []Invitation `json:"invitations"`
}

type InvitationByTokenResult struct {
	ID               string `json:"id"`
	OrganizationID   string `json:"organization_id"`
	OrganizationName string `json:"organization_name,omitempty"`
	Email            string `json:"email"`
	Role             string `json:"role"`
	Status           string `json:"status"`
	InvitedAt        string `json:"invited_at,omitempty"`
	InviterName      string `json:"inviter_name,omitempty"`
	ExpiresAt        string `json:"expires_at,omitempty"`
}

// ---- Auth -----------------------------------------------------------------.

type OrganizationInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type LoginRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`
	Remember          bool   `json:"remember"`
}

type LoginResult struct {
	SelectedOrganization *OrganizationInfo  `json:"selected_organization,omitempty"`
	OperatorID           string             `json:"operator_id"`
	Email                string             `json:"email"`
	Name                 string             `json:"name"`
	LastOrganizationID   string             `json:"last_organization_id,omitempty"`
	SigningKey           string             `json:"signing_key"`
	MFASession           string             `json:"mfa_session,omitempty"`
	DeviceFingerprint    string             `json:"device_fingerprint,omitempty"`
	Organizations        []OrganizationInfo `json:"organizations,omitempty"`
	MFAEnabled           bool               `json:"mfa_enabled"`
	NeedsOrganization    bool               `json:"needs_organization"`
	RequiresMFA          bool               `json:"requires_mfa,omitempty"`
}

type LoginWithTokensResult struct {
	SelectedOrganization *OrganizationInfo  `json:"selected_organization,omitempty"`
	MFASession           string             `json:"mfa_session,omitempty"`
	Email                string             `json:"email"`
	Name                 string             `json:"name"`
	LastOrganizationID   string             `json:"last_organization_id,omitempty"`
	AccessToken          string             `json:"access_token"`
	RefreshToken         string             `json:"refresh_token"`
	SessionID            string             `json:"session_id"`
	SigningKey           string             `json:"signing_key"`
	OperatorID           string             `json:"operator_id"`
	DeviceFingerprint    string             `json:"device_fingerprint,omitempty"`
	Organizations        []OrganizationInfo `json:"organizations,omitempty"`
	ExpiresAt            int64              `json:"expires_at"`
	NeedsOrganization    bool               `json:"needs_organization"`
	RequiresMFA          bool               `json:"requires_mfa,omitempty"`
	MFAEnabled           bool               `json:"mfa_enabled"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role,omitempty"`
}

type RegisterResult struct {
	OperatorID string `json:"operator_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SessionID    string `json:"session_id"`
	ExpiresAt    int64  `json:"expires_at"`
}

type LogoutRequest struct {
	AllDevices bool `json:"all_devices"`
}

type MeResult struct {
	SelectedOrganization *OrganizationInfo  `json:"selected_organization,omitempty"`
	ID                   string             `json:"id"`
	Email                string             `json:"email"`
	Name                 string             `json:"name"`
	LastOrganizationID   string             `json:"last_organization_id,omitempty"`
	Organizations        []OrganizationInfo `json:"organizations"`
	MFAEnabled           bool               `json:"mfa_enabled"`
	EmailVerified        bool               `json:"email_verified"`
	NeedsOrganization    bool               `json:"needs_organization"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

type EmailVerifyRequest struct {
	Token string `json:"token"`
}

type EmailVerifyResult struct {
	Email    string `json:"email,omitempty"`
	Verified bool   `json:"verified"`
}

type PollVerificationResult struct {
	Status     string `json:"status"`
	Email      string `json:"email,omitempty"`
	EmailError string `json:"emailError,omitempty"`
}

type SuccessResult struct {
	Message string `json:"message,omitempty"`
	Success bool   `json:"success"`
}

type ResendVerificationRequest struct {
	Email string `json:"email"`
}

type CancelVerificationRequest struct {
	Email string `json:"email"`
}

// ---- MFA ------------------------------------------------------------------.

type MFAStatusResult struct {
	MFAEnabled bool `json:"mfa_enabled"`
}

type MFAEnrollResult struct {
	Secret string `json:"secret"`
	URI    string `json:"uri"`
}

type MFAVerifySetupRequest struct {
	Code  string `json:"code"`
	Token string `json:"token"`
}

type MFAEnableRequest struct {
	Code  string `json:"code"`
	Token string `json:"token"`
}

type MFAEnableResult struct {
	BackupCodes []string `json:"backup_codes"`
	Success     bool     `json:"success"`
}

type MFADisableRequest struct {
	Code string `json:"code"`
}

type MFABackupCodeRequest struct {
	Code string `json:"code"`
}

type MFABackupCodeResult struct {
	Valid bool `json:"valid"`
}

type MFARegenerateResult struct {
	BackupCodes []string `json:"backup_codes"`
}

type MFAVerifyRequest struct {
	OperatorID string `json:"operator_id"`
	Code       string `json:"code"`
}

type MFAVerifyResult struct {
	SessionID    string      `json:"session_id"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	SigningKey   string      `json:"signing_key"`
	Operator     MFAOperator `json:"operator"`
	ExpiresAt    int64       `json:"expires_at"`
	Success      bool        `json:"success"`
}

type MFAOperator struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	MFAEnabled bool   `json:"mfa_enabled"`
}

// ---- Lockout --------------------------------------------------------------.

type LockoutStatusResult struct {
	UnlockAt    *int64 `json:"unlock_at,omitempty"`
	Reason      string `json:"reason,omitempty"`
	Attempts    int    `json:"attempts"`
	MaxAttempts int    `json:"max_attempts,omitempty"`
	Locked      bool   `json:"locked"`
}

// ---- Sessions -------------------------------------------------------------.

type SessionInfo struct {
	ID                     string `json:"id"`
	IPAddress              string `json:"ip_address"`
	UserAgent              string `json:"user_agent"`
	CreatedAt              string `json:"created_at"`
	ExpiresAt              string `json:"expires_at"`
	SelectedOrganizationID string `json:"selected_organization_id,omitempty"`
	IsCurrent              bool   `json:"is_current"`
}

type SessionListResult struct {
	Sessions []SessionInfo `json:"sessions"`
	Total    int           `json:"total"`
}

type ConcurrentSessionsResult struct {
	Sessions      []SessionInfo `json:"sessions"`
	Count         int           `json:"count"`
	HasConcurrent bool          `json:"has_concurrent"`
}

type RevokeResult struct {
	Message      string `json:"message,omitempty"`
	RevokedCount int    `json:"revoked_count,omitempty"`
	Success      bool   `json:"success"`
}

// ---- Admin operators ------------------------------------------------------.

type AdminOperator struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Role          string `json:"role,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
	MFAEnabled    bool   `json:"mfa_enabled"`
	EmailVerified bool   `json:"email_verified"`
}

type AdminOperatorListResult struct {
	Operators []AdminOperator `json:"operators"`
	Total     int             `json:"total"`
}

type CreateOperatorRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role,omitempty"`
}

type UpdateOperatorRequest struct {
	Email         *string `json:"email,omitempty"`
	Name          *string `json:"name,omitempty"`
	Role          *string `json:"role,omitempty"`
	MFAEnabled    *bool   `json:"mfa_enabled,omitempty"`
	EmailVerified *bool   `json:"email_verified,omitempty"`
}

// ---- Client credentials ---------------------------------------------------.

type ClientCredential struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ClientID  string `json:"clientId"`
	Secret    string `json:"secret,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type ClientCredentialListResult struct {
	Clients []ClientCredential `json:"clients"`
}

type CreateClientCredentialRequest struct {
	Name string `json:"name"`
}

type UpdateClientCredentialRequest struct {
	Name *string `json:"name,omitempty"`
}

// ---- Operator settings ----------------------------------------------------.

type EmailNotifications struct {
	ThresholdBreach     bool `json:"thresholdBreach"`
	DeviceOffline       bool `json:"deviceOffline"`
	DeviceOnline        bool `json:"deviceOnline"`
	UpdateAvailable     bool `json:"updateAvailable"`
	CommandFailed       bool `json:"commandFailed"`
	RegistrationRequest bool `json:"registrationRequest"`
}

type PushNotifications struct {
	ThresholdBreach     bool `json:"thresholdBreach"`
	DeviceOffline       bool `json:"deviceOffline"`
	DeviceOnline        bool `json:"deviceOnline"`
	UpdateAvailable     bool `json:"updateAvailable"`
	CommandFailed       bool `json:"commandFailed"`
	RegistrationRequest bool `json:"registrationRequest"`
}

type WebhookNotifications struct {
	URL     string   `json:"url"`
	Secret  string   `json:"secret,omitempty"`
	Types   []string `json:"types"`
	Enabled bool     `json:"enabled"`
}

type NotificationSettings struct {
	Channels []string             `json:"channels"`
	Webhook  WebhookNotifications `json:"webhook"`
	Email    EmailNotifications   `json:"email"`
	Push     PushNotifications    `json:"push"`
	Enabled  bool                 `json:"enabled"`
}

type ClientSettings struct {
	ServerURL            string `json:"serverUrl"`
	DeviceID             string `json:"deviceId"`
	RequestTimeoutMs     int    `json:"requestTimeoutMs"`
	LogBufferLimit       int    `json:"logBufferLimit"`
	SignalHistoryLimit   int    `json:"signalHistoryLimit"`
	AutoReconnect        bool   `json:"autoReconnect"`
	StrictHmac           bool   `json:"strictHmac"`
	NotificationsEnabled bool   `json:"notificationsEnabled"`
}

type OperatorThresholds struct {
	RiskWarn    int `json:"riskWarn,omitempty"`
	RiskCrit    int `json:"riskCrit,omitempty"`
	ThermalWarn int `json:"thermalWarn,omitempty"`
	ThermalCrit int `json:"thermalCrit,omitempty"`
	BufferWarn  int `json:"bufferWarn,omitempty"`
	BufferCrit  int `json:"bufferCrit,omitempty"`
}

type OperatorSettingsResult struct {
	Notifications *NotificationSettings `json:"notifications,omitempty"`
	Client        *ClientSettings       `json:"client,omitempty"`
	Thresholds    *OperatorThresholds   `json:"thresholds,omitempty"`
}

type SettingsResponseResult struct {
	Notifications *NotificationSettings `json:"notifications,omitempty"`
	Client        *ClientSettings       `json:"client,omitempty"`
	Preferences   map[string]any        `json:"preferences,omitempty"`
}

type UpdateSettingsRequest struct {
	Client *ClientSettings `json:"client,omitempty"`
	Name   *string         `json:"name,omitempty"`
	Reset  bool            `json:"reset,omitempty"`
}

type ThresholdUpdateRequest struct {
	RiskWarn    *int `json:"riskWarn,omitempty"`
	RiskCrit    *int `json:"riskCrit,omitempty"`
	ThermalWarn *int `json:"thermalWarn,omitempty"`
	ThermalCrit *int `json:"thermalCrit,omitempty"`
	BufferWarn  *int `json:"bufferWarn,omitempty"`
	BufferCrit  *int `json:"bufferCrit,omitempty"`
}

type NotificationUpdateRequest struct {
	Enabled  *bool                 `json:"enabled,omitempty"`
	Channels *[]string             `json:"channels,omitempty"`
	Email    *EmailNotifications   `json:"email,omitempty"`
	Push     *PushNotifications    `json:"push,omitempty"`
	Webhook  *WebhookNotifications `json:"webhook,omitempty"`
}

type OperatorSettingsResultLegacy struct {
	Client        *ClientSettings `json:"client,omitempty"`
	ID            string          `json:"id"`
	Email         string          `json:"email"`
	Name          string          `json:"name"`
	Role          string          `json:"role,omitempty"`
	MFAEnabled    bool            `json:"mfa_enabled"`
	EmailVerified bool            `json:"email_verified"`
}

type UpdateNameRequest struct {
	Name string `json:"name"`
}

type ThresholdsResult struct {
	Thresholds *OperatorThresholds `json:"thresholds,omitempty"`
}

type PreferencesResult struct {
	Preferences map[string]any `json:"preferences"`
}

type WebhookTestResult struct {
	Error        string `json:"error,omitempty"`
	Message      string `json:"message,omitempty"`
	StatusCode   int    `json:"statusCode,omitempty"`
	ResponseTime int64  `json:"responseTime,omitempty"`
	Success      bool   `json:"success"`
}

type WebhookSecretResult struct {
	Secret string `json:"secret"`
}

// ---- Organization settings ------------------------------------------------.

type OrganizationSettingsResult struct {
	DefaultThresholds    *OperatorThresholds `json:"defaultThresholds,omitempty"`
	ID                   string              `json:"id"`
	OrganizationID       string              `json:"organizationId"`
	Timezone             string              `json:"timezone"`
	DateFormat           string              `json:"dateFormat"`
	CreatedAt            string              `json:"createdAt"`
	UpdatedAt            string              `json:"updatedAt"`
	AlertCooldownMinutes int                 `json:"alertCooldownMinutes"`
}

type UpdateOrganizationSettingsRequest struct {
	Timezone             *string             `json:"timezone,omitempty"`
	DateFormat           *string             `json:"dateFormat,omitempty"`
	AlertCooldownMinutes *int                `json:"alertCooldownMinutes,omitempty"`
	DefaultThresholds    *OperatorThresholds `json:"defaultThresholds,omitempty"`
}

// ---- Inbox ----------------------------------------------------------------.

type InboxRequest struct {
	IMEI              string `json:"imei"`
	DeviceName        string `json:"deviceName"`
	DeviceClass       string `json:"deviceClass"`
	Model             string `json:"model"`
	Manufacturer      string `json:"manufacturer"`
	OSVersion         string `json:"osVersion"`
	AppVersion        string `json:"appVersion"`
	FCMToken          string `json:"fcmToken"`
	FirebaseInstallID string `json:"firebaseInstallId"`
	IdempotencyKey    string `json:"idempotencyKey,omitempty"`
}

type InboxAckRequest struct {
	Action string `json:"action"`
	Notes  string `json:"notes,omitempty"`
}

type InboxEntryResponse struct {
	AcknowledgedAt    *int64 `json:"acknowledgedAt,omitempty"`
	ApprovingAt       *int64 `json:"approvingAt,omitempty"`
	ApprovedAt        *int64 `json:"approvedAt,omitempty"`
	RejectedAt        *int64 `json:"rejectedAt,omitempty"`
	Model             string `json:"model"`
	FirebaseInstallID string `json:"firebaseInstallId"`
	ID                string `json:"id"`
	Manufacturer      string `json:"manufacturer"`
	OSVersion         string `json:"osVersion"`
	AppVersion        string `json:"appVersion"`
	FCMToken          string `json:"fcmToken"`
	DeviceClass       string `json:"deviceClass,omitempty"`
	Status            string `json:"status"`
	OperatorID        string `json:"operatorId,omitempty"`
	DeviceName        string `json:"deviceName,omitempty"`
	IMEI              string `json:"imei"`
	Notes             string `json:"notes,omitempty"`
	CreatedAt         int64  `json:"createdAt"`
}

type InboxListResult struct {
	Requests   []InboxEntryResponse `json:"requests"`
	Pagination Pagination           `json:"pagination"`
}

type InboxAckResult struct {
	AcknowledgedAt *int64 `json:"acknowledgedAt,omitempty"`
	ApprovingAt    *int64 `json:"approvingAt,omitempty"`
	ApprovedAt     *int64 `json:"approvedAt,omitempty"`
	RejectedAt     *int64 `json:"rejectedAt,omitempty"`
	ID             string `json:"id"`
	IMEI           string `json:"imei"`
	Status         string `json:"status"`
	CommandSecret  string `json:"commandSecret,omitempty"`
	Notes          string `json:"notes,omitempty"`
	FCMPushSent    bool   `json:"fcmPushSent"`
}

type InboxResendResult struct {
	ID          string `json:"id"`
	IMEI        string `json:"imei"`
	Status      string `json:"status"`
	FCMPushSent bool   `json:"fcmPushSent"`
}

type UpdateInboxEntryRequest struct {
	Status string `json:"status"`
}

// ---- Telemetry ------------------------------------------------------------.

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

// ---- Updates --------------------------------------------------------------.

type UpdateVersionManifest struct {
	Version      string `json:"version"`
	APKFilename  string `json:"apk_filename"`
	SHA256       string `json:"sha256"`
	ReleaseType  string `json:"release_type"`
	ReleaseNotes string `json:"release_notes"`
	APKSize      int64  `json:"apk_size"`
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
	LatestVersion   string `json:"latest_version,omitempty"`
	CurrentVersion  string `json:"current_version"`
	ReleaseNotes    string `json:"release_notes,omitempty"`
	DownloadURL     string `json:"download_url,omitempty"`
	SHA256          string `json:"sha256,omitempty"`
	APKSize         int64  `json:"apk_size,omitempty"`
	UpdateAvailable bool   `json:"update_available"`
}

type DownloadProgressRequest struct {
	DeviceID string `json:"device_id"`
	Version  string `json:"version"`
	Progress int    `json:"progress"`
}

type DownloadProgressResult struct {
	Recorded bool `json:"recorded"`
}

// ---- Updater (public OTA) -------------------------------------------------.

type UpdaterVersionManifestResult struct {
	Version      string `json:"version"`
	APKFilename  string `json:"apk_filename"`
	APKSHA256    string `json:"apk_sha256"`
	ReleaseNotes string `json:"release_notes"`
	VersionCode  int    `json:"version_code"`
	APKSizeBytes int64  `json:"apk_size_bytes"`
}

type UpdaterCheckResult struct {
	Version         string `json:"version"`
	APKFilename     string `json:"apk_filename"`
	APKSHA256       string `json:"apk_sha256"`
	ReleaseNotes    string `json:"release_notes"`
	VersionCode     int    `json:"version_code"`
	APKSizeBytes    int64  `json:"apk_size_bytes"`
	UpdateAvailable bool   `json:"update_available"`
}

// ---- Updates (admin/scoped) -----------------------------------------------.

type UpdateVersionResponse struct {
	Version      string `json:"version"`
	ReleaseType  string `json:"releaseType"`
	Status       string `json:"status"`
	APKFilename  string `json:"apkFilename"`
	SHA256       string `json:"sha256"`
	ReleaseNotes string `json:"releaseNotes,omitempty"`
	APKSize      int64  `json:"apkSize"`
	ReleasedAt   int64  `json:"releasedAt"`
	IsLatest     bool   `json:"isLatest"`
}

type UpdateVersionListResult struct {
	Versions   []UpdateVersionResponse `json:"versions"`
	Pagination Pagination              `json:"pagination"`
}

type UpdateChangelogEntryResult struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Type    string `json:"type"`
	Notes   string `json:"notes"`
}

type UpdateChangelogResult struct {
	Changelog []UpdateChangelogEntryResult `json:"changelog"`
}

type UpdateSyncStatusInfo struct {
	Status        string `json:"status"`
	Error         string `json:"error,omitempty"`
	LastSyncAt    int64  `json:"lastSyncAt,omitempty"`
	NextSyncAt    int64  `json:"nextSyncAt,omitempty"`
	VersionsFound int    `json:"versionsFound,omitempty"`
}

type UpdateLatestVersionInfo struct {
	Version     string `json:"version"`
	ReleaseType string `json:"releaseType"`
	APKFilename string `json:"apkFilename"`
	SHA256      string `json:"sha256"`
	ReleasedAt  int64  `json:"releasedAt"`
	APKSize     int64  `json:"apkSize"`
}

type UpdateDeviceStatusInfo struct {
	CurrentVersion string `json:"currentVersion,omitempty"`
	NeedsUpdate    bool   `json:"needsUpdate"`
}

type UpdateStatusResult struct {
	Latest UpdateLatestVersionInfo `json:"latest,omitempty"`
	Device *UpdateDeviceStatusInfo `json:"device,omitempty"`
	Sync   UpdateSyncStatusInfo    `json:"sync"`
}

type UpdateSyncResponse struct {
	Status        string `json:"status"`
	Message       string `json:"message,omitempty"`
	StartedAt     int64  `json:"startedAt"`
	VersionsFound int    `json:"versionsFound,omitempty"`
}

type UpdateExportResult struct {
	Format     string                       `json:"format"`
	Versions   []UpdateVersionResponse      `json:"versions"`
	Changelog  []UpdateChangelogEntryResult `json:"changelog"`
	ExportedAt int64                        `json:"exportedAt"`
}

type UpdatePushDeviceCounts struct {
	Total        int `json:"total"`
	Pending      int `json:"pending"`
	Sent         int `json:"sent"`
	Acknowledged int `json:"acknowledged"`
	Failed       int `json:"failed"`
}

type UpdateFailedDevice struct {
	DeviceID string `json:"deviceId"`
	Reason   string `json:"reason"`
}

type UpdatePushResult struct {
	PushID        string                 `json:"pushId"`
	Version       string                 `json:"version"`
	InstallType   string                 `json:"installType"`
	ScheduledAt   *int64                 `json:"scheduledAt,omitempty"`
	InitiatedBy   string                 `json:"initiatedBy"`
	Status        string                 `json:"status"`
	DeviceIDs     []string               `json:"deviceIds"`
	FailedDevices []UpdateFailedDevice   `json:"failedDevices,omitempty"`
	Devices       UpdatePushDeviceCounts `json:"devices"`
	InitiatedAt   int64                  `json:"initiatedAt"`
}

type UpdatePushHistoryDeviceCounts struct {
	Pending      int `json:"pending,omitempty"`
	Sent         int `json:"sent,omitempty"`
	Acknowledged int `json:"acknowledged,omitempty"`
	Failed       int `json:"failed,omitempty"`
}

type UpdatePushHistoryEntry struct {
	CompletedAt *int64                        `json:"completedAt,omitempty"`
	CancelledAt *int64                        `json:"cancelledAt,omitempty"`
	ScheduledAt *int64                        `json:"scheduledAt,omitempty"`
	ID          string                        `json:"id"`
	Version     string                        `json:"version"`
	InstallType string                        `json:"installType"`
	Status      string                        `json:"status"`
	InitiatedBy string                        `json:"initiatedBy"`
	Devices     UpdatePushHistoryDeviceCounts `json:"devices"`
	DeviceCount int                           `json:"deviceCount"`
	InitiatedAt int64                         `json:"initiatedAt"`
}

type UpdatePushHistoryListResult struct {
	Pushes     []UpdatePushHistoryEntry `json:"pushes"`
	Pagination Pagination               `json:"pagination"`
}

type UpdatePushDetailDevice struct {
	ID             string `json:"id"`
	DeviceID       string `json:"deviceId"`
	DeviceName     string `json:"deviceName,omitempty"`
	Status         string `json:"status"`
	SentAt         *int64 `json:"sentAt,omitempty"`
	AcknowledgedAt *int64 `json:"acknowledgedAt,omitempty"`
	Error          string `json:"error,omitempty"`
}

type UpdatePushDetailResult struct {
	ScheduledAt *int64                   `json:"scheduledAt,omitempty"`
	CompletedAt *int64                   `json:"completedAt,omitempty"`
	CancelledAt *int64                   `json:"cancelledAt,omitempty"`
	ID          string                   `json:"id"`
	Version     string                   `json:"version"`
	InstallType string                   `json:"installType"`
	Status      string                   `json:"status"`
	InitiatedBy string                   `json:"initiatedBy"`
	Devices     []UpdatePushDetailDevice `json:"devices"`
	InitiatedAt int64                    `json:"initiatedAt"`
}

type UpdateCancelPushResult struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	CancelledBy string `json:"cancelledBy"`
	CancelledAt int64  `json:"cancelledAt"`
}

type UpdatePushRequest struct {
	ScheduledAt *int64   `json:"scheduledAt,omitempty"`
	Version     string   `json:"version"`
	InstallType string   `json:"installType"`
	DeviceIDs   []string `json:"deviceIds"`
}

type UpdateSyncStatusResult struct {
	Status        string `json:"status"`
	Error         string `json:"error,omitempty"`
	LastSyncAt    int64  `json:"lastSyncAt,omitempty"`
	NextSyncAt    int64  `json:"nextSyncAt,omitempty"`
	VersionsFound int    `json:"versionsFound,omitempty"`
}

type DeviceUpdateStatusRequest struct {
	DispatchID string `json:"dispatchId"`
	DeviceID   string `json:"deviceId"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
}

type DeviceUpdateStatusResponse struct {
	Message      string `json:"message"`
	Acknowledged bool   `json:"acknowledged"`
}

// ---- Command confirmation -------------------------------------------------.

type CommandConfirmRequest struct {
	Command string `json:"command"`
}

type CommandConfirmResult struct {
	ConfirmationToken    string `json:"confirmation_token,omitempty"`
	RiskTier             string `json:"risk_tier"`
	TraceID              string `json:"trace_id,omitempty"`
	ExpiresAt            int64  `json:"expires_at,omitempty"`
	ConfirmationRequired bool   `json:"confirmation_required"`
}

// ---- Dashboard ------------------------------------------------------------.

type DashboardStats struct {
	TotalDevices   int `json:"total_devices"`
	OnlineDevices  int `json:"online_devices"`
	OfflineDevices int `json:"offline_devices"`
	PendingDevices int `json:"pending_devices,omitempty"`
}

// ---- Diagnostics ----------------------------------------------------------.

type DeviceInspection struct {
	Battery       *float64 `json:"battery,omitempty"`
	DeviceID      string   `json:"device_id"`
	AppVersion    string   `json:"app_version"`
	LastSeen      int64    `json:"last_seen"`
	Online        bool     `json:"online"`
	FCMTokenValid bool     `json:"fcm_token_valid"`
}

type DiagnosticsConnectionInfo struct {
	WebSocketStatus string `json:"webSocketStatus"`
	FCMStatus       string `json:"fcmStatus"`
	ClientIP        string `json:"clientIp,omitempty"`
	Protocol        string `json:"protocol"`
	ConnectedAt     int64  `json:"connectedAt,omitempty"`
	LastSeen        int64  `json:"lastSeen,omitempty"`
}

type DiagnosticsIdentityInfo struct {
	IMEI         string `json:"imei"`
	DeviceName   string `json:"deviceName,omitempty"`
	Model        string `json:"model,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
}

type DiagnosticsSoftwareInfo struct {
	OSVersion     string `json:"osVersion"`
	AppVersion    string `json:"appVersion"`
	SecurityPatch string `json:"securityPatch,omitempty"`
	BuildID       string `json:"buildId,omitempty"`
}

type DiagnosticsRegistrationInfo struct {
	Status              string `json:"status"`
	RegisteredAt        int64  `json:"registeredAt,omitempty"`
	FCMTokenRefreshedAt int64  `json:"fcmTokenRefreshedAt,omitempty"`
	FCMTokenValid       bool   `json:"fcmTokenValid"`
	CommandSecretSet    bool   `json:"commandSecretSet"`
}

type DiagnosticsTelemetryInfo struct {
	LastTimestamp   int64 `json:"lastTimestamp"`
	FramesToday     int   `json:"framesToday"`
	AvgLatencyMs    int   `json:"avgLatencyMs,omitempty"`
	TotalBytesToday int64 `json:"totalBytesToday"`
	SessionsToday   int   `json:"sessionsToday"`
}

type DeviceInspectionResult struct {
	Connection   DiagnosticsConnectionInfo   `json:"connection"`
	Identity     DiagnosticsIdentityInfo     `json:"identity"`
	Software     DiagnosticsSoftwareInfo     `json:"software"`
	Registration DiagnosticsRegistrationInfo `json:"registration"`
	Telemetry    DiagnosticsTelemetryInfo    `json:"telemetry"`
}

// ---- Usage stats / admin --------------------------------------------------.

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
	UsageStats      *UsageStatsSnapshot `json:"usage_stats,omitempty"`
	LatestVersion   string              `json:"latest_version,omitempty"`
	CurrentVersion  string              `json:"current_version"`
	ReleaseName     string              `json:"release_name,omitempty"`
	ReleaseURL      string              `json:"release_url,omitempty"`
	UpdateAvailable bool                `json:"update_available"`
}

type TimelineEventResult struct {
	Data      map[string]any `json:"data,omitempty"`
	ID        string         `json:"id"`
	DeviceID  string         `json:"deviceId"`
	Type      string         `json:"type"`
	Timestamp string         `json:"timestamp"`
}

type TimelineResult struct {
	NextCursor string                `json:"nextCursor,omitempty"`
	Events     []TimelineEventResult `json:"events"`
	HasMore    bool                  `json:"hasMore"`
}

type WebhookTestRequest struct {
	URL string `json:"url"`
}

type MetricAggregateResult struct {
	Avg float64 `json:"avg"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type TelemetryStatsResult struct {
	DeviceID    string                `json:"deviceId"`
	LatestEntry string                `json:"latestEntry"`
	OldestEntry string                `json:"oldestEntry"`
	RiskScore   MetricAggregateResult `json:"riskScore"`
	BufferLevel MetricAggregateResult `json:"bufferLevel"`
	ThermalTemp MetricAggregateResult `json:"thermalTemp"`
	SampleCount int                   `json:"sampleCount"`
}

type UpdateOrganizationRequest struct {
	Name       *string `json:"name,omitempty"`
	MaxMembers *int    `json:"maxMembers,omitempty"`
	IsActive   *bool   `json:"isActive,omitempty"`
}

type TelemetryHistoryEntry struct {
	ReceivedAt  time.Time `json:"receivedAt"`
	ID          string    `json:"id"`
	DeviceID    string    `json:"deviceId"`
	Payload     string    `json:"payload,omitempty"`
	RiskScore   int       `json:"riskScore,omitempty"`
	BufferLevel int       `json:"bufferLevel,omitempty"`
	ThermalTemp float64   `json:"thermalTemp,omitempty"`
}

type TelemetryHistoryQueryResult struct {
	DeviceID   string                  `json:"deviceId"`
	Entries    []TelemetryHistoryEntry `json:"entries"`
	TotalCount int                     `json:"totalCount"`
	StartTime  int64                   `json:"startTime"`
	EndTime    int64                   `json:"endTime"`
	QueryTime  int64                   `json:"queryTime"`
}

// AdminAPIKey extends APIKey with operator identity for the super-admin
// "list all keys" view (domain.AdminAPIKeyResponse).
type AdminAPIKey struct {
	OperatorID   string `json:"operator_id,omitempty"`
	OperatorName string `json:"operator_name,omitempty"`
	APIKey
}

type AdminAPIKeyListResult struct {
	Keys       []AdminAPIKey `json:"keys"`
	Pagination Pagination    `json:"pagination"`
}

// DeviceLogEvent mirrors application/logs.LogEvent (event-based device logs).
type DeviceLogEvent struct {
	Data      map[string]any `json:"data,omitempty"`
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Timestamp int64          `json:"timestamp"`
}

type CursorPaginationResult struct {
	NextCursor string `json:"nextCursor,omitempty"`
	Limit      int    `json:"limit"`
	HasMore    bool   `json:"hasMore"`
}

// DeviceLogEventListResult mirrors application/logs.ListLogsResponse.
type DeviceLogEventListResult struct {
	Events     []DeviceLogEvent       `json:"events"`
	Pagination CursorPaginationResult `json:"pagination"`
}
