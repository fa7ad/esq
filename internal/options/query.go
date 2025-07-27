package options

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"k8s.io/utils/ptr"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/operator"
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

// normalize normalizes the query options into a single DSL query.
func (q *QueryOptions) normalize() (string, error) {
	queryBody := types.SearchRequestBody{
		Query: &types.Query{},
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
		var dslBody types.SearchRequestBody
		if err := json.Unmarshal([]byte(q.DSL), &dslBody); err != nil {
			return "", fmt.Errorf("invalid JSON for DSL query: %w", err)
		}
		queryBody = dslBody
	case q.KQL != "":
		queryBody.Query = &types.Query{
			QueryString: &types.QueryStringQuery{
				Query:           q.KQL,
				AnalyzeWildcard: ptr.To(true),
				Lenient:         ptr.To(true),
			},
		}
	case q.Lucene != "":
		queryBody.Query = &types.Query{
			QueryString: &types.QueryStringQuery{
				Query:           q.Lucene,
				DefaultOperator: &operator.And,
				Lenient:         ptr.To(true),
			},
		}
	}

	var tsQuery *types.Query
	if q.From != "" || q.To != "" {
		tsRange := types.DateRangeQuery{}
		if q.From != "" {
			tsRange.Gte = &q.From
		}
		if q.To != "" {
			tsRange.Lte = &q.To
		}
		tsQuery = &types.Query{
			Range: map[string]types.RangeQuery{
				"timestamp": &tsRange,
			},
		}
	}

	if tsQuery != nil {
		existingQuery := queryBody.Query
		isQueryEmpty := q.KQL == "" && q.DSL == "" && q.Lucene == "" && q.QueryFile == ""
		if isQueryEmpty {
			existingQuery = &types.Query{MatchAll: &types.MatchAllQuery{}}
		}
		queryBody.Query = &types.Query{
			Bool: &types.BoolQuery{
				Must: []types.Query{
					*existingQuery,
					*tsQuery,
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

// ToQueryBody converts the query options into a query body reader.
func (q *QueryOptions) ToQueryBody() (io.Reader, error) {
	dsl, err := q.normalize()
	if err != nil {
		return nil, fmt.Errorf("failed to normalize query options: %w", err)
	}

	return strings.NewReader(dsl), nil
}
