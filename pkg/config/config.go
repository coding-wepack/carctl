package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/docker/docker/pkg/homedir"
	"github.com/pkg/errors"
)

var (
	initConfigDir = new(sync.Once)
	configDir     string
	homeDir       string
)

const (
	DefaultConfigFileName = "config.json"
	configFileDir         = ".carctl"
	contextsDir           = "contexts"
)

// resetHomeDir is used in testing to reset the "homeDir" package variable to
// force re-lookup of the home directory between tests.
func resetHomeDir() {
	homeDir = ""
}

func GetHomeDir() string {
	if homeDir == "" {
		homeDir = homedir.Get()
	}
	return homeDir
}

// resetConfigDir is used in testing to reset the "configDir" package variable
// and its sync.Once to force re-lookup between tests.
func resetConfigDir() {
	configDir = ""
	initConfigDir = new(sync.Once)
}

func setConfigDir() {
	if configDir != "" {
		return
	}
	configDir = os.Getenv("CARCTL_CONFIG")
	if configDir == "" {
		configDir = filepath.Join(GetHomeDir(), configFileDir)
	}
}

func DefaultConfigFilePath() string {
	return filepath.Join(Dir(), DefaultConfigFileName)
}

// Dir returns the directory the configuration file is stored in
func Dir() string {
	initConfigDir.Do(setConfigDir)
	return configDir
}

// ContextStoreDir returns the directory the docker contexts are stored in
func ContextStoreDir() string {
	return filepath.Join(Dir(), contextsDir)
}

// SetDir sets the directory the configuration file is stored in
func SetDir(dir string) {
	configDir = filepath.Clean(dir)
}

// Path returns the path to a file relative to the config dir
func Path(p ...string) (string, error) {
	path := filepath.Join(append([]string{Dir()}, p...)...)
	if !strings.HasPrefix(path, Dir()+string(filepath.Separator)) {
		return "", errors.Errorf("path %q is outside of root config directory %q", path, Dir())
	}
	return path, nil
}

// LegacyLoadFromReader is a convenience function that creates a ConfigFile object from
// a non-nested reader
func LegacyLoadFromReader(configData io.Reader) (*Config, error) {
	configFile := Config{
		Registry: &Registry{
			AuthConfigs: make(map[string]AuthConfig),
		},
	}
	err := configFile.LegacyLoadFromReader(configData)
	return &configFile, err
}

// LoadFromReader is a convenience function that creates a ConfigFile object from
// a reader
func LoadFromReader(configData io.Reader) (*Config, error) {
	configFile := Config{
		Registry: &Registry{
			AuthConfigs: make(map[string]AuthConfig),
		},
	}
	err := configFile.LoadFromReader(configData)
	return &configFile, err
}

// Load reads the configuration files in the given directory, and sets up
// the auth config information and returns values.
// FIXME: use the internal golang config parser
func Load(configDir string) (*Config, error) {
	if configDir == "" {
		configDir = Dir()
	}

	filename := filepath.Join(configDir, DefaultConfigFileName)
	configFile := New(filename)

	// Try happy path first - latest config file
	if file, err := os.Open(filename); err == nil {
		defer file.Close()
		err = configFile.LoadFromReader(file)
		if err != nil {
			err = errors.Wrap(err, filename)
		}
		return configFile, err
	} else if !os.IsNotExist(err) {
		// if file is there but we can't stat it for any reason other
		// than it doesn't exist then stop
		return configFile, errors.Wrap(err, filename)
	}

	return configFile, nil
}

// LoadDefaultConfigFile attempts to load the default config file and returns
// an initialized ConfigFile struct if none is found.
func LoadDefaultConfigFile(stderr io.Writer) *Config {
	configFile, err := Load(Dir())
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "WARNING: Error loading config file: %v\n", err)
	}
	return configFile
}
