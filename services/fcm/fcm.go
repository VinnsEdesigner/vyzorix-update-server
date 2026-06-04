package fcm

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
)

var ErrDisabled = errors.New("fcm notifier disabled: FIREBASE_CREDENTIALS is empty")

type Client struct {
	log       *slog.Logger
	enabled   bool
	projectID string
}

type serviceAccount struct {
	ProjectID   string `json:"project_id"`
	ClientEmail string `json:"client_email"`
}

func Init(log *slog.Logger, rawCredentials string) (*Client, error) {
	c := &Client{log: log}
	if strings.TrimSpace(rawCredentials) == "" {
		log.Warn("fcm disabled; FIREBASE_CREDENTIALS not configured")
		return c, nil
	}
	var sa serviceAccount
	if err := json.Unmarshal([]byte(rawCredentials), &sa); err != nil {
		return nil, err
	}
	c.enabled = true
	c.projectID = sa.ProjectID
	log.Info("fcm credentials loaded", "projectId", sa.ProjectID, "clientEmail", sa.ClientEmail)
	return c, nil
}

func (c *Client) Enabled() bool { return c != nil && c.enabled }
