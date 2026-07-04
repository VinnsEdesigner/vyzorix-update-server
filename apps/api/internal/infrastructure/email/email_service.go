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

// Service handles sending emails via Resend API.
type Service struct {
	client    *http.Client
	apiKey    string
	fromEmail string
	fromName  string
	baseURL   string
}

// NewService creates a new email service instance.
func NewService() *Service {
	return &Service{
		apiKey:    os.Getenv("RESEND_API_KEY"),
		fromEmail: os.Getenv("EMAIL_FROM"),
		fromName:  os.Getenv("EMAIL_FROM_NAME"),
		baseURL:   os.Getenv("BASE_URL"),
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Data contains data for email template rendering.
type Data struct {
	Name         string
	VerifyURL    string
	ResetURL     string
	TokenPreview string
	ExpiryHours  int
	ExpiryMins   int
}

// SendVerificationEmail sends a welcome email with email verification link.
func (s *Service) SendVerificationEmail(ctx context.Context, to, name, token string) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	verifyURL := fmt.Sprintf("%s/auth/waitVerify?token=%s&type=verify", s.baseURL, token)

	html, err := s.parseTemplate(verificationEmailTemplate, Data{
		Name:        name,
		VerifyURL:   verifyURL,
		ExpiryHours: 24,
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

	html, err := s.parseTemplate(passwordResetEmailTemplate, Data{
		Name:       name,
		ResetURL:   resetURL,
		ExpiryMins: 60,
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

	html, err := s.parseTemplate(passwordChangedEmailTemplate, Data{
		Name: name,
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

	html, err := s.parseTemplate(thresholdBreachEmailTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to parse threshold breach template: %w", err)
	}

	subject := fmt.Sprintf("⚠️ Alert: %s threshold breached on device %s", data.AlertType, data.DeviceID)
	return s.send(ctx, to, subject, html)
}

// SendDeviceOfflineEmail sends a device offline notification email.
func (s *Service) SendDeviceOfflineEmail(ctx context.Context, to string, data NotificationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(deviceOfflineEmailTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to parse device offline template: %w", err)
	}

	subject := fmt.Sprintf("🔴 Device Offline: %s", data.DeviceID)
	return s.send(ctx, to, subject, html)
}

// SendDeviceOnlineEmail sends a device online notification email.
func (s *Service) SendDeviceOnlineEmail(ctx context.Context, to string, data NotificationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(deviceOnlineEmailTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to parse device online template: %w", err)
	}

	subject := fmt.Sprintf("🟢 Device Online: %s", data.DeviceID)
	return s.send(ctx, to, subject, html)
}

// SendCommandFailedEmail sends a command failed notification email.
func (s *Service) SendCommandFailedEmail(ctx context.Context, to string, data NotificationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(commandFailedEmailTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to parse command failed template: %w", err)
	}

	subject := fmt.Sprintf("❌ Command Failed: %s on device %s", data.CommandName, data.DeviceID)
	return s.send(ctx, to, subject, html)
}

// SendUpdateAvailableEmail sends an update available notification email.
func (s *Service) SendUpdateAvailableEmail(ctx context.Context, to string, data NotificationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(updateAvailableEmailTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to parse update available template: %w", err)
	}

	subject := fmt.Sprintf("📦 Update Available: Version %s for device %s", data.UpdateVersion, data.DeviceID)
	return s.send(ctx, to, subject, html)
}

// SendRegistrationRequestEmail sends a device registration request notification.
func (s *Service) SendRegistrationRequestEmail(ctx context.Context, to string, data NotificationData) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(registrationRequestEmailTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to parse registration request template: %w", err)
	}

	subject := fmt.Sprintf("📋 Registration Request: %s", data.RequesterName)
	return s.send(ctx, to, subject, html)
}

// SendErrorAlertEmail sends an error alert email.
func (s *Service) SendErrorAlertEmail(ctx context.Context, to string, data NotificationData) error {
        if s.apiKey == "" {
                return errors.New("RESEND_API_KEY not configured")
        }

        html, err := s.parseTemplate(errorAlertEmailTemplate, data)
        if err != nil {
                return fmt.Errorf("failed to parse error alert template: %w", err)
        }

        subject := fmt.Sprintf("🔴 Error Alert: %s", data.DeviceID)
        return s.send(ctx, to, subject, html)
}

// errorAlertEmailTemplate is the HTML template for error alerts.
const errorAlertEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Error Alert</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; margin: 0; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { background: #dc3545; color: white; padding: 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; }
        .content { padding: 20px; }
        .alert-box { background: #f8d7da; border: 1px solid #f5c6cb; border-radius: 4px; padding: 15px; margin: 15px 0; }
        .alert-title { font-weight: bold; color: #721c24; margin-bottom: 10px; }
        .alert-value { font-family: monospace; background: white; padding: 8px; border-radius: 4px; margin: 5px 0; }
        .footer { background: #f8f9fa; padding: 15px; text-align: center; font-size: 12px; color: #6c757d; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔴 Error Alert</h1>
        </div>
        <div class="content">
            <p>Hello {{.OperatorName}},</p>
            <p>An error event has been detected on your device. Please investigate immediately.</p>
            <div class="alert-box">
                <div class="alert-title">Error Details:</div>
                <div class="alert-value"><strong>Device ID:</strong> {{.DeviceID}}</div>
                <div class="alert-value"><strong>Device Name:</strong> {{.DeviceName}}</div>
                <div class="alert-value"><strong>Error Message:</strong> {{.ErrorMessage}}</div>
                <div class="alert-value"><strong>Time:</strong> {{.Timestamp}}</div>
            </div>
            <p>Please check your device and take appropriate action.</p>
        </div>
        <div class="footer">
            <p>Vyzorix Update Server</p>
            <p>This is an automated alert. Do not reply.</p>
        </div>
    </div>
</body>
</html>`

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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

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

// verificationEmailTemplate is the HTML template for email verification.
const verificationEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify your email</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 24px; }
        .button-wrapper { text-align: center; margin: 32px 0; }
        .button { display: inline-block; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 600; padding: 16px 32px; border-radius: 8px; transition: transform 0.2s; }
        .button:hover { transform: translateY(-2px); }
        .expiry { text-align: center; font-size: 14px; color: #888888; margin-top: 24px; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
        .footer a { color: #667eea; text-decoration: none; }
        .ignore { background: #f0f0f0; padding: 20px; border-radius: 8px; margin-top: 24px; font-size: 14px; color: #666666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="logo">Vyzorix</div>
            </div>
            <div class="content">
                <h1>Verify your email address</h1>
                <p>Hi {{.Name}},</p>
                <p>Thanks for signing up! Please verify your email address by clicking the button below. This helps us keep your account secure.</p>
                <div class="button-wrapper">
                    <a href="{{.VerifyURL}}" class="button">Verify Email Address</a>
                </div>
                <p class="expiry">This link expires in {{.ExpiryHours}} hours</p>
                <div class="ignore">
                    <p>If you didn't create an account with Vyzorix, you can safely ignore this email. Someone may have entered your email address by mistake.</p>
                </div>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p>This email was sent automatically. Please do not reply.</p>
            </div>
        </div>
    </div>
</body>
</html>`

// passwordResetEmailTemplate is the HTML template for password reset.
const passwordResetEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset your password</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 24px; }
        .button-wrapper { text-align: center; margin: 32px 0; }
        .button { display: inline-block; background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%); color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 600; padding: 16px 32px; border-radius: 8px; transition: transform 0.2s; }
        .button:hover { transform: translateY(-2px); }
        .warning { background: #fff3cd; border: 1px solid #ffeeba; color: #856404; padding: 16px 20px; border-radius: 8px; margin: 24px 0; font-size: 14px; }
        .warning strong { color: #856404; }
        .expiry { text-align: center; font-size: 14px; color: #888888; margin-top: 24px; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="logo">Vyzorix</div>
            </div>
            <div class="content">
                <h1>Reset your password</h1>
                <p>Hi {{.Name}},</p>
                <p>We received a request to reset your password. Click the button below to create a new password for your account.</p>
                <div class="button-wrapper">
                    <a href="{{.ResetURL}}" class="button">Reset Password</a>
                </div>
                <div class="warning">
                    <strong> Security Notice:</strong> This link expires in {{.ExpiryMins}} minutes and can only be used once. If you didn't request a password reset, please ignore this email.
                </div>
                <p class="expiry">Link expires in {{.ExpiryMins}} minutes</p>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p>If you need help, contact support.</p>
            </div>
        </div>
    </div>
</body>
</html>`

// passwordChangedEmailTemplate is the HTML template for password change confirmation.
const passwordChangedEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password changed</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #27ae60 0%, #2ecc71 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 24px; }
        .checkmark { text-align: center; font-size: 64px; margin: 24px 0; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="logo">Vyzorix</div>
            </div>
            <div class="content">
                <div class="checkmark"></div>
                <h1>Password changed successfully</h1>
                <p>Hi {{.Name}},</p>
                <p>Your password has been changed successfully. If you did not make this change, please contact support immediately.</p>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
            </div>
        </div>
    </div>
</body>
</html>`
// thresholdBreachEmailTemplate is the HTML template for threshold breach alerts.
const thresholdBreachEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Threshold Alert</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .alert-icon { font-size: 48px; margin-bottom: 16px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 16px; }
        .alert-box { background: #fff3cd; border: 1px solid #ffc107; border-radius: 8px; padding: 20px; margin: 24px 0; }
        .alert-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #eee; }
        .alert-row:last-child { border-bottom: none; }
        .alert-label { font-weight: 600; color: #333; }
        .alert-value { color: #666; }
        .severity-critical { background: #f8d7da; border-color: #f5c6cb; }
        .severity-warning { background: #fff3cd; border-color: #ffc107; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="alert-icon">⚠️</div>
                <div class="logo">Threshold Alert</div>
            </div>
            <div class="content">
                <h1>{{.AlertType}} Threshold Exceeded</h1>
                <p>Hi {{.OperatorName}},</p>
                <p>A metric on your device has exceeded its threshold and requires attention.</p>
                <div class="alert-box">
                    <div class="alert-row">
                        <span class="alert-label">Device ID:</span>
                        <span class="alert-value">{{.DeviceID}}</span>
                    </div>
                    <div class="alert-row">
                        <span class="alert-label">Alert Type:</span>
                        <span class="alert-value">{{.AlertType}}</span>
                    </div>
                    <div class="alert-row">
                        <span class="alert-label">Current Value:</span>
                        <span class="alert-value">{{.CurrentValue}}</span>
                    </div>
                    <div class="alert-row">
                        <span class="alert-label">Threshold:</span>
                        <span class="alert-value">{{.Threshold}}</span>
                    </div>
                    <div class="alert-row">
                        <span class="alert-label">Time:</span>
                        <span class="alert-value">{{.Timestamp}}</span>
                    </div>
                </div>
                <p>Please check your device and take appropriate action.</p>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p>This is an automated alert. Do not reply.</p>
            </div>
        </div>
    </div>
</body>
</html>`

// deviceOfflineEmailTemplate is the HTML template for device offline notifications.
const deviceOfflineEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Device Offline</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #6c757d 0%, #495057 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .icon { font-size: 48px; margin-bottom: 16px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 16px; }
        .info-box { background: #f8f9fa; border-radius: 8px; padding: 20px; margin: 24px 0; }
        .info-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #eee; }
        .info-row:last-child { border-bottom: none; }
        .info-label { font-weight: 600; color: #333; }
        .info-value { color: #666; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="icon">🔴</div>
                <div class="logo">Device Offline</div>
            </div>
            <div class="content">
                <h1>Device Disconnected</h1>
                <p>Hi {{.OperatorName}},</p>
                <p>One of your devices has gone offline and is no longer connected to the server.</p>
                <div class="info-box">
                    <div class="info-row">
                        <span class="info-label">Device ID:</span>
                        <span class="info-value">{{.DeviceID}}</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Disconnected At:</span>
                        <span class="info-value">{{.Timestamp}}</span>
                    </div>
                </div>
                <p>The device will automatically reconnect when it comes back online. You will receive a notification when it reconnects.</p>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p>This is an automated notification. Do not reply.</p>
            </div>
        </div>
    </div>
</body>
</html>`

// deviceOnlineEmailTemplate is the HTML template for device online notifications.
const deviceOnlineEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Device Online</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #28a745 0%, #20c997 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .icon { font-size: 48px; margin-bottom: 16px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 16px; }
        .info-box { background: #f8f9fa; border-radius: 8px; padding: 20px; margin: 24px 0; }
        .info-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #eee; }
        .info-row:last-child { border-bottom: none; }
        .info-label { font-weight: 600; color: #333; }
        .info-value { color: #666; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="icon">🟢</div>
                <div class="logo">Device Online</div>
            </div>
            <div class="content">
                <h1>Device Reconnected</h1>
                <p>Hi {{.OperatorName}},</p>
                <p>Your device has successfully reconnected to the server.</p>
                <div class="info-box">
                    <div class="info-row">
                        <span class="info-label">Device ID:</span>
                        <span class="info-value">{{.DeviceID}}</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Connected At:</span>
                        <span class="info-value">{{.Timestamp}}</span>
                    </div>
                </div>
                <p>All systems are operational. You can now send commands to your device.</p>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p>This is an automated notification. Do not reply.</p>
            </div>
        </div>
    </div>
</body>
</html>`

// commandFailedEmailTemplate is the HTML template for command failed notifications.
const commandFailedEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Command Failed</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #dc3545 0%, #bd2139 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .icon { font-size: 48px; margin-bottom: 16px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 16px; }
        .error-box { background: #f8d7da; border: 1px solid #f5c6cb; border-radius: 8px; padding: 20px; margin: 24px 0; }
        .error-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid rgba(0,0,0,0.1); }
        .error-row:last-child { border-bottom: none; }
        .error-label { font-weight: 600; color: #333; }
        .error-value { color: #666; }
        .failure-reason { background: #721c24; color: #f8d7da; padding: 12px; border-radius: 4px; margin-top: 12px; font-family: monospace; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="icon">❌</div>
                <div class="logo">Command Failed</div>
            </div>
            <div class="content">
                <h1>Command Execution Failed</h1>
                <p>Hi {{.OperatorName}},</p>
                <p>A command you sent to your device has failed to execute.</p>
                <div class="error-box">
                    <div class="error-row">
                        <span class="error-label">Device ID:</span>
                        <span class="error-value">{{.DeviceID}}</span>
                    </div>
                    <div class="error-row">
                        <span class="error-label">Command:</span>
                        <span class="error-value">{{.CommandName}}</span>
                    </div>
                    <div class="error-row">
                        <span class="error-label">Time:</span>
                        <span class="error-value">{{.Timestamp}}</span>
                    </div>
                    <div class="failure-reason">Reason: {{.FailureReason}}</div>
                </div>
                <p>Please check your device status and retry the command if necessary.</p>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p>This is an automated notification. Do not reply.</p>
            </div>
        </div>
    </div>
</body>
</html>`

// updateAvailableEmailTemplate is the HTML template for update available notifications.
const updateAvailableEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Update Available</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #007bff 0%, #0056b3 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .icon { font-size: 48px; margin-bottom: 16px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 16px; }
        .update-box { background: #e7f3ff; border: 1px solid #b3d7ff; border-radius: 8px; padding: 20px; margin: 24px 0; }
        .update-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #b3d7ff; }
        .update-row:last-child { border-bottom: none; }
        .update-label { font-weight: 600; color: #333; }
        .update-value { color: #0066cc; font-weight: 500; }
        .button-wrapper { text-align: center; margin: 32px 0; }
        .button { display: inline-block; background: linear-gradient(135deg, #007bff 0%, #0056b3 100%); color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 600; padding: 16px 32px; border-radius: 8px; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="icon">📦</div>
                <div class="logo">Update Available</div>
            </div>
            <div class="content">
                <h1>New Update Ready</h1>
                <p>Hi {{.OperatorName}},</p>
                <p>A software update is available for your device.</p>
                <div class="update-box">
                    <div class="update-row">
                        <span class="update-label">Device ID:</span>
                        <span class="update-value">{{.DeviceID}}</span>
                    </div>
                    <div class="update-row">
                        <span class="update-label">New Version:</span>
                        <span class="update-value">{{.UpdateVersion}}</span>
                    </div>
                </div>
                <div class="button-wrapper">
                    <a href="{{.BaseURL}}/devices/{{.DeviceID}}/updates" class="button">View Update</a>
                </div>
                <p>You can deploy this update from your dashboard at any time.</p>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p>This is an automated notification. Do not reply.</p>
            </div>
        </div>
    </div>
</body>
</html>`

// registrationRequestEmailTemplate is the HTML template for registration request notifications.
const registrationRequestEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Registration Request</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #6f42c1 0%, #5a3d8a 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .icon { font-size: 48px; margin-bottom: 16px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 16px; }
        .request-box { background: #f8f0ff; border: 1px solid #d4b8e8; border-radius: 8px; padding: 20px; margin: 24px 0; }
        .request-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #d4b8e8; }
        .request-row:last-child { border-bottom: none; }
        .request-label { font-weight: 600; color: #333; }
        .request-value { color: #6f42c1; }
        .button-wrapper { text-align: center; margin: 32px 0; }
        .button { display: inline-block; background: linear-gradient(135deg, #28a745 0%, #20c997 100%); color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 600; padding: 16px 32px; border-radius: 8px; margin-right: 12px; }
        .button-secondary { background: linear-gradient(135deg, #6c757d 0%, #495057 100%); }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="icon">📋</div>
                <div class="logo">Registration Request</div>
            </div>
            <div class="content">
                <h1>New Device Registration</h1>
                <p>Hi {{.OperatorName}},</p>
                <p>A new device is requesting to be registered to your account.</p>
                <div class="request-box">
                    <div class="request-row">
                        <span class="request-label">Requester Name:</span>
                        <span class="request-value">{{.RequesterName}}</span>
                    </div>
                    <div class="request-row">
                        <span class="request-label">Device ID:</span>
                        <span class="request-value">{{.DeviceID}}</span>
                    </div>
                    <div class="request-row">
                        <span class="request-label">Requested At:</span>
                        <span class="request-value">{{.Timestamp}}</span>
                    </div>
                </div>
                <div class="button-wrapper">
                    <a href="{{.BaseURL}}/registrations/pending" class="button">Review Request</a>
                    <a href="{{.BaseURL}}/registrations" class="button button-secondary">View All</a>
                </div>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p>This is an automated notification. Do not reply.</p>
            </div>
        </div>
    </div>
</body>
</html>`
