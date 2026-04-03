package gomime

import "strings"

// parseHeaderParams parses a header value with parameters, e.g.
// "text/plain; charset=utf-8; name=\"file.txt\"". Returns the main
// value and a map of lowercase parameter names to their values.
func parseHeaderParams(s string) (string, map[string]string) {
	params := make(map[string]string)
	parts := splitSemicolon(s)
	if len(parts) == 0 {
		return "", params
	}

	mainValue := strings.TrimSpace(parts[0])

	for _, part := range parts[1:] {
		key, value := parseParam(strings.TrimSpace(part))
		if key != "" {
			params[strings.ToLower(key)] = value
		}
	}

	return mainValue, params
}

// splitSemicolon splits s on semicolons not inside quoted strings.
func splitSemicolon(s string) []string {
	var parts []string
	var buf strings.Builder
	inQuote := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"':
			inQuote = !inQuote
			buf.WriteByte(c)
		case c == '\\' && inQuote && i+1 < len(s):
			buf.WriteByte(c)
			i++
			buf.WriteByte(s[i])
		case c == ';' && !inQuote:
			parts = append(parts, buf.String())
			buf.Reset()
		default:
			buf.WriteByte(c)
		}
	}
	if buf.Len() > 0 {
		parts = append(parts, buf.String())
	}
	return parts
}

// parseParam parses "key=value" or "key=\"value\"".
func parseParam(s string) (string, string) {
	i := strings.IndexByte(s, '=')
	if i < 0 {
		return strings.TrimSpace(s), ""
	}
	key := strings.TrimSpace(s[:i])
	value := strings.TrimSpace(s[i+1:])

	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = unquote(value[1 : len(value)-1])
	}
	return key, value
}

// unquote removes backslash escapes from a quoted string body.
func unquote(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}
	var buf strings.Builder
	buf.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
		}
		buf.WriteByte(s[i])
	}
	return buf.String()
}
