package esclient

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/fa7ad/esq/internal/options"
)

// esClient represents an Elasticsearch client.
type esClient struct {
	client *elasticsearch.Client
}

// NewElasticsearchClient creates a new Elasticsearch client.
func NewElasticsearchClient(authOpts options.AuthOptions, opts options.ElasticOptions) (*esClient, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{opts.Node},
	}

	authOpts.UpdateConfig(&cfg)

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	return &esClient{client}, nil
}

// Search executes a search query against a specified index.
func (c *esClient) Search(esOpts options.ElasticOptions) (map[string]any, error) {
	queryBody, err := esOpts.ToQueryBody()
	if err != nil {
		return nil, err
	}

	// Execute the search
	res, err := c.client.Search(
		c.client.Search.WithIndex(esOpts.Index),
		c.client.Search.WithBody(queryBody),
		c.client.Search.WithSize(esOpts.Size),
		c.client.Search.WithTrackTotalHits(true),
		c.client.Search.WithPretty(),
		c.client.Search.WithFilterPath(
			"hits.hits",
			"took",
			"timed_out",
			"_shards",
		),
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
	var r map[string]any
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("failed to parse search response body: %w", err)
	}

	if hits, found := r["hits"].(map[string]any); found {
		if hitsArray, ok := hits["hits"].([]any); ok {
			r["hits"] = hitsArray // Replace "hits" with the array of hits
		}
	}

	return r, nil
}
