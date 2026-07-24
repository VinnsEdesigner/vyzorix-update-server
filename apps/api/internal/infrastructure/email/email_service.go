// Package email provides email infrastructure using the Resend API.
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email/templates"
)

// NotificationData contains data for notification email templates.
type NotificationData struct {
	OperatorName  string
	DeviceID      string
	DeviceName    string
	EventType     string
	AlertType     string
	CurrentValue  string
	Threshold     string
	CommandName   string
	FailureReason string
	UpdateVersion string
	RequesterName string
	ErrorMessage  string
	Timestamp     string
	BaseURL       string
}

// InvitationData contains data for invitation email templates.
type InvitationData struct {
	InviteeName      string
	InviterName      string
	OrganizationName string
	Role             string
	InviterNotes     string
	InviteeNotes     string
	AcceptURL        string
	AcceptedAt       string
	BaseURL          string
	ExpiryDays       int
}

// Service handles sending emails via Resend API.
type Service struct {
	client    *http.Client
	apiKey    string
	fromEmail string
	fromName  string
	baseURL   string
	apiURL    string
}

// NewService creates a new email service instance.
func NewService() *Service {
	apiURL := os.Getenv("RESEND_API_URL")
	if apiURL == "" {
		apiURL = "https://api.resend.com"
	}
	return &Service{
		apiKey:    os.Getenv("RESEND_API_KEY"),
		fromEmail: os.Getenv("EMAIL_FROM"),
		fromName:  os.Getenv("EMAIL_FROM_NAME"),
		baseURL:   os.Getenv("BASE_URL"),
		client:    &http.Client{Timeout: 10 * time.Second},
		apiURL:    apiURL,
	}
}

// Data contains data for email template rendering.
type Data struct {
	Name         string
	VerifyURL    string
	ResetURL     string
	BaseURL      string
	TokenPreview string
	ExpiryHours  int
	ExpiryMins   int
}

// LoginNotificationData contains data for new login notification emails.
// 10: Added for login notification feature.
type LoginNotificationData struct {
	OperatorName string
	IPAddress    string
	UserAgent    string
	Location     string
	Device       string
	Timestamp    string
	BaseURL      string
}

// SendVerificationEmail sends a welcome email with email verification link.
func (s *Service) SendVerificationEmail(ctx context.Context, to, name, token string) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	verifyURL := fmt.Sprintf("%s/auth/waitVerify?token=%s&type=verify", s.baseURL, token)

	html, err := s.parseTemplate(templates.VerificationEmail, Data{
		Name:        name,
		VerifyURL:   verifyURL,
		ExpiryHours: 24,
		BaseURL:     s.baseURL,
	})
	if err != nil {
		return fmt.Errorf("failed to parse verification template: %w", err)
	}

	return s.send(ctx, to, "Verify your Vyzorix account", html)
}

// SendPasswordResetEmail sends a password reset email.
func (s *Service) SendPasswordResetEmail(ctx context.Context, to, name, token string) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	resetURL := fmt.Sprintf("%s/auth/waitVerify?token=%s&type=reset", s.baseURL, token)

	html, err := s.parseTemplate(templates.PasswordResetEmail, Data{
		Name:       name,
		ResetURL:   resetURL,
		ExpiryMins: 60,
		BaseURL:    s.baseURL,
	})
	if err != nil {
		return fmt.Errorf("failed to parse reset template: %w", err)
	}

	return s.send(ctx, to, "Reset your Vyzorix password", html)
}

// SendPasswordChangedEmail sends a confirmation when password is changed.
func (s *Service) SendPasswordChangedEmail(ctx context.Context, to, name string) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(templates.PasswordChangedEmail, Data{
		Name:    name,
		BaseURL: s.baseURL,
	})
	if err != nil {
		return fmt.Errorf("failed to parse password changed template: %w", err)
	}

	return s.send(ctx, to, "Your password was changed", html)
}

// SendThresholdBreachEmail sends a threshold breach alert email.
func (s *Service) SendThresholdBreachEmail(ctx context.Context, to string, data NotificationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(templates.ThresholdBreachEmail, data)
	if err != nil {
		return fmt.Errorf("failed to parse threshold breach template: %w", err)
	}

	subject := fmt.Sprintf(" Alert: %s threshold breached on device %s", data.AlertType, data.DeviceID)
	return s.send(ctx, to, subject, html)
}

// SendDeviceOfflineEmail sends a device offline notification email.
func (s *Service) SendDeviceOfflineEmail(ctx context.Context, to string, data NotificationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(templates.DeviceOfflineEmail, data)
	if err != nil {
		return fmt.Errorf("failed to parse device offline template: %w", err)
	}

	subject := fmt.Sprintf(" Device Offline: %s", data.DeviceID)
	return s.send(ctx, to, subject, html)
}

