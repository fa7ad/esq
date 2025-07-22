package options

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// QueryOptions holds query-related fields.
type QueryOptions struct {
	KQL       string
	DSL       string
	Lucene    string
	QueryFile string

	From string
	To   string

	Size int
}

type anyMap = map[string]any

func (q *QueryOptions) normalize() (string, error) {
	queryBody := anyMap{
		"query": anyMap{},
	}
	var err error

	switch {

	case q.QueryFile != "":
		data, err := os.ReadFile(q.QueryFile)
		if err != nil {
			return "", fmt.Errorf("error reading query file '%s': %w", q.QueryFile, err)
		}
		q.DSL = strings.TrimSpace(string(data))
		fallthrough
	case q.DSL != "":
		queryBody = anyMap{}
		if err = json.Unmarshal([]byte(q.DSL), &queryBody); err != nil {
			return "", fmt.Errorf("invalid JSON for DSL query. Please ensure it's a valid JSON object: %w", err)
		}
	case q.KQL != "":
		if queryMap, ok := queryBody["query"].(anyMap); ok {
			queryMap["query_string"] = anyMap{
				"query":            q.KQL,
				"analyze_wildcard": true,
			}
		}
	case q.Lucene != "":
		if queryMap, ok := queryBody["query"].(anyMap); ok {
			queryMap["query_string"] = anyMap{
				"query":            q.Lucene,
				"default_operator": "AND",
			}
		}

	}

	var tsRange anyMap
	if q.From != "" || q.To != "" {
		tsRange = make(anyMap)
		if q.From != "" {
			tsRange["gte"] = q.From
		}
		if q.To != "" {
			tsRange["lte"] = q.To
		}

	}
	if tsRange != nil {
		existingQuery, hasQueryKey := queryBody["query"]
		if !hasQueryKey {
			existingQuery = anyMap{"match_all": anyMap{}}
		}
		queryBody["query"] = anyMap{
			"bool": anyMap{
				"must": []any{
					existingQuery,
					anyMap{
						"range": anyMap{
							"@timestamp": tsRange,
						},
					},
				},
			},
		}
	}

	jsonData, err := json.Marshal(queryBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal query: %w", err)
	}

	return string(jsonData), nil
}

func (q *QueryOptions) ToQueryBody() (io.Reader, error) {
	dsl, err := q.normalize()
	if err != nil {
		return nil, fmt.Errorf("failed to normalize query options: %w", err)
	}

	return strings.NewReader(dsl), nil
}
