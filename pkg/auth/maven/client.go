package maven

import (
	"os"

	"github.com/pkg/errors"

	"e.coding.net/codingcorp/carctl/pkg/auth"
	"e.coding.net/codingcorp/carctl/pkg/config"
)

// Client provides authentication operations for maven registries.
type Client struct {
	config *config.Config
}

func NewClient(configPath string) (auth.Client, error) {
	if configPath == "" {
		cfg, err := config.Load(config.Dir())
		if err != nil {
			return nil, err
		}

		return &Client{
			config: cfg,
		}, nil
	} else {
		cfg, err := loadConfigFile(configPath)
		if err != nil {
			return nil, errors.Wrap(err, configPath)
		}

		return &Client{
			config: cfg,
		}, nil
	}
}

// loadConfigFile reads the configuration files from the given path.
func loadConfigFile(path string) (*config.Config, error) {
	cfg := config.New(path)
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		if err := cfg.LoadFromReader(file); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return cfg, nil
}
