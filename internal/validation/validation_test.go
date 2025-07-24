package validation

import (
	"os"
	"testing"

	"github.com/fa7ad/esq/internal/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateQueryOptions(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "query-*.json")
	require.NoError(t, err, "Failed to create temp file")
	defer os.Remove(tmpFile.Name())

	testCases := []struct {
		name    string
		opts    options.QueryOptions
		wantErr bool
	}{
		{"Valid KQL", options.QueryOptions{KQL: "user:test"}, false},
		{"Valid DSL", options.QueryOptions{DSL: `{"match_all":{}}`}, false},
		{"No Query Provided", options.QueryOptions{}, true},
		{"Multiple Queries (KQL and DSL)", options.QueryOptions{KQL: "user:test", DSL: `{"match_all":{}}`}, true},
		{"Invalid From Timestamp", options.QueryOptions{KQL: "a", From: "not-a-date"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateQueryOptions(tc.opts)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOutputOptions(t *testing.T) {
	existingFile, err := os.CreateTemp("", "output-*.json")
	require.NoError(t, err, "Failed to create temp file")
	defer os.Remove(existingFile.Name())

	testCases := []struct {
		name    string
		opts    options.OutputOptions
		wantErr bool
	}{
		{"Valid JSON Output", options.OutputOptions{Output: "json"}, false},
		{"Invalid Output Format", options.OutputOptions{Output: "xml"}, true},
		{"Valid JQ Path", options.OutputOptions{Output: "json", JqPath: ".hits"}, false},
		{"Invalid JQ Path", options.OutputOptions{Output: "json", JqPath: "{"}, true},
		{"Output File Already Exists", options.OutputOptions{Output: "json", OutputFile: existingFile.Name()}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOutputOptions(tc.opts)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAuthOptions(t *testing.T) {
	testCases := []struct {
		name    string
		opts    options.AuthOptions
		wantErr bool
	}{
		{"API Key Only", options.AuthOptions{APIKey: "key"}, false},
		{"Username/Password", options.AuthOptions{Username: "user", Password: "pw"}, false},
		{"API Key with Username", options.AuthOptions{APIKey: "key", Username: "user"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAuthOptions(tc.opts)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateElasticOptions(t *testing.T) {
	testCases := []struct {
		name    string
		opts    options.ElasticOptions
		wantErr bool
	}{
		{"Valid", options.ElasticOptions{Node: "url", Index: "idx"}, false},
		{"No Node", options.ElasticOptions{Index: "idx"}, true},
		{"No Index", options.ElasticOptions{Node: "url"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateElasticOptions(tc.opts)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