// SendDeviceOnlineEmail sends a device online notification email.
func (s *Service) SendDeviceOnlineEmail(ctx context.Context, to string, data NotificationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(templates.DeviceOnlineEmail, data)
	if err != nil {
		return fmt.Errorf("failed to parse device online template: %w", err)
	}

	subject := fmt.Sprintf(" Device Online: %s", data.DeviceID)
	return s.send(ctx, to, subject, html)
}

// SendCommandFailedEmail sends a command failed notification email.
func (s *Service) SendCommandFailedEmail(ctx context.Context, to string, data NotificationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(templates.CommandFailedEmail, data)
	if err != nil {
		return fmt.Errorf("failed to parse command failed template: %w", err)
	}

	subject := fmt.Sprintf(" Command Failed: %s on device %s", data.CommandName, data.DeviceID)
	return s.send(ctx, to, subject, html)
}

// SendUpdateAvailableEmail sends an update available notification email.
func (s *Service) SendUpdateAvailableEmail(ctx context.Context, to string, data NotificationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(templates.UpdateAvailableEmail, data)
	if err != nil {
		return fmt.Errorf("failed to parse update available template: %w", err)
	}

	subject := fmt.Sprintf(" Update Available: Version %s for device %s", data.UpdateVersion, data.DeviceID)
	return s.send(ctx, to, subject, html)
}

// SendRegistrationRequestEmail sends a device registration request notification.
func (s *Service) SendRegistrationRequestEmail(ctx context.Context, to string, data NotificationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(templates.RegistrationRequestEmail, data)
	if err != nil {
		return fmt.Errorf("failed to parse registration request template: %w", err)
	}

	subject := fmt.Sprintf(" Registration Request: %s", data.RequesterName)
	return s.send(ctx, to, subject, html)
}

// SendErrorAlertEmail sends an error alert email.
func (s *Service) SendErrorAlertEmail(ctx context.Context, to string, data NotificationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(templates.ErrorAlertEmail, data)
	if err != nil {
		return fmt.Errorf("failed to parse error alert template: %w", err)
	}

	subject := fmt.Sprintf(" Error Alert: %s", data.DeviceID)
	return s.send(ctx, to, subject, html)
}

// SendNewLoginNotificationEmail sends a new login notification email to the operator.
// 10: Added for login notification feature.
func (s *Service) SendNewLoginNotificationEmail(ctx context.Context, to string, data LoginNotificationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	// Ensure BaseURL is set for the template.
	if data.BaseURL == "" {
		data.BaseURL = s.baseURL
	}

	html, err := s.parseTemplate(templates.NewLoginEmail, data)
	if err != nil {
		return fmt.Errorf("failed to parse new login template: %w", err)
	}

	subject := " New login to your account"
	return s.send(ctx, to, subject, html)
}

// SendInvitationEmail sends an organization invitation email to the invitee.
func (s *Service) SendInvitationEmail(ctx context.Context, to string, data InvitationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(templates.InvitationEmail, data)
	if err != nil {
		return fmt.Errorf("failed to parse invitation template: %w", err)
	}

	subject := fmt.Sprintf("You've been invited to join %s", data.OrganizationName)
	return s.send(ctx, to, subject, html)
}

// SendInvitationAcceptedEmail sends a notification when an invitation is accepted.
func (s *Service) SendInvitationAcceptedEmail(ctx context.Context, to string, data InvitationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(templates.InvitationAcceptedEmail, data)
	if err != nil {
		return fmt.Errorf("failed to parse invitation accepted template: %w", err)
	}

	subject := fmt.Sprintf("Invitation accepted - %s", data.OrganizationName)
	return s.send(ctx, to, subject, html)
}

// SendInvitationRejectedEmail sends a notification when an invitation is rejected.
func (s *Service) SendInvitationRejectedEmail(ctx context.Context, to string, data InvitationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(templates.InvitationRejectedEmail, data)
	if err != nil {
		return fmt.Errorf("failed to parse invitation rejected template: %w", err)
	}

	subject := fmt.Sprintf("Invitation declined - %s", data.OrganizationName)
	return s.send(ctx, to, subject, html)
}

// send sends an email via the Resend API.
func (s *Service) send(ctx context.Context, to, subject, html string) error {
	if s.fromEmail == "" {
		s.fromEmail = "noreply@vyzorix.app"
	}

	if s.fromName == "" {
		s.fromName = "Vyzorix"
	}

	payload := map[string]any{
		"from":    fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		"to":      []string{to},
		"subject": subject,
		"html":    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL+"/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API returned status %d", resp.StatusCode)
	}

	return nil
}

// parseTemplate renders an email template with the given data.
func (s *Service) parseTemplate(tmpl string, data any) (string, error) {
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// IsConfigured returns true if the email service is properly configured.
func (s *Service) IsConfigured() bool {
	return s.apiKey != "" && s.fromEmail != ""
}
