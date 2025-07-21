package validation

import (
	"fmt"
	"os"
	"strings"

	"github.com/fa7ad/esq/internal/types" // Import CliArgs from types
)

func ValidateCliArgs(args types.CliArgs) error {
	querySources := []string{}
	for i, source := range []bool{
		args.KQL != "",
		args.DSL != "",
		args.Lucene != "",
		args.QueryFile != "",
	} {
		if source {
			querySources = append(querySources, []string{"kql", "dsl", "lucene", "query-file"}[i])
		}
	}

	if len(querySources) == 0 {
		return fmt.Errorf("error: One of --kql, --dsl, --lucene, or --query-file must be provided")
	}
	if len(querySources) > 1 {
		return fmt.Errorf("error: Only one query source (%s) can be used at a time", strings.Join(querySources, ", "))
	}

	if args.Password != "" && args.Username == "" {
		return fmt.Errorf("error: --username must be provided if --password is used")
	}

	if args.QueryFile != "" {
		if _, err := os.Stat(args.QueryFile); os.IsNotExist(err) {
			return fmt.Errorf("error: Query file not found at: %s", args.QueryFile)
		} else if err != nil {
			return fmt.Errorf("error checking query file '%s': %w", args.QueryFile, err)
		}
	}

	validOutputs := map[string]bool{"json": true, "text": true}
	if _, ok := validOutputs[args.Output]; !ok {
		return fmt.Errorf("error: Invalid output format '%s'. Must be one of: %s", args.Output, strings.Join(getKeys(validOutputs), ", "))
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
