package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/PaesslerAG/jsonpath"
)

func ProcessAndOutputResults(results interface{}, outputFormat, outputFile, jsonPathExpr string) error {
	var formattedOutput []byte
	var err error

	switch outputFormat {
	case "json":
		if jsonPathExpr != "" {
			v, err := jsonpath.JsonPathLookup(results, jsonPathExpr)
			if err != nil {
				return fmt.Errorf("failed to apply JSONPath '%s': %w", jsonPathExpr, err)
			}
			formattedOutput, err = json.MarshalIndent(v, "", "  ")
		} else {
			formattedOutput, err = json.MarshalIndent(results, "", "  ")
		}
		if err != nil {
			return fmt.Errorf("failed to marshal results to JSON: %w", err)
		}
	case "normal":
		formattedOutput = []byte(fmt.Sprintf("Raw results (normal output not fully implemented):\n%v", results))
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	if outputFile != "" {
		err = os.WriteFile(outputFile, formattedOutput, 0644)
		if err != nil {
			return fmt.Errorf("failed to write output to file '%s': %w", outputFile, err)
		}
		fmt.Printf("Output written to %s\n", outputFile)
	} else {
		fmt.Println(string(formattedOutput))
	}

	return nil
}
