package alert

import (
	"context"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/alert"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/uuid"
)

var (
	// ErrInvalidRule is returned when a rule payload fails validation.
	ErrInvalidRule = errors.New("invalid alert rule")
	// ErrRuleNotFound is returned when a rule does not exist.
	ErrRuleNotFound = errors.New("alert rule not found")
)

// RuleInput carries the mutable fields of a rule create/update request.
type RuleInput struct {
	OrgID                 string
	Name                  string
	WebhookURL            string
	Metric                alert.Metric
	Condition             alert.Condition
	Threshold             float64
	ForSeconds            int
	NotifyIntervalSeconds int
	Enabled               bool
}

// RuleInputFromArgs builds RuleInput from GraphQL args map. Missing optional
// fields use zero defaults (ForSeconds, NotifyIntervalSeconds, WebhookURL).
func RuleInputFromArgs(args map[string]interface{}, orgID string) *RuleInput {
	in := &RuleInput{OrgID: orgID}
	if v, ok := args["name"].(string); ok {
		in.Name = v
	}
	if v, ok := args["webhookUrl"].(string); ok {
		in.WebhookURL = v
	}
	if v, ok := args["metric"].(string); ok {
		in.Metric = alert.Metric(v)
	}
	if v, ok := args["condition"].(string); ok {
		in.Condition = alert.Condition(v)
	}
	if v, ok := args["threshold"].(float64); ok {
		in.Threshold = v
	}
	if v, ok := args["forSeconds"].(int); ok {
		in.ForSeconds = v
	}
	if v, ok := args["notifyIntervalSeconds"].(int); ok {
		in.NotifyIntervalSeconds = v
	}
	if v, ok := args["enabled"].(bool); ok {
		in.Enabled = v
	}
	return in
}

// RuleView pairs a rule with its runtime instance state.
type RuleView struct {
	EvaluatedAt time.Time
	Rule        *alert.Rule
	State       alert.State
	Value       float64
}

// Service provides org-scoped alert rule CRUD and history access.
type Service struct {
	rules   alert.Repository
	states  alert.StateRepository
	history alert.HistoryRepository
}

func NewService(rules alert.Repository, states alert.StateRepository, history alert.HistoryRepository) *Service {
	return &Service{rules: rules, states: states, history: history}
}

// History returns transition events for an org, optionally narrowed to a rule.
func (s *Service) History(ctx context.Context, orgID, ruleID string, limit int) ([]*alert.Event, error) {
	if s.history == nil {
		return nil, nil
	}
	return s.history.ListByOrg(ctx, orgID, ruleID, limit)
}

// CreateRule validates and persists a new rule.
func (s *Service) CreateRule(ctx context.Context, in *RuleInput) (*alert.Rule, error) {
	now := time.Now()
	rule := &alert.Rule{
		ID:                    uuid.New(),
		CreatedAt:             now,
		UpdatedAt:             now,
		OrgID:                 in.OrgID,
		Name:                  in.Name,
		WebhookURL:            in.WebhookURL,
		Metric:                in.Metric,
		Condition:             in.Condition,
		Threshold:             in.Threshold,
		ForSeconds:            in.ForSeconds,
		NotifyIntervalSeconds: in.NotifyIntervalSeconds,
		Enabled:               in.Enabled,
	}
	if err := rule.Validate(); err != nil {
		return nil, errors.Join(ErrInvalidRule, err)
	}
	if err := s.rules.Save(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

// UpdateRule replaces the mutable fields of an existing rule. Disabling a
// rule clears its instance so it leaves no stale firing state behind.
func (s *Service) UpdateRule(ctx context.Context, orgID, id string, in *RuleInput) (*alert.Rule, error) {
	rule, err := s.getScoped(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	rule.Name = in.Name
	rule.WebhookURL = in.WebhookURL
	rule.Metric = in.Metric
	rule.Condition = in.Condition
	rule.Threshold = in.Threshold
	rule.ForSeconds = in.ForSeconds
	rule.NotifyIntervalSeconds = in.NotifyIntervalSeconds
	rule.Enabled = in.Enabled
	rule.UpdatedAt = time.Now()

	if err := rule.Validate(); err != nil {
		return nil, errors.Join(ErrInvalidRule, err)
	}
	if err := s.rules.Save(ctx, rule); err != nil {
		return nil, err
	}
	if !rule.Enabled {
		if err := s.states.DeleteByRuleID(ctx, rule.ID); err != nil {
			return nil, err
		}
	}
	return rule, nil
}

// DeleteRule removes a rule and its instance.
func (s *Service) DeleteRule(ctx context.Context, orgID, id string) error {
	if _, err := s.getScoped(ctx, orgID, id); err != nil {
		return err
	}
	deleted, err := s.rules.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrRuleNotFound
	}
	return nil
}

// GetRule returns one rule with its instance state.
func (s *Service) GetRule(ctx context.Context, orgID, id string) (*RuleView, error) {
	rule, err := s.getScoped(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	return s.view(ctx, rule)
}

// ListRules returns all rules of an org with their instance states.
func (s *Service) ListRules(ctx context.Context, orgID string) ([]*RuleView, error) {
	rules, err := s.rules.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	instances, err := s.states.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}

	views := make([]*RuleView, 0, len(rules))
	for _, rule := range rules {
		v := &RuleView{Rule: rule, State: alert.StateInactive}
		if inst, ok := instances[rule.ID]; ok {
			v.State = inst.State
			v.Value = inst.LastValue
			v.EvaluatedAt = inst.LastEvaluated
		}
		views = append(views, v)
	}
	return views, nil
}

func (s *Service) getScoped(ctx context.Context, orgID, id string) (*alert.Rule, error) {
	rule, err := s.rules.GetByID(ctx, id)
	if errors.Is(err, alert.ErrNotFound) {
		return nil, ErrRuleNotFound
	}
	if err != nil {
		return nil, err
	}
	if rule.OrgID != orgID {
		return nil, ErrRuleNotFound
	}
	return rule, nil
}

func (s *Service) view(ctx context.Context, rule *alert.Rule) (*RuleView, error) {
	inst, err := s.states.GetByRuleID(ctx, rule.ID)
	if err != nil {
		return nil, err
	}
	return &RuleView{
		Rule:        rule,
		State:       inst.State,
		Value:       inst.LastValue,
		EvaluatedAt: inst.LastEvaluated,
	}, nil
}
