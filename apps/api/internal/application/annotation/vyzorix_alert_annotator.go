package annotation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/alert"
)

// AlertAnnotator marks alert transitions on the fleet timeline. Implements
// alert.Annotator.
type AlertAnnotator struct {
	service *Service
}

// NewAlertAnnotator creates an AlertAnnotator.
func NewAlertAnnotator(service *Service) *AlertAnnotator {
	return &AlertAnnotator{service: service}
}

// Annotate writes a timeline annotation for a firing or resolved alert.
func (a *AlertAnnotator) Annotate(ctx context.Context, rule *alert.Rule, transition *alert.Transition) error {
	title := rule.Name
	if transition.Firing() {
		title = "[firing] " + rule.Name
	} else if transition.Resolved() {
		title = "[resolved] " + rule.Name
	}

	in := &AnnotationInput{
		OrgID:     rule.OrgID,
		Title:     title,
		Text:      buildAnnotationText(rule, transition),
		Tags:      []string{"alert", string(rule.Metric), string(transition.To)},
		Source:    "alert",
		StartTime: transition.At,
	}
	_, err := a.service.Create(ctx, in)
	return err
}

// AnnotateRollout marks an update rollout start on the fleet timeline.
func (a *AlertAnnotator) AnnotateRollout(ctx context.Context, orgID, version, deviceCount, initiatedBy string) error {
	in := &AnnotationInput{
		OrgID:     orgID,
		Title:     "rollout " + version + " started",
		Text:      "Rollout of firmware " + version + " started to " + deviceCount + " devices by " + initiatedBy,
		Tags:      []string{"rollout", "firmware", version},
		Source:    "rollout",
		StartTime: time.Now(),
	}
	_, err := a.service.Create(ctx, in)
	return err
}

func buildAnnotationText(rule *alert.Rule, transition *alert.Transition) string {
	return "metric " + string(rule.Metric) + " " + string(rule.Condition) + " " +
		formatFloat(rule.Threshold) + " → " + string(transition.To) +
		" at value " + formatFloat(transition.Value)
}

func formatFloat(v float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", v), "0"), ".")
}
