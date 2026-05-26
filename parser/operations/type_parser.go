package operations

import (
	"fmt"
	"strings"
	"unicode"
)

type responseComment struct {
	status      string
	jsonType    string
	goType      string
	description string
}

func normalizeGoType(goType string) string {
	goType = strings.TrimSpace(goType)
	goType = normalizeMapTypes(goType)
	goType = strings.ReplaceAll(goType, " ", "")
	goType = strings.ReplaceAll(goType, "*", "")
	if strings.HasPrefix(goType, "[]") {
		return "[]" + normalizeGoType(goType[2:])
	}
	return goType
}

func normalizeMapTypes(goType string) string {
	var b strings.Builder
	for i := 0; i < len(goType); i++ {
		if strings.HasPrefix(goType[i:], "map[") {
			end := strings.Index(goType[i:], "]")
			if end > -1 {
				b.WriteString("map[]")
				i += end
				continue
			}
		}
		b.WriteByte(goType[i])
	}
	return b.String()
}

func parseParamFields(comment string) ([]string, error) {
	if attr, rest, ok := cutSpaceField(comment); ok && strings.EqualFold(attr, "@param") {
		comment = rest
	}

	fields := make([]string, 0, 6)
	for i := 0; i < 4; i++ {
		field, rest, ok := cutSpaceField(comment)
		if !ok {
			return nil, fmt.Errorf("parseParamComment can not parse param comment \"%s\"", comment)
		}
		fields = append(fields, field)
		comment = rest
	}

	description, rest, ok := cutQuotedField(comment)
	if !ok {
		return nil, fmt.Errorf("parseParamComment can not parse param comment \"%s\"", comment)
	}
	fields = append(fields, description)
	rest = strings.TrimSpace(rest)
	if rest == "" {
		fields = append(fields, "")
		return fields, nil
	}

	example, rest, ok := cutQuotedField(rest)
	if !ok || strings.TrimSpace(rest) != "" {
		return nil, fmt.Errorf("parseParamComment can not parse param comment \"%s\"", comment)
	}
	fields = append(fields, example)
	return fields, nil
}

func parseResponseFields(comment string) (responseComment, error) {
	var parsed responseComment
	status, rest, ok := cutSpaceField(comment)
	if !ok {
		return parsed, fmt.Errorf("parseResponseComment can not parse response comment \"%s\"", comment)
	}
	parsed.status = status

	jsonType, rest, ok := cutSpaceField(rest)
	if !ok {
		parsed.description = strings.Trim(strings.TrimSpace(rest), "\"")
		return parsed, nil
	}
	if strings.HasPrefix(jsonType, "\"") {
		description := strings.TrimSpace(jsonType + " " + rest)
		parsed.description = strings.Trim(description, "\"")
		return parsed, nil
	}
	parsed.jsonType = jsonType

	goType, rest, ok := cutTypeField(rest)
	if ok {
		parsed.goType = normalizeGoType(goType)
	}
	parsed.description = strings.Trim(strings.TrimSpace(rest), "\"")
	return parsed, nil
}

func cutSpaceField(s string) (string, string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", "", false
	}
	for i, r := range s {
		if unicode.IsSpace(r) {
			return s[:i], strings.TrimSpace(s[i:]), true
		}
	}
	return s, "", true
}

func cutQuotedField(s string) (string, string, bool) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "\"") {
		return "", "", false
	}
	escaped := false
	for i := 1; i < len(s); i++ {
		switch {
		case escaped:
			escaped = false
		case s[i] == '\\':
			escaped = true
		case s[i] == '"':
			return s[1:i], s[i+1:], true
		}
	}
	return "", "", false
}

func cutTypeField(s string) (string, string, bool) {
	s = strings.TrimSpace(s)
	if s == "" || strings.HasPrefix(s, "\"") {
		return "", s, false
	}

	var squareDepth, curlyDepth int
	for i, r := range s {
		switch r {
		case '[':
			squareDepth++
		case ']':
			if squareDepth > 0 {
				squareDepth--
			}
		case '{':
			curlyDepth++
		case '}':
			if curlyDepth > 0 {
				curlyDepth--
			}
		default:
			if unicode.IsSpace(r) && squareDepth == 0 && curlyDepth == 0 {
				return s[:i], strings.TrimSpace(s[i:]), true
			}
		}
	}
	return s, "", true
}
