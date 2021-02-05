package templater

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"
)

// Evaluate evaluates the go template to create the output text
func Evaluate(funcMap template.FuncMap, templateData map[string]interface{}, templateText, path, message string) (string, error) {
	tmpl, err := template.New("value.gotmpl").Option("missingkey=error").Funcs(funcMap).Parse(templateText)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse template: %s for %s", path, message)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, templateData)
	if err != nil {
		return "", errors.Wrapf(err, "failed to evaluate template file %s for %s", path, message)
	}
	return buf.String(), nil
}
