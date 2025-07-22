package options

import "github.com/elastic/go-elasticsearch/v9"

// AuthOptions holds authentication-related fields.
type AuthOptions struct {
	APIKey   string
	Username string
	Password string
}

func (a *AuthOptions) UpdateConfig(receiver *elasticsearch.Config) {
	if a.APIKey != "" {
		receiver.APIKey = a.APIKey
	} else if a.Username != "" && a.Password != "" {
		receiver.Username = a.Username
		receiver.Password = a.Password
	}
}
