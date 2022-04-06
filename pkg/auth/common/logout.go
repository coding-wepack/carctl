package common

import (
	"context"

	"e.coding.net/codingcorp/carctl/pkg/auth"
)

func (c *Client) Logout(ctx context.Context, hostname string) error {
	_, ok := c.config.Auths[hostname]
	if !ok {
		return auth.ErrNotLoggedIn
	}

	return c.config.RemoveAuthConfig(hostname)
}
