package provisioning

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

type Config struct {
	Organizations []OrgConfig      `json:"organizations"`
	Operators     []OperatorConfig `json:"operators"`
	DeviceGroups  []GroupConfig    `json:"device_groups"`
	APIKeys       []APIKeyConfig   `json:"api_keys"`
	Grants        []GrantConfig    `json:"permission_grants"`
}

type OrgConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type OperatorConfig struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	OrgName  string `json:"org_name"`
	Role     string `json:"role"`
}

type GroupConfig struct {
	Name           string   `json:"name"`
	OrgName        string   `json:"org_name"`
	OperatorEmails []string `json:"operator_emails"`
}

type APIKeyConfig struct {
	Name    string `json:"name"`
	OrgName string `json:"org_name"`
	Scope   string `json:"scope"`
}

type GrantConfig struct {
	SubjectType string   `json:"subject_type"`
	SubjectID   string   `json:"subject_id"`
	OrgName     string   `json:"org_name"`
	Scope       string   `json:"scope"`
	Actions     []string `json:"actions"`
}

type Provisioner struct {
	log *slog.Logger
}

func New(log *slog.Logger) *Provisioner {
	return &Provisioner{log: log}
}

func (p *Provisioner) LoadAndApply(ctx context.Context, path string) error {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			p.log.Info("no provisioning file found, skipping", "path", path)
			return nil
		}
		return fmt.Errorf("read provisioning file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse provisioning file: %w", err)
	}

	p.log.Info("provisioning file loaded",
		"orgs", len(cfg.Organizations),
		"operators", len(cfg.Operators),
		"groups", len(cfg.DeviceGroups),
		"keys", len(cfg.APIKeys),
		"grants", len(cfg.Grants),
	)

	return nil
}
