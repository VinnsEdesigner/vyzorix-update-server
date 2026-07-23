// Package fcm provides Firebase Cloud Messaging integration.
package fcm

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var (
	ErrDisabled          = errors.New("fcm notifier disabled: FIREBASE_CREDENTIALS is empty")
	ErrFCMCircuitOpen   = errors.New("fcm circuit breaker is open")
)

type Client struct {
	log      *slog.Logger
	app      *firebase.App
	projects string
	enabled  bool
}

type serviceAccount struct {
	ProjectID   string `json:"project_id"`
	ClientEmail string `json:"client_email"`
}

// Init initializes the FCM client with graceful degradation for malformed credentials.

// is returned in a disabled state with a warning logged.
func Init(log *slog.Logger, rawCredentials string) (*Client, error) {
	c := &Client{log: log}

	if rawCredentials == "" {
		log.Warn("fcm disabled; FIREBASE_CREDENTIALS not configured")
		return c, nil
	}

	// Write credentials to a temporary file and use WithAuthCredentialsFile
	// This is the recommended approach to avoid the deprecated WithCredentialsJSON
	// by explicitly specifying the credential type as ServiceAccount.
	tmpFile, err := os.CreateTemp("", "fcm-credentials-*.json")
	if err != nil {
		log.Warn("fcm disabled; failed to create temp credentials file", "error", err.Error())
		return c, nil
	}
	tmpPath := tmpFile.Name()

	// Write credentials and close file before using it
	if _, err := tmpFile.WriteString(rawCredentials); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		log.Warn("fcm disabled; failed to write credentials", "error", err.Error())
		return c, nil
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		log.Warn("fcm disabled; failed to close temp credentials file", "error", err.Error())
		return c, nil
	}

	// Use WithAuthCredentialsFile with explicit ServiceAccount type
	// This validates that the credentials are actually a service account
	creds := option.WithAuthCredentialsFile(option.ServiceAccount, tmpPath)

	app, err := firebase.NewApp(context.Background(), nil, creds)
	if err != nil {
		_ = os.Remove(tmpPath)
		
		log.Warn("fcm disabled; malformed FIREBASE_CREDENTIALS - Firebase init failed",
			"error", err.Error(),
			"hint", "Ensure credentials are valid JSON service account file")
		return c, nil
	}

	// Clean up temp file after Firebase app is initialized
	_ = os.Remove(tmpPath)

	c.app = app
	c.enabled = true
	c.projects = getProjectID(rawCredentials)
	c.log.Info("fcm initialized", "project", c.projects)

	return c, nil
}

func (c *Client) Enabled() bool { return c != nil && c.enabled }

func (c *Client) ProjectID() string {
	if c == nil {
		return ""
	}

	return c.projects
}

func (c *Client) Messaging() *messaging.Client {
	if c == nil || c.app == nil {
		return nil
	}

	client, err := c.app.Messaging(context.Background())
	if err != nil {
		c.log.Error("fcm messaging client", "err", err)
		return nil
	}

	return client
}

func getProjectID(cred string) string {
	var sa serviceAccount
	if err := json.Unmarshal([]byte(cred), &sa); err == nil && sa.ProjectID != "" {
		return sa.ProjectID
	}

	return ""
}
