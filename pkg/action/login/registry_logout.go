package login

import (
	"io"

	"e.coding.net/codingcorp/carctl/pkg/action"
)

// RegistryLogout performs a registry login operation.
type RegistryLogout struct {
	cfg *action.Configuration
}

// NewRegistryLogout creates a new RegistryLogout object with the given configuration.
func NewRegistryLogout(cfg *action.Configuration) *RegistryLogout {
	return &RegistryLogout{
		cfg: cfg,
	}
}

// Run executes the registry logout operation
func (r *RegistryLogout) Run(out io.Writer, hostname string) error {
	return r.cfg.RegistryClient.Logout(hostname)
}
