package registry

import (
	"context"
	"fmt"
	"io"

	"e.coding.net/codingcorp/carctl/pkg/auth/common"

	"e.coding.net/codingcorp/carctl/pkg/auth"
	"e.coding.net/codingcorp/carctl/pkg/config"
	"e.coding.net/codingcorp/carctl/pkg/constants"
	"e.coding.net/codingcorp/carctl/pkg/util/artutil"
	"github.com/pkg/errors"
)

type Client struct {
	// configFile is path to registry config file.
	// e.g., $HOME/.carctl/config.json
	configFile string

	verbose bool

	out io.Writer

	authorizer auth.Client
}

// ClientOption allows specifying various settings configurable by the user for overriding the defaults
// used when creating a new default client
type ClientOption func(*Client)

// NewClient returns a new registry client with config
func NewClient(options ...ClientOption) (*Client, error) {
	client := &Client{}
	for _, option := range options {
		option(client)
	}
	if client.configFile == "" {
		client.configFile = config.DefaultConfigFilePath()
	}

	return client, nil
}

func (c *Client) ConfigFilePath() string {
	return c.configFile
}

func (c *Client) ConfigFile() (*config.Config, error) {
	return config.LoadConfigFile(c.configFile)
}

// ClientOptVerbose returns a function that sets the debug setting on client options set
func ClientOptVerbose(verbose bool) ClientOption {
	return func(client *Client) {
		client.verbose = verbose
	}
}

// ClientOptWriter returns a function that sets the writer setting on client options set
func ClientOptWriter(out io.Writer) ClientOption {
	return func(client *Client) {
		client.out = out
	}
}

// ClientOptConfigFile returns a function that sets the credentialsFile setting on a client options set
func ClientOptConfigFile(configFile string) ClientOption {
	return func(client *Client) {
		client.configFile = configFile
	}
}

type (
	// LoginOption allows specifying various settings on login
	LoginOption func(*loginOperation)

	loginOperation struct {
		username string
		password string
		insecure bool
	}
)

// Login logs into a registry
func (c *Client) Login(host string, options ...LoginOption) error {
	operation := &loginOperation{}
	for _, option := range options {
		option(operation)
	}

	authorizerLoginOpts := []auth.LoginOption{
		// auth.WithLoginContext(ctx(c.out, c.debug)),
		auth.WithLoginContext(context.Background()),
		auth.WithLoginHostname(host),
		auth.WithLoginUsername(operation.username),
		auth.WithLoginSecret(operation.password),
		// auth.WithLoginUserAgent(version.GetUserAgent()),
	}
	if operation.insecure {
		authorizerLoginOpts = append(authorizerLoginOpts, auth.WithLoginInsecure())
	}

	var err error
	if c.authorizer == nil {
		c.authorizer, err = getAuthorizer(host, c.configFile)
		if err != nil {
			return err
		}
	}

	if err := c.authorizer.LoginWithOpts(authorizerLoginOpts...); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(c.out, "Login Succeeded")

	return nil
}

// LoginOptBasicAuth returns a function that sets the username/password settings on login
func LoginOptBasicAuth(username string, password string) LoginOption {
	return func(operation *loginOperation) {
		operation.username = username
		operation.password = password
	}
}

// LoginOptInsecure returns a function that sets the insecure setting on login
func LoginOptInsecure(insecure bool) LoginOption {
	return func(operation *loginOperation) {
		operation.insecure = insecure
	}
}

type (
	// LogoutOption allows specifying various settings on logout
	LogoutOption func(*logoutOperation)

	logoutOperation struct{}
)

// Logout logs out of a registry
func (c *Client) Logout(host string, opts ...LogoutOption) error {
	operation := &logoutOperation{}
	for _, opt := range opts {
		opt(operation)
	}

	var err error
	if c.authorizer == nil {
		c.authorizer, err = getAuthorizer(host, c.configFile)
		if err != nil {
			return err
		}
	}

	if err = c.authorizer.Logout(context.Background(), host); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(c.out, "Removing login credentials for %s\n", host)

	return nil
}

func getAuthorizer(host, configFilePath string) (auth.Client, error) {
	artType, err := artutil.ArtifactTypeFromHost(host)
	if err != nil {
		return nil, err
	}

	switch artType {
	case constants.TypeGeneric, constants.TypeDocker, constants.TypeMaven, constants.TypeNpm, constants.TypeComposer, constants.TypePypi:
		return common.NewClient(configFilePath)
	default:
		return nil, errors.New("unsupported artifact type")
	}
}
