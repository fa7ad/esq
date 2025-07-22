package options

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// QueryOptions holds query-related fields.
// It includes KQL, DSL, Lucene, and a query file.
type QueryOptions struct {
	KQL       string
	DSL       string
	Lucene    string
	QueryFile string

	Size int
}

func (q *QueryOptions) normalize() (string, error) {
	if q.QueryFile != "" {
		// read the content of the query file into the DSL field
		content, err := os.ReadFile(q.QueryFile)
		if err != nil {
			return "", fmt.Errorf("failed to read query file '%s': %w", q.QueryFile, err)
		}
		q.DSL = string(content)
		q.QueryFile = "" // Clear the QueryFile field after reading
	}

	if q.KQL != "" {
		q.DSL = fmt.Sprintf(`{"query": {"query_string": {"query": %q, "analyze_wildcard": true}}}`, q.KQL)
		q.KQL = "" // Clear the KQL field after serialization
	}
	if q.Lucene != "" {
		q.DSL = fmt.Sprintf(`{"query": {"query_string": {"query": %q, "default_operator": "AND"}}}`, q.Lucene)
		q.Lucene = "" // Clear the Lucene field after serialization
	}

	return q.DSL, nil
}

func (q *QueryOptions) ToQueryBody() (io.Reader, error) {
	dsl, err := q.normalize()
	if err != nil {
		return nil, fmt.Errorf("failed to normalize query options: %w", err)
	}
	if dsl == "" {
		return nil, fmt.Errorf("no query provided")
	}

	return strings.NewReader(dsl), nil
}
