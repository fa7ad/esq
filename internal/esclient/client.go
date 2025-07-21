package esclient

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v9"
)

type esClient struct {
	node     string
	apiKey   string
	username string
	password string
	client   *elasticsearch.Client
}

func NewElasticsearchClient(node, apiKey, username, password string) (*esClient, error) {
	if node == "" {
		return nil, fmt.Errorf("elasticsearch node URL must be provided")
	}
	cfg := elasticsearch.Config{
		Addresses: []string{
			node,
		},
	}

	if apiKey != "" {
		cfg.APIKey = apiKey
	} else if username != "" && password != "" {
		cfg.Username = username
		cfg.Password = password
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	return &esClient{
		node:     node,
		apiKey:   apiKey,
		username: username,
		password: password,
		client:   client,
	}, nil
}

// buildQueryBody creates an io.Reader for the search request body.
// It builds the query based on the first available query type in the order: KQL, DSL, Lucene, Query File.
func buildQueryBody(kql, dsl, lucene, queryFile string) (io.Reader, error) {
	if kql != "" {
		// For KQL, adding analyze_wildcard is a good practice.
		query := fmt.Sprintf(
			`{"query": {"query_string": {"query": %q, "analyze_wildcard": true}}}`,
			kql,
		)
		return strings.NewReader(query), nil
	}
	if dsl != "" {
		// DSL is used as the raw query body
		return strings.NewReader(dsl), nil
	}
	if lucene != "" {
		// For Lucene, setting a default operator is common.
		query := fmt.Sprintf(
			`{"query": {"query_string": {"query": %q, "default_operator": "AND"}}}`,
			lucene,
		)
		return strings.NewReader(query), nil
	}
	if queryFile != "" {
		// Read query from file
		content, err := os.ReadFile(queryFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read query file '%s': %w", queryFile, err)
		}
		return strings.NewReader(string(content)), nil
	}

	return nil, fmt.Errorf("no query provided: please use --kql, --dsl, --lucene, or --query-file")
}

// Search executes a search query against a specified index.
func (c *esClient) Search(index, kql, dsl, lucene, queryFile string, size int) (map[string]interface{}, error) {
	// Build the query body using the helper function
	queryBody, err := buildQueryBody(kql, dsl, lucene, queryFile)
	if err != nil {
		return nil, err
	}

	// Execute the search
	res, err := c.client.Search(
		c.client.Search.WithIndex(index),
		c.client.Search.WithBody(queryBody),
		c.client.Search.WithSize(size),
		c.client.Search.WithTrackTotalHits(true),
		c.client.Search.WithPretty(),
	)

	if err != nil {
		return nil, fmt.Errorf("elasticsearch search failed: %w", err)
	}
	defer res.Body.Close()

	// Check for Elasticsearch-specific errors
	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error: [%s] %s", res.Status(), string(bodyBytes))
	}

	// Decode the JSON response into a map
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("failed to parse search response body: %w", err)
	}

	return r, nil
}
