package utils

import (
	"bytes"
	"fmt"
	"sort"
)

func MapToString(m map[string]any) string {
	var b bytes.Buffer
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Sort keys for consistent string representation

	b.WriteString("{")
	for i, k := range keys {
		v := m[k]
		if i > 0 {
			b.WriteString(", ")
		}
		// Handle different types within 'any'
		switch val := v.(type) {
		case string:
			fmt.Fprintf(&b, "\"%s\":\"%s\"", k, val)
		case int, int8, int16, int32, int64:
			fmt.Fprintf(&b, "\"%s\":%d", k, val)
		case float32, float64:
			fmt.Fprintf(&b, "\"%s\":%f", k, val)
		case bool:
			fmt.Fprintf(&b, "\"%s\":%t", k, val)
		default:
			// Fallback for other types, using default string representation
			fmt.Fprintf(&b, "\"%s\":%v", k, val)
		}
	}
	b.WriteString("}")
	return b.String()
}
