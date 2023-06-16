package action

import (
	"github.com/coding-wepack/carctl/pkg/registry"
)

type Configuration struct {
	// RegistryClient is a client for working with registries
	RegistryClient *registry.Client
}
