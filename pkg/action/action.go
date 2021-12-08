package action

import (
	"e.coding.net/codingcorp/carctl/pkg/registry"
)

type Configuration struct {
	// RegistryClient is a client for working with registries
	RegistryClient *registry.Client
}
