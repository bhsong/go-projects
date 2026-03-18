package env

import "regexp"

var varPattern = regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

func Expand(value string, env map[string]string) string {
	return varPattern.ReplaceAllStringFunc(value, func(match string) string {
		var key string

		groups := varPattern.FindStringSubmatch(match)
		if groups[1] != "" {
			key = groups[1]
		} else {
			key = groups[2]
		}

		if v, ok := env[key]; ok {
			return v
		}

		return ""

	})
}
