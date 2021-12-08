package config

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"e.coding.net/codingcorp/carctl/pkg/log"
	"e.coding.net/codingcorp/carctl/pkg/log/logfields"
	"e.coding.net/codingcorp/carctl/pkg/util/jsonutil"
)

func New(fn string) *Config {
	return &Config{
		Filename: fn,
		Registry: &Registry{
			AuthConfigs: map[string]AuthConfig{},
		},
	}
}

// LoadFromReader reads the configuration data given and sets up the auth config
// information with given directory and populates the receiver object
func (c *Config) LoadFromReader(r io.Reader) error {
	var err error
	if err = jsonutil.NewDecoder(r).Decode(&c); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	for addr, ac := range c.Registry.AuthConfigs {
		if ac.Auth != "" {
			ac.Username, ac.Password, err = decodeAuth(ac.Auth)
			if err != nil {
				return err
			}
		}
		ac.Auth = ""
		ac.ServerAddress = addr
		c.Registry.AuthConfigs[addr] = ac
	}
	return nil
}

// LegacyLoadFromReader reads the non-nested configuration data given and sets up the
// auth config information with given directory and populates the receiver object
func (c *Config) LegacyLoadFromReader(configData io.Reader) error {
	b, err := ioutil.ReadAll(configData)
	if err != nil {
		return err
	}

	if err := jsonutil.Unmarshal(b, &c.Registry.AuthConfigs); err != nil {
		arr := strings.Split(string(b), "\n")
		if len(arr) < 2 {
			return errors.Errorf("The Auth config file is empty")
		}
		authConfig := AuthConfig{}
		origAuth := strings.Split(arr[0], " = ")
		if len(origAuth) != 2 {
			return errors.Errorf("Invalid Auth config file")
		}
		authConfig.Username, authConfig.Password, err = decodeAuth(origAuth[1])
		if err != nil {
			return err
		}
		// c.Registry.AuthConfigs[defaultIndexServer] = authConfig
	} else {
		for k, authConfig := range c.Registry.AuthConfigs {
			authConfig.Username, authConfig.Password, err = decodeAuth(authConfig.Auth)
			if err != nil {
				return err
			}
			authConfig.Auth = ""
			authConfig.ServerAddress = k
			c.Registry.AuthConfigs[k] = authConfig
		}
	}
	return nil
}

// GetFilename returns the file name that this config file is based on.
func (c *Config) GetFilename() string {
	return c.Filename
}

// ContainsAuth returns whether there is authentication configured
// in this file or not.
// func (c *Config) ContainsAuth() bool {
// 	return len(c.Registry.AuthConfigs) > 0
// }

// GetAllAuthConfigs returns the mapping of repo to auth configuration
func (c *Config) GetAllAuthConfigs() map[string]AuthConfig {
	return c.Registry.AuthConfigs
}

// GetAuthConfig for a repository from the credential store
func (c *Config) GetAuthConfig(serverAddress string) (AuthConfig, error) {
	authConfig, ok := c.Registry.AuthConfigs[serverAddress]
	if !ok {
		// Maybe they have a legacy config file, we will iterate the keys converting
		// them to the new format and testing
		for r, ac := range c.Registry.AuthConfigs {
			if serverAddress == ConvertToHostname(r) {
				return ac, nil
			}
		}

		authConfig = AuthConfig{}
	}
	return authConfig, nil
}

func (c *Config) StoreAuth(authConfig AuthConfig) error {
	c.Registry.AuthConfigs[authConfig.ServerAddress] = authConfig
	return c.Save()
}

func (c *Config) RemoveAuthConfig(serverAddress string) error {
	delete(c.Registry.AuthConfigs, serverAddress)
	return c.Save()
}

// SaveToWriter encodes and writes out all the authorization information to
// the given writer
func (c *Config) SaveToWriter(w io.Writer) error {
	// Encode sensitive data into a new/temp struct
	tmpAuthConfigs := make(map[string]AuthConfig, len(c.Registry.AuthConfigs))
	for k, authConfig := range c.Registry.AuthConfigs {
		authCopy := authConfig
		// encode and save the authstring, while blanking out the original fields
		authCopy.Auth = encodeAuth(&authCopy)
		authCopy.Username = ""
		authCopy.Password = ""
		authCopy.ServerAddress = ""
		tmpAuthConfigs[k] = authCopy
	}

	saveAuthConfigs := c.Registry.AuthConfigs
	c.Registry.AuthConfigs = tmpAuthConfigs
	defer func() { c.Registry.AuthConfigs = saveAuthConfigs }()

	// data, err := jsonutil.MarshalIndent(c, "", "    ")
	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// Save encodes and writes out all the authorization information
func (c *Config) Save() (retErr error) {
	if c.Filename == "" {
		return errors.Errorf("Can't save config with empty filename")
	}

	dir := filepath.Dir(c.Filename)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	temp, err := ioutil.TempFile(dir, filepath.Base(c.Filename))
	if err != nil {
		return err
	}
	defer func() {
		_ = temp.Close()
		if retErr != nil {
			if err := os.Remove(temp.Name()); err != nil {
				log.Debug("Error cleaning up temp file",
					logfields.String("file", temp.Name()),
					logfields.String("error", err.Error()))
			}
		}
	}()

	err = c.SaveToWriter(temp)
	if err != nil {
		return err
	}

	if err := temp.Close(); err != nil {
		return errors.Wrap(err, "error closing temp file")
	}

	// Handle situation where the configfile is a symlink
	cfgFile := c.Filename
	if f, err := os.Readlink(cfgFile); err == nil {
		cfgFile = f
	}

	// Try copying the current config file (if any) ownership and permissions
	copyFilePermissions(cfgFile, temp.Name())
	return os.Rename(temp.Name(), cfgFile)
}

// ParseProxyConfig computes proxy configuration by retrieving the config for the provided host and
// then checking this against any environment variables provided to the container
func (c *Config) ParseProxyConfig(host string, runOpts map[string]*string) map[string]*string {
	// TODO: implement me
	return nil
}

// encodeAuth creates a base64 encoded string to containing authorization information
func encodeAuth(authConfig *AuthConfig) string {
	if authConfig.Username == "" && authConfig.Password == "" {
		return ""
	}

	authStr := authConfig.Username + ":" + authConfig.Password
	msg := []byte(authStr)
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(msg)))
	base64.StdEncoding.Encode(encoded, msg)
	return string(encoded)
}

// decodeAuth decodes a base64 encoded string and returns username and password
func decodeAuth(authStr string) (string, string, error) {
	if authStr == "" {
		return "", "", nil
	}

	decLen := base64.StdEncoding.DecodedLen(len(authStr))
	decoded := make([]byte, decLen)
	authByte := []byte(authStr)
	n, err := base64.StdEncoding.Decode(decoded, authByte)
	if err != nil {
		return "", "", err
	}
	if n > decLen {
		return "", "", errors.Errorf("Something went wrong decoding auth config")
	}
	arr := strings.SplitN(string(decoded), ":", 2)
	if len(arr) != 2 {
		return "", "", errors.Errorf("Invalid auth configuration file")
	}
	password := strings.Trim(arr[1], "\x00")
	return arr[0], password, nil
}

// ConvertToHostname converts a registry url which has http|https prepended
// to just an hostname.
// Copied from github.com/docker/docker/registry.ConvertToHostname to reduce dependencies.
func ConvertToHostname(url string) string {
	stripped := url
	if strings.HasPrefix(url, "http://") {
		stripped = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		stripped = strings.TrimPrefix(url, "https://")
	}

	nameParts := strings.SplitN(stripped, "/", 2)

	return nameParts[0]
}
