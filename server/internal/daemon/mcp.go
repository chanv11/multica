package daemon

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// varPattern matches ${VAR_NAME} placeholders in string values.
var varPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// resolveMCPVars walks a JSON config recursively and replaces ${VAR} placeholders
// with values from the env map. Returns an error listing all variables that
// could not be resolved.
func resolveMCPVars(config []byte, env map[string]string) ([]byte, error) {
	if env == nil {
		env = map[string]string{}
	}

	var root any
	if err := json.Unmarshal(config, &root); err != nil {
		return nil, fmt.Errorf("invalid JSON config: %w", err)
	}

	var missing []string
	resolved := resolveValue(root, env, &missing)

	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("unresolved ${%s} in MCP config", strings.Join(missing, "}, ${"))
	}

	return json.Marshal(resolved)
}

// resolveValue recursively walks a JSON value and resolves placeholders in strings.
func resolveValue(v any, env map[string]string, missing *[]string) any {
	switch val := v.(type) {
	case map[string]any:
		result := make(map[string]any, len(val))
		for k, v := range val {
			result[k] = resolveValue(v, env, missing)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, elem := range val {
			result[i] = resolveValue(elem, env, missing)
		}
		return result
	case string:
		return resolveString(val, env, missing)
	default:
		// Numbers, booleans, null — return as-is.
		return v
	}
}

// resolveString replaces ${VAR} patterns in a single string value.
func resolveString(s string, env map[string]string, missing *[]string) string {
	if !strings.Contains(s, "${") {
		return s
	}

	// Collect unique missing vars from this string before replacing.
	vars := varPattern.FindAllStringSubmatch(s, -1)
	seen := make(map[string]bool, len(vars))
	for _, match := range vars {
		name := match[1]
		if _, ok := env[name]; !ok && !seen[name] {
			*missing = append(*missing, name)
			seen[name] = true
		}
	}

	if len(*missing) > 0 {
		// Don't do partial replacement — leave string as-is on error.
		return s
	}

	return varPattern.ReplaceAllStringFunc(s, func(match string) string {
		sub := varPattern.FindStringSubmatch(match)
		if len(sub) >= 2 {
			return env[sub[1]]
		}
		return match
	})
}
