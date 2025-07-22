package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/itchyny/gojq"
)

// ApplyJQ runs a jq expression against the given input data using gojq.
func ApplyJQ(input any, jqExpr string) (any, error) {
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

func SerializeResults(results any, format string) ([]byte, error) {
	switch format {
	case "json":
		return json.MarshalIndent(results, "", "  ")
	case "text":
		return fmt.Appendf(nil, "%v", results), nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}

func WriteToFile(serialized []byte, outputFile string) error {
	switch outputFile {
	case "*stdout":
		_, err := os.Stdout.Write(serialized)
		if err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
	default:
		err := os.WriteFile(outputFile, serialized, 0644)
		if err != nil {
			return fmt.Errorf("failed to write output to file '%s': %w", outputFile, err)
		}
	}
	return nil
}
