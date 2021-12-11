package action

import (
	"io"

	"e.coding.net/codingcorp/carctl/pkg/registry"
)

type RegistryLogin struct {
	cfg *Configuration
}

func NewRegistryLogin(cfg *Configuration) *RegistryLogin {
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
