package auth

import (
	"context"
)

type (
	// LoginOption allows specifying various settings on login.
	LoginOption func(*LoginSettings)

	// LoginSettings represent all the various settings on login.
	LoginSettings struct {
		Context   context.Context
		Hostname  string
		Username  string
		Secret    string
		Insecure  bool
		UserAgent string
	}
)

// WithLoginContext returns a function that sets the Context setting on login.
func WithLoginContext(context context.Context) LoginOption {
	return func(settings *LoginSettings) {
		settings.Context = context
	}
}

// WithLoginHostname returns a function that sets the Hostname setting on login.
func WithLoginHostname(hostname string) LoginOption {
	return func(settings *LoginSettings) {
		settings.Hostname = hostname
	}
}

// WithLoginUsername returns a function that sets the Username setting on login.
func WithLoginUsername(username string) LoginOption {
	return func(settings *LoginSettings) {
		settings.Username = username
	}
}

// WithLoginSecret returns a function that sets the Secret setting on login.
func WithLoginSecret(secret string) LoginOption {
	return func(settings *LoginSettings) {
		settings.Secret = secret
	}
}

// WithLoginInsecure returns a function that sets the Insecure setting to true on login.
func WithLoginInsecure() LoginOption {
	return func(settings *LoginSettings) {
		settings.Insecure = true
	}
}

// WithLoginUserAgent returns a function that sets the UserAgent setting on login.
func WithLoginUserAgent(userAgent string) LoginOption {
	return func(settings *LoginSettings) {
		settings.UserAgent = userAgent
	}
}
