package login

import (
	"io"

	"e.coding.net/codingcorp/carctl/pkg/action"
	"e.coding.net/codingcorp/carctl/pkg/registry"
)

type RegistryLogin struct {
	cfg *action.Configuration
}

func NewRegistryLogin(cfg *action.Configuration) *RegistryLogin {
	return &RegistryLogin{
		cfg: cfg,
	}
}

func (r *RegistryLogin) Run(out io.Writer, host, username, password string, insecure bool) error {
	return r.cfg.RegistryClient.Login(
		host,
		registry.LoginOptBasicAuth(username, password),
		registry.LoginOptInsecure(insecure),
	)
}
