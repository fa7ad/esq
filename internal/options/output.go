package options

import (
	"fmt"

	"github.com/fa7ad/esq/internal/output"
)

// OutputOptions holds output-related fields.
// It includes output format, output file, and jq path.
type OutputOptions struct {
	Output     string
	OutputFile string
	JqPath     string
}

func (o *OutputOptions) processResults(results any) (any, error) {
	if o.JqPath == "" {
		return results, nil
	}
	parsed, err := output.ApplyJQ(results, o.JqPath)
	if err != nil {
		return nil, fmt.Errorf("failed to apply jq: %w", err)
	}
	return parsed, nil
}

func (o *OutputOptions) OutputResults(results any) error {
	processed, err := o.processResults(results)
	if err != nil {
		return err
	}

	// now serialize to the specified format
	serialized, err := output.SerializeResults(processed, o.Output)
	if err != nil {
		return fmt.Errorf("failed to serialize results: %w", err)
	}

	outputFile := o.OutputFile
	if outputFile == "" {
		outputFile = "*stdout"
	}
	return output.WriteToFile(serialized, outputFile)
}
