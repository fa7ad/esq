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

type JSONObject = map[string]any

func (q *QueryOptions) normalize() (string, error) {
	queryBody := JSONObject{
		"query": JSONObject{},
	}

	switch {
	case q.QueryFile != "":
		data, err := os.ReadFile(q.QueryFile)
		if err != nil {
			return "", fmt.Errorf("error reading query file '%s': %w", q.QueryFile, err)
		}
		q.DSL = strings.TrimSpace(string(data))
		fallthrough
	case q.DSL != "":
		var dslBody JSONObject
		if err := json.Unmarshal([]byte(q.DSL), &dslBody); err != nil {
			return "", fmt.Errorf("invalid JSON for DSL query: %w", err)
		}
		queryBody = dslBody
	case q.KQL != "":
		if queryClause, ok := queryBody["query"].(JSONObject); ok {
			queryClause["query_string"] = JSONObject{
				"query":            q.KQL,
				"analyze_wildcard": true,
			}
		}
	case q.Lucene != "":
		if queryClause, ok := queryBody["query"].(JSONObject); ok {
			queryClause["query_string"] = JSONObject{
				"query":            q.Lucene,
				"default_operator": "AND",
			}
		}
	}

	var tsQuery JSONObject
	if q.From != "" || q.To != "" {
		tsRange := JSONObject{}
		if q.From != "" {
			tsRange["gte"] = q.From
		}
		if q.To != "" {
			tsRange["lte"] = q.To
		}
		tsQuery = JSONObject{
			"range": JSONObject{
				"@timestamp": tsRange,
			},
		}
	}

	if tsQuery != nil {
		existingQuery, hasQuery := queryBody["query"]
		if !hasQuery {
			existingQuery = JSONObject{"match_all": JSONObject{}}
		}
		queryBody["query"] = JSONObject{
			"bool": JSONObject{
				"must": []any{
					existingQuery,
					tsQuery,
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
