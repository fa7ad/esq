package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fa7ad/esq/internal/options"
	"github.com/itchyny/gojq"
)

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

func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

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

	return nil
}

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

func ValidateAuthOptions(authOptions options.AuthOptions) error {
	if authOptions.APIKey != "" && (authOptions.Username != "" || authOptions.Password != "") {
		return fmt.Errorf("--api-key cannot be used with --username or --password")
	}

	if authOptions.Password != "" && authOptions.Username == "" {
		return fmt.Errorf("--password must be used with --username")
	}

	return nil
}

func ValidateElasticOptions(elasticOptions options.ElasticOptions) error {
	if elasticOptions.Node == "" {
		return fmt.Errorf("--node must be provided")
	}

	if elasticOptions.Index == "" {
		return fmt.Errorf("--index must be provided")
	}

	return nil
}
