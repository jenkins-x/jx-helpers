package outputformat

import (
	"encoding/json"
	"fmt"
	"io"

	"sigs.k8s.io/yaml"
)

// Marshal marshals the value to the given output stream with a format string using yaml or json
func Marshal(value interface{}, out io.Writer, format string) error {
	switch format {
	case "json":
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		_, e := fmt.Fprint(out, string(data))
		return e
	case "yaml":
		data, err := yaml.Marshal(value)
		if err != nil {
			return err
		}
		_, e := fmt.Fprint(out, string(data))
		return e
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
