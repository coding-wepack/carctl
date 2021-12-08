package auth

import (
	"context"

	"github.com/pkg/errors"
)

// Common errors
var (
	ErrNotLoggedIn = errors.New("not logged in")
)

// Client provides authentication operations for remotes.
type Client interface {
	// Login logs in to a remote server identified by the hostname.
	// Deprecated: use LoginWithOpts
	Login(ctx context.Context, hostname, username, secret string, insecure bool) error
	// LoginWithOpts logs in to a remote server identified by the hostname with custom options
	LoginWithOpts(options ...LoginOption) error
	// Logout logs out from a remote server identified by the hostname.
	Logout(ctx context.Context, hostname string) error
	// Resolver returns a new authenticated resolver.
	// Deprecated: use ResolverWithOpts
	// Resolver(ctx context.Context, client *http.Client, plainHTTP bool) (remotes.Resolver, error)
	// ResolverWithOpts returns a new authenticated resolver with custom options.
	// ResolverWithOpts(options ...ResolverOption) (remotes.Resolver, error)
}
