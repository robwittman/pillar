package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

// RenderTaskTemplate renders a Go text/template against a JSON payload.
// If the template is empty, the raw payload is returned as the prompt.
func RenderTaskTemplate(tmpl string, payload json.RawMessage) (string, error) {
	if tmpl == "" {
		return string(payload), nil
	}

	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return "", fmt.Errorf("parsing payload for template: %w", err)
	}

	t, err := template.New("task").Option("missingkey=zero").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parsing task template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing task template: %w", err)
	}

	return buf.String(), nil
}
