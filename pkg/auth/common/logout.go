package common

import (
	"context"

	"github.com/coding-wepack/carctl/pkg/auth"
)

func (c *Client) Logout(ctx context.Context, hostname string) error {
	_, ok := c.config.Auths[hostname]
	if !ok {
		return auth.ErrNotLoggedIn
	}

	return c.config.RemoveAuthConfig(hostname)
}
