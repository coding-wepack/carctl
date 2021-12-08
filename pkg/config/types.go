package config

type Config struct {
	Filename string    `json:"-"` // Note: for internal use only
	Registry *Registry `json:"registry,omitempty"`
	Proxy    *Proxy    `json:"proxy,omitempty"`
}

type Registry struct {
	AuthConfigs map[string]AuthConfig `json:"auths,omitempty"`
}

// AuthConfig contains authorization information for connecting to a Registry
type AuthConfig struct {
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	Auth          string `json:"auth,omitempty"`
	ServerAddress string `json:"serveraddress,omitempty"`
}

type Proxy struct {
	Default ProxyConfig `json:"default,omitempty"`
}

type ProxyConfig struct {
	HttpProxy  string `json:"httpProxy,omitempty"`
	HttpsProxy string `json:"httpsProxy,omitempty"`
	NoProxy    string `json:"noProxy,omitempty"`
	FtpProxy   string `json:"ftpProxy,omitempty"`
}
