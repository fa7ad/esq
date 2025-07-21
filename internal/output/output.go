package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/itchyny/gojq"
)

// ProcessAndOutputResults processes the Elasticsearch results, applies optional jq filtering,
// and writes to stdout or to the specified output file.
func ProcessAndOutputResults(results any, format string, outputFile string, jqExpr string) error {
	rawJson, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal results to JSON: %w", err)
	}

	var parsed any
	if err := json.Unmarshal(rawJson, &parsed); err != nil {
		return fmt.Errorf("failed to unmarshal JSON for jq processing: %w", err)
	}

	// Apply jq expression if provided
	if jqExpr != "" {
		parsed, err = applyJQ(parsed, jqExpr)
		if err != nil {
			return fmt.Errorf("failed to apply jq: %w", err)
		}
	}

	var output []byte

	switch format {
	case "json", "":
		output, err = json.MarshalIndent(parsed, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal final output: %w", err)
		}

	case "text":
		output = fmt.Append(nil, parsed)

	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	if outputFile != "" {
		err := os.WriteFile(outputFile, output, 0644)
		if err != nil {
			return fmt.Errorf("failed to write output to file: %w", err)
		}
	} else {
		fmt.Println(string(output))
	}

	return nil
}

// applyJQ runs a jq expression against the given input data using gojq.
func applyJQ(input any, jqExpr string) (any, error) {
	query, err := gojq.Parse(jqExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid jq expression: %w", err)
	}

	iter := query.Run(input)

	var results []interface{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := v.(error); isErr {
			return nil, err
		}
		results = append(results, v)
	}

	// Return simplified structure if only one result
	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}
