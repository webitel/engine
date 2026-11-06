package wsapi

import "fmt"

func parseWSInputVariable(v any) string {
	if r, ok := v.(string); ok {
		return r
	}

	return fmt.Sprintf("%v", v)
}

func variablesFromMap(m map[string]any, name string) map[string]string {
	vi, ok := m[name].(map[string]any)
	if !ok {
		return nil
	}

	vars := make(map[string]string, len(vi))
	for k, v := range vi {
		vars[k] = fmt.Sprintf("%v", v)
	}

	return vars
}
