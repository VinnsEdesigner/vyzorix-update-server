// Package services provides business logic services.
package services

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

// EmailService handles sending emails via Resend API.
type EmailService struct {
	client    *http.Client
	apiKey    string
	fromEmail string
	fromName  string
	baseURL   string
}

// NewEmailService creates a new email service instance.
func NewEmailService() *EmailService {
	return &EmailService{
		apiKey:    os.Getenv("RESEND_API_KEY"),
		fromEmail: os.Getenv("EMAIL_FROM"),
		fromName:  os.Getenv("EMAIL_FROM_NAME"),
		baseURL:   os.Getenv("BASE_URL"),
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// EmailData contains data for email template rendering.
type EmailData struct {
	Name         string
	VerifyURL    string
	ResetURL     string
	TokenPreview string
	ExpiryHours  int
	ExpiryMins   int
}

// SendVerificationEmail sends a welcome email with email verification link.
func (s *EmailService) SendVerificationEmail(ctx context.Context, to, name, token string) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", s.baseURL, token)

	// Parse template
	html, err := s.parseTemplate(verificationEmailTemplate, EmailData{
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
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, to, name, token string) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", s.baseURL, token)

	// Parse template
	html, err := s.parseTemplate(passwordResetEmailTemplate, EmailData{
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
func (s *EmailService) SendPasswordChangedEmail(ctx context.Context, to, name string) error {
	if s.apiKey == "" {
		return errors.New("RESEND_API_KEY not configured")
	}

	html, err := s.parseTemplate(passwordChangedEmailTemplate, EmailData{
		Name: name,
	})
	if err != nil {
		return fmt.Errorf("failed to parse password changed template: %w", err)
	}

	return s.send(ctx, to, "Your password was changed", html)
}

// send sends an email via the Resend API.
func (s *EmailService) send(ctx context.Context, to, subject, html string) error {
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
func (s *EmailService) parseTemplate(tmpl string, data EmailData) (string, error) {
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
func (s *EmailService) IsConfigured() bool {
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
                    <strong>⚠️ Security Notice:</strong> This link expires in {{.ExpiryMins}} minutes and can only be used once. If you didn't request a password reset, please ignore this email.
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
                <div class="checkmark">✓</div>
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
