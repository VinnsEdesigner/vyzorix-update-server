package notifications

import (
	"bytes"
	"fmt"
	"text/template"
)

// Slack message template for alert notifications. Renders a Slack
// attachment-style payload with color, title, and fields.
const slackAlertTemplateText = `{
	"attachments": [
		{
			"color": "{{.Color}}",
			"title": "{{.Title}}",
			"text": "{{.Text}}",
			"fields": [
				{{range $i, $f := .Fields}}{{if $i}}, {{end}}{"title": "{{$f.Title}}", "value": "{{$f.Value}}", "short": {{$f.Short}}}{{end}}
			]
		}
	]
}`

var slackAlertTemplate = template.Must(template.New("slack").Parse(slackAlertTemplateText))

// SlackField is one Slack attachment field.
type SlackField struct {
	Title string
	Value string
	Short bool
}

// SlackAlertPayload is the input for slackAlertTemplateText.
type SlackAlertPayload struct {
	Color  string
	Title  string
	Text   string
	Fields []SlackField
}

// RenderSlackAlert renders the alert Slack payload; returns the JSON bytes.
func RenderSlackAlert(state, ruleName, metric, condition, threshold, value string, alertFields map[string]string) ([]byte, error) {
	color := "good"
	switch state {
	case "firing":
		color = "danger"
	case "pending":
		color = "warning"
	}

	fields := make([]SlackField, 0, 4+len(alertFields))
	fields = append(fields,
		SlackField{Title: "Metric", Value: metric, Short: true},
		SlackField{Title: "Condition", Value: condition, Short: true},
		SlackField{Title: "Threshold", Value: threshold, Short: true},
		SlackField{Title: "Value", Value: value, Short: true},
	)
	for k, v := range alertFields {
		fields = append(fields, SlackField{Title: k, Value: v, Short: false})
	}

	payload := SlackAlertPayload{
		Color:  color,
		Title:  fmt.Sprintf("[%s] %s", state, ruleName),
		Text:   fmt.Sprintf("Alert state changed: %s", state),
		Fields: fields,
	}

	var buf bytes.Buffer
	if err := slackAlertTemplate.ExecuteTemplate(&buf, "slack", payload); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
