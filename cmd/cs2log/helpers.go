package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type multiFlag []string

func (f *multiFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *multiFlag) Set(value string) error {
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			*f = append(*f, part)
		}
	}
	return nil
}

func payloadPrefix(payload string, prefixLen int) string {
	payload = strings.TrimSpace(payload)
	if prefixLen <= 0 || len(payload) <= prefixLen {
		return payload
	}
	return payload[:prefixLen]
}

func printCounts(w io.Writer, counts map[string]int, limit int) {
	rows := make([]countRow, 0, len(counts))
	for key, count := range counts {
		rows = append(rows, countRow{key: key, count: count})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].count == rows[j].count {
			return rows[i].key < rows[j].key
		}
		return rows[i].count > rows[j].count
	})
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	for _, row := range rows {
		fmt.Fprintf(w, "%8d  %s\n", row.count, row.key)
	}
}

type countRow struct {
	key   string
	count int
}
