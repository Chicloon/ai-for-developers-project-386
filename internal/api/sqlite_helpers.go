package api

import (
	"database/sql"
	"strings"
)

func placeholders(n int) string {
	if n == 0 {
		return ""
	}
	b := make([]byte, 0, n*2)
	for i := 0; i < n; i++ {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, '?')
	}
	return string(b)
}

func parseGroupConcat(ns sql.NullString) []string {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	parts := strings.Split(ns.String, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return filterEmptyUUIDs(out)
}
