package templater

import (
	"bytes"
	"fmt"
	"text/template"
)

// Evaluate evaluates the go template to create the output text
func Evaluate(funcMap template.FuncMap, templateData interface{}, templateText, path, message string) (string, error) {
	tmpl, err := template.New("value.gotmpl").Option("missingkey=error").Funcs(funcMap).Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %s for %s: %w", path, message, err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, templateData)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate template file %s for %s: %w", path, message, err)
	}
	return buf.String(), nil
}
