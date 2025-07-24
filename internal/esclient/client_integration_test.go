//go:build integration

package esclient

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Constants for the test setup.
const (
	esURL     = "http://localhost:9200"
	esIndex   = "makelogs-*"
	esUser    = "elastic"
	esqBinary = "../../esq"
	queryFile = "../../test_query.json"
)

func TestIntegrationESQ(t *testing.T) {
	esPassword := os.Getenv("ES_LOCAL_PASSWORD")
	require.NotEmpty(t, esPassword, "ES_LOCAL_PASSWORD environment variable is not set")

	// Cleanup the query file in case a previous run failed.
	defer os.Remove(queryFile)

	baseArgs := []string{
		"-n", esURL,
		"-i", esIndex,
		"--username", esUser,
		"--password", esPassword,
	}

	t.Run("KQL query with basic auth", func(t *testing.T) {
		args := append(baseArgs, "--kql", "response: 200", "--size", "1")
		cmd := exec.Command(esqBinary, args...)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Command failed. Output: %s", output)
		assert.Contains(t, string(output), "200", "Expected output to contain '200'")
	})

	t.Run("DSL query from file", func(t *testing.T) {
		dslQuery := `{"query":{"match":{"agent":"Mozilla"}}}`
		err := os.WriteFile(queryFile, []byte(dslQuery), 0644)
		require.NoError(t, err, "Failed to write query file")

		args := append(baseArgs, "-f", queryFile, "--size", "1")
		cmd := exec.Command(esqBinary, args...)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Command failed. Output: %s", output)
		assert.Contains(t, string(output), "Mozilla", "Expected output to contain 'Mozilla'")
	})

	t.Run("JSON output with jq processing", func(t *testing.T) {
		args := append(baseArgs, "--kql", "machine.os: osx", "-o", "json", "-j", ".hits[0]._source.machine.os", "--size", "1")
		cmd := exec.Command(esqBinary, args...)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Command failed. Output: %s", output)
		assert.Contains(t, string(output), "osx", "Expected output to contain 'osx'")
	})

	t.Run("Invalid query returns non-zero exit code", func(t *testing.T) {
		// FIX: Use a syntactically invalid query to force a real error from Elasticsearch.
		args := append(baseArgs, "--kql", "field: (unclosed")
		cmd := exec.Command(esqBinary, args...)
		err := cmd.Run()
		assert.Error(t, err, "Expected command to fail with an exit code, but it succeeded")
	})
}
