package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fa7ad/esq/internal/options"
	"github.com/itchyny/gojq"
)

// ValidateCliArgs validates the command-line arguments.
func ValidateCliArgs(args options.CliArgs) error {
	err := ValidateElasticOptions(args.ElasticOptions)
	if err != nil {
		return fmt.Errorf("error validating elastic options: %w", err)
	}

	err = ValidateQueryOptions(args.QueryOptions)
	if err != nil {
		return fmt.Errorf("error validating query options: %w", err)
	}
	err = ValidateAuthOptions(args.AuthOptions)
	if err != nil {
		return fmt.Errorf("error validating auth options: %w", err)
	}

	err = ValidateOutputOptions(args.OutputOptions)
	if err != nil {
		return fmt.Errorf("error validating output options: %w", err)
	}

	return nil
}

// getKeys returns the keys of a map as a slice.
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ValidateQueryOptions validates the query options.
func ValidateQueryOptions(queryOptions options.QueryOptions) error {
	if queryOptions.KQL == "" && queryOptions.DSL == "" && queryOptions.Lucene == "" && queryOptions.QueryFile == "" {
		return fmt.Errorf("one of --kql, --dsl, --lucene, or --query-file must be provided")
	}

	if queryOptions.QueryFile != "" && (queryOptions.KQL+queryOptions.DSL+queryOptions.Lucene != "") {
		return fmt.Errorf("--query-file cannot be used with --kql, --dsl, or --lucene")
	}

	if queryOptions.QueryFile != "" {
		if _, err := os.Stat(queryOptions.QueryFile); os.IsNotExist(err) {
			return fmt.Errorf("query file does not exist: %s", queryOptions.QueryFile)
		}
	}

	if queryOptions.DSL != "" {
		if !json.Valid([]byte(queryOptions.DSL)) {
			return fmt.Errorf("DSL must be a valid JSON string")
		}
	}

	if (queryOptions.KQL != "" && queryOptions.DSL != "") || (queryOptions.KQL != "" && queryOptions.Lucene != "") || (queryOptions.DSL != "" && queryOptions.Lucene != "") {
		return fmt.Errorf("only one of --kql, --dsl, or --lucene can be provided at a time")
	}

	if queryOptions.From != "" && !strings.HasPrefix(queryOptions.From, "now") {
		if _, err := time.Parse(time.RFC3339, queryOptions.From); err != nil {
			return fmt.Errorf("invalid --from timestamp: %v", err)
		}
	}
	if queryOptions.To != "" && !strings.HasPrefix(queryOptions.To, "now") {
		if _, err := time.Parse(time.RFC3339, queryOptions.To); err != nil {
			return fmt.Errorf("invalid --to timestamp: %v", err)
		}
	}

	return nil
}

// ValidateOutputOptions validates the output options.
func ValidateOutputOptions(outputOptions options.OutputOptions) error {
	// check if format is valid
	validOutputs := map[string]bool{"json": true, "text": true}
	if _, ok := validOutputs[outputOptions.Output]; !ok {
		return fmt.Errorf("invalid output format '%s'. Must be one of: %s", outputOptions.Output, strings.Join(getKeys(validOutputs), ", "))
	}

	// check if output file already exists
	if outputOptions.OutputFile != "" {
		if _, err := os.Stat(outputOptions.OutputFile); err == nil {
			return fmt.Errorf("output file already exists: %s", outputOptions.OutputFile)
		}
	}

	if outputOptions.JqPath != "" {
		_, err := gojq.Parse(outputOptions.JqPath)
		if err != nil {
			return fmt.Errorf("invalid jq expression: %s", err)
		}
	}

	return nil
}

// ValidateAuthOptions validates the authentication options.
func ValidateAuthOptions(authOptions options.AuthOptions) error {
	if authOptions.APIKey != "" && (authOptions.Username != "" || authOptions.Password != "") {
		return fmt.Errorf("--api-key cannot be used with --username or --password")
	}

	if authOptions.Password != "" && authOptions.Username == "" {
		return fmt.Errorf("--password must be used with --username")
	}

	return nil
}

// ValidateElasticOptions validates the Elasticsearch options.
func ValidateElasticOptions(elasticOptions options.ElasticOptions) error {
	if elasticOptions.Node == "" {
		return fmt.Errorf("--node must be provided")
	}

	if elasticOptions.Index == "" {
		return fmt.Errorf("--index must be provided")
	}

	return nil
}
