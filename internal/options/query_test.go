package options

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryOptions_normalize(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "query-*.json")
	require.NoError(t, err, "Failed to create temp file")
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString(`{"query":{"term":{"file_field":"file_value"}}}`)
	require.NoError(t, err, "Failed to write to temp file")
	tmpFile.Close()

	testCases := []struct {
		name           string
		opts           QueryOptions
		wantContain    []string
		wantNotContain []string
	}{
		{
			name:        "KQL only",
			opts:        QueryOptions{KQL: "user:test"},
			wantContain: []string{`"query_string"`, `"query":"user:test"`, `"analyze_wildcard":true`},
		},
		{
			name: "Lucene only",
			opts: QueryOptions{Lucene: "field:value"},
			// FIX: Changed "AND" to "and" to match the library's JSON output.
			wantContain: []string{`"query_string"`, `"query":"field:value"`, `"default_operator":"and"`},
		},
		{
			name:        "KQL with time range",
			opts:        QueryOptions{KQL: "user:test", From: "now-1h", To: "now"},
			wantContain: []string{`"bool"`, `"must"`, `"range"`, `"timestamp"`, `"gte":"now-1h"`, `"lte":"now"`},
		},
		{
			name:        "Time range only",
			opts:        QueryOptions{From: "2025-01-01T00:00:00Z"},
			wantContain: []string{`"bool"`, `"must"`, `"range"`, `"timestamp"`, `"match_all"`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.opts.normalize()
			require.NoError(t, err)

			for _, s := range tc.wantContain {
				assert.Contains(t, got, s)
			}
			for _, s := range tc.wantNotContain {
				assert.NotContains(t, got, s)
			}
		})
	}
}
