package browser

import (
	"regexp"
	"strings"
)

var createTableTargetPattern = regexp.MustCompile(`(?is)^\s*create\s+table\s+(?:if\s+not\s+exists\s+)?((?:"[^"]+"|[a-zA-Z_][\w$]*)(?:\.(?:"[^"]+"|[a-zA-Z_][\w$]*))?)`)
var alterTableTargetPattern = regexp.MustCompile(`(?is)^\s*alter\s+table\s+(?:if\s+exists\s+)?((?:"[^"]+"|[a-zA-Z_][\w$]*)(?:\.(?:"[^"]+"|[a-zA-Z_][\w$]*))?)`)
var dropTableTargetPattern = regexp.MustCompile(`(?is)^\s*drop\s+table\s+(?:if\s+exists\s+)?((?:"[^"]+"|[a-zA-Z_][\w$]*)(?:\.(?:"[^"]+"|[a-zA-Z_][\w$]*))?)`)
var insertTargetPattern = regexp.MustCompile(`(?is)^\s*insert\s+into\s+((?:"[^"]+"|[a-zA-Z_][\w$]*)(?:\.(?:"[^"]+"|[a-zA-Z_][\w$]*))?)`)
var updateTargetPattern = regexp.MustCompile(`(?is)^\s*update\s+((?:"[^"]+"|[a-zA-Z_][\w$]*)(?:\.(?:"[^"]+"|[a-zA-Z_][\w$]*))?)`)
var deleteTargetPattern = regexp.MustCompile(`(?is)^\s*delete\s+from\s+((?:"[^"]+"|[a-zA-Z_][\w$]*)(?:\.(?:"[^"]+"|[a-zA-Z_][\w$]*))?)`)
var createTablePattern = regexp.MustCompile(`(?is)^\s*create\s+table\b`)
var alterTablePattern = regexp.MustCompile(`(?is)^\s*alter\s+table\b`)
var dropTablePattern = regexp.MustCompile(`(?is)^\s*drop\s+table\b`)

func parseCreateTableTarget(sql string) (schema, table string, ok bool) {
	sql = stripLeadingSQLComments(sql)
	match := createTableTargetPattern.FindStringSubmatch(sql)
	return parseQualifiedTarget(match)
}

func parseTargetByMode(sql string, mode editorMode) (schema, table string, ok bool) {
	sql = stripLeadingSQLComments(sql)
	switch mode {
	case editorCreate:
		return parseCreateTableTarget(sql)
	case editorAlter:
		return parseQualifiedTarget(alterTableTargetPattern.FindStringSubmatch(sql))
	case editorDrop:
		return parseQualifiedTarget(dropTableTargetPattern.FindStringSubmatch(sql))
	case editorInsert:
		return parseQualifiedTarget(insertTargetPattern.FindStringSubmatch(sql))
	case editorUpdate:
		return parseQualifiedTarget(updateTargetPattern.FindStringSubmatch(sql))
	case editorDelete:
		return parseQualifiedTarget(deleteTargetPattern.FindStringSubmatch(sql))
	default:
		return "", "", false
	}
}

func parseQualifiedTarget(match []string) (schema, table string, ok bool) {
	if len(match) < 2 {
		return "", "", false
	}

	parts := splitQualifiedIdentifier(match[1])
	if len(parts) == 1 {
		return "", unquoteIdentifier(parts[0]), true
	}
	if len(parts) == 2 {
		return unquoteIdentifier(parts[0]), unquoteIdentifier(parts[1]), true
	}

	return "", "", false
}

func splitQualifiedIdentifier(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	parts := make([]string, 0, 2)
	var current strings.Builder
	inQuotes := false
	for i := 0; i < len(value); i++ {
		ch := value[i]
		switch ch {
		case '"':
			inQuotes = !inQuotes
			current.WriteByte(ch)
		case '.':
			if inQuotes {
				current.WriteByte(ch)
				continue
			}
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}
	return parts
}

func unquoteIdentifier(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
		value = strings.ReplaceAll(value, `""`, `"`)
	}
	return value
}

func detectEditorMode(sql string, fallback editorMode) editorMode {
	sql = stripLeadingSQLComments(sql)
	switch {
	case createTablePattern.MatchString(sql):
		return editorCreate
	case alterTablePattern.MatchString(sql):
		return editorAlter
	case dropTablePattern.MatchString(sql):
		return editorDrop
	default:
		return fallback
	}
}

func stripLeadingSQLComments(sql string) string {
	for {
		sql = strings.TrimSpace(sql)
		if strings.HasPrefix(sql, "--") {
			if idx := strings.IndexByte(sql, '\n'); idx >= 0 {
				sql = sql[idx+1:]
				continue
			}
			return ""
		}
		if strings.HasPrefix(sql, "/*") {
			if idx := strings.Index(sql, "*/"); idx >= 0 {
				sql = sql[idx+2:]
				continue
			}
			return ""
		}
		return sql
	}
}
