package options

import (
	"testing"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/stretchr/testify/assert"
)

func TestAuthOptions_UpdateConfig(t *testing.T) {
	testCases := []struct {
		name         string
		opts         AuthOptions
		wantAPIKey   string
		wantUsername string
		wantPassword string
	}{
		{
			name:       "API Key Auth",
			opts:       AuthOptions{APIKey: "my-secret-key"},
			wantAPIKey: "my-secret-key",
		},
		{
			name:         "Username/Password Auth",
			opts:         AuthOptions{Username: "user", Password: "pw"},
			wantUsername: "user",
			wantPassword: "pw",
		},
		{
			name: "No Auth",
			opts: AuthOptions{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &elasticsearch.Config{}
			tc.opts.UpdateConfig(cfg)

			assert.Equal(t, tc.wantAPIKey, cfg.APIKey)
			assert.Equal(t, tc.wantUsername, cfg.Username)
			assert.Equal(t, tc.wantPassword, cfg.Password)
		})
	}
}
