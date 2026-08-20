// Package notification provides contact points and delivery channels for
// alert and system notifications. A contact point is a named, org-scoped
// destination (email, webhook, or slack) that receives templated messages
// from alert rules and system events.
package notification

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrNotFound is returned when a contact point is not found.
var ErrNotFound = errors.New("contact point not found")

// ChannelType identifies the transport of a contact point.
type ChannelType string

const (
	ChannelTypeEmail   ChannelType = "email"
	ChannelTypeWebhook ChannelType = "webhook"
	ChannelTypeSlack   ChannelType = "slack"
)

// Valid reports whether the channel type is supported.
func (c ChannelType) Valid() bool {
	switch c {
	case ChannelTypeEmail, ChannelTypeWebhook, ChannelTypeSlack:
		return true
	}
	return false
}

// ContactPoint is an org-scoped notification destination.
type ContactPoint struct {
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ID         string
	OrgID      string
	Name       string
	Channel    ChannelType
	Secret     string
	Config     map[string]string
	TemplateID string
	Enabled    bool
}

// Validate checks the contact point is well-formed for its channel type.
func (cp *ContactPoint) Validate() error {
	if strings.TrimSpace(cp.OrgID) == "" {
		return errors.New("org_id is required")
	}
	if strings.TrimSpace(cp.Name) == "" {
		return errors.New("name is required")
	}
	if !cp.Channel.Valid() {
		return fmt.Errorf("invalid channel type %q", cp.Channel)
	}

	switch cp.Channel {
	case ChannelTypeEmail:
		to := strings.TrimSpace(cp.Config["to"])
		if to == "" {
			return errors.New("email contact point requires config key 'to'")
		}
		if !strings.Contains(to, "@") {
			return fmt.Errorf("invalid email address %q", to)
		}
	case ChannelTypeWebhook:
		url := strings.TrimSpace(cp.Config["url"])
		if url == "" {
			return errors.New("webhook contact point requires config key 'url'")
		}
		if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://localhost") {
			return fmt.Errorf("webhook url must be https or localhost: %q", url)
		}
	case ChannelTypeSlack:
		webhook := strings.TrimSpace(cp.Config["webhook"])
		if webhook == "" {
			return errors.New("slack contact point requires config key 'webhook'")
		}
		if !strings.HasPrefix(webhook, "https://hooks.slack.com/") {
			return fmt.Errorf("invalid slack webhook url: %q", webhook)
		}
	}
	return nil
}

// Repository persists contact points.
type Repository interface {
	Save(ctx context.Context, cp *ContactPoint) error
	GetByID(ctx context.Context, id string) (*ContactPoint, error)
	ListByOrg(ctx context.Context, orgID string) ([]*ContactPoint, error)
	ListEnabledByOrg(ctx context.Context, orgID string) ([]*ContactPoint, error)
	Delete(ctx context.Context, id string) (bool, error)
}
