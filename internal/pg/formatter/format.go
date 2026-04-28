package formatter

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgtype"
)

func Value(oid uint32, value any) string {
	if value == nil {
		return "NULL"
	}

	switch oid {
	case pgtype.UUIDOID:
		if text, ok := formatUUIDValue(value); ok {
			return text
		}
	case pgtype.ByteaOID:
		if text, ok := formatByteaValue(value); ok {
			return text
		}
	case pgtype.JSONOID, pgtype.JSONBOID:
		if text, ok := formatJSONValue(value); ok {
			return text
		}
	case pgtype.DateOID:
		if ts, ok := value.(time.Time); ok {
			return ts.Format("2006-01-02")
		}
	case pgtype.TimestampOID:
		if ts, ok := value.(time.Time); ok {
			return ts.Format("2006-01-02 15:04:05")
		}
	case pgtype.TimestamptzOID:
		if ts, ok := value.(time.Time); ok {
			return ts.Format(time.RFC3339)
		}
	case pgtype.NumericOID:
		if text, ok := formatNumericValue(value); ok {
			return text
		}
	}

	switch v := value.(type) {
	case string:
		return compactWhitespace(v)
	case []byte:
		if utf8.Valid(v) {
			return compactWhitespace(string(v))
		}
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	case fmt.Stringer:
		return compactWhitespace(v.String())
	}

	if text, ok := marshalCompactJSON(value); ok {
		return text
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		parts := make([]string, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			parts = append(parts, Value(0, rv.Index(i).Interface()))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case reflect.Map:
		if text, ok := marshalCompactJSON(value); ok {
			return text
		}
	}

	return compactWhitespace(fmt.Sprintf("%v", value))
}

func formatUUIDValue(value any) (string, bool) {
	switch v := value.(type) {
	case [16]byte:
		return formatUUIDBytes(v[:]), true
	case []byte:
		if len(v) == 16 {
			return formatUUIDBytes(v), true
		}
	case string:
		return compactWhitespace(v), true
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Array && rv.Len() == 16 {
		buf := make([]byte, 16)
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i)
			if !elem.CanUint() {
				return "", false
			}
			buf[i] = byte(elem.Uint())
		}
		return formatUUIDBytes(buf), true
	}

	return "", false
}

func formatUUIDBytes(value []byte) string {
	if len(value) != 16 {
		return "\\x" + strings.ToUpper(hex.EncodeToString(value))
	}

	return fmt.Sprintf(
		"%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		value[0], value[1], value[2], value[3],
		value[4], value[5],
		value[6], value[7],
		value[8], value[9],
		value[10], value[11], value[12], value[13], value[14], value[15],
	)
}

func formatByteaValue(value any) (string, bool) {
	raw, ok := value.([]byte)
	if !ok {
		return "", false
	}
	if len(raw) == 0 {
		return "", true
	}
	if utf8.Valid(raw) {
		return compactWhitespace(string(raw)), true
	}

	const previewBytes = 32
	if len(raw) <= previewBytes {
		return "\\x" + strings.ToUpper(hex.EncodeToString(raw)), true
	}

	return fmt.Sprintf("\\x%s… (%d bytes)", strings.ToUpper(hex.EncodeToString(raw[:previewBytes])), len(raw)), true
}

func formatJSONValue(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return compactJSONString(v), true
	case []byte:
		return compactJSONString(string(v)), true
	default:
		return marshalCompactJSON(value)
	}
}

func compactJSONString(value string) string {
	var parsed any
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		return compactWhitespace(value)
	}

	buf, err := json.Marshal(parsed)
	if err != nil {
		return compactWhitespace(value)
	}

	return compactWhitespace(string(buf))
}

func formatNumericValue(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return compactWhitespace(v), true
	case []byte:
		return compactWhitespace(string(v)), true
	case fmt.Stringer:
		return compactWhitespace(v.String()), true
	}
	return "", false
}

func marshalCompactJSON(value any) (string, bool) {
	buf, err := json.Marshal(value)
	if err != nil {
		return "", false
	}
	return compactWhitespace(string(buf)), true
}

func compactWhitespace(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\t", " ")
	return strings.Join(strings.Fields(value), " ")
}
