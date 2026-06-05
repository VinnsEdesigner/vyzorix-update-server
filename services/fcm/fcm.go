package fcm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var ErrDisabled = errors.New("fcm notifier disabled: FIREBASE_CREDENTIALS is empty")

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

func Init(log *slog.Logger, rawCredentials string) (*Client, error) {
	c := &Client{log: log}
	if rawCredentials == "" {
		log.Warn("fcm disabled; FIREBASE_CREDENTIALS not configured")
		return c, nil
	}
	creds := option.WithCredentialsJSON([]byte(rawCredentials))
	app, err := firebase.NewApp(context.Background(), nil, creds)
	if err != nil {
		return nil, fmt.Errorf("firebase init: %w", err)
	}
	c.app = app
	c.enabled = true
	c.log.Info("fcm initialized", "app", app.ProjectID)
	return c, nil
}

func (c *Client) Enabled() bool { return c != nil && c.enabled }

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
