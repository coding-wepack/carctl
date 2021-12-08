package maven

import (
	"context"

	"e.coding.net/codingcorp/carctl/pkg/auth"
	"e.coding.net/codingcorp/carctl/pkg/config"
)

// Login logs in to a docker registry identified by the hostname.
// Deprecated: use LoginWithOpts
func (c *Client) Login(ctx context.Context, hostname, username, secret string, insecure bool) error {
	settings := &auth.LoginSettings{
		Context:  ctx,
		Hostname: hostname,
		Username: username,
		Secret:   secret,
		Insecure: insecure,
	}

	return c.login(settings)
}

func (c *Client) LoginWithOpts(options ...auth.LoginOption) error {
	settings := &auth.LoginSettings{}
	for _, option := range options {
		option(settings)
	}
	return c.login(settings)
}

func (c *Client) login(settings *auth.LoginSettings) error {
	// TODO: ping to server

	// store to config file
	cred := config.AuthConfig{
		Username:      settings.Username,
		Password:      settings.Secret,
		ServerAddress: settings.Hostname,
	}

	return c.config.StoreAuth(cred)
}
