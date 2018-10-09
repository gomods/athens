package config

// ProxyConfig specifies the properties required to run the proxy
type ProxyConfig struct {
	StorageType           string `validate:"required" envconfig:"ATHENS_STORAGE_TYPE"`
	OlympusGlobalEndpoint string `envconfig:"OLYMPUS_GLOBAL_ENDPOINT"`
	Port                  string `validate:"required" envconfig:"PORT"`
	FilterOff             bool   `validate:"required" envconfig:"PROXY_FILTER_OFF"`
	BasicAuthUser         string `envconfig:"BASIC_AUTH_USER"`
	BasicAuthPass         string `envconfig:"BASIC_AUTH_PASS"`
	ForceSSL              bool   `envconfig:"PROXY_FORCE_SSL"`
	ValidatorHook         string `envconfig:"ATHENS_PROXY_VALIDATOR"`
	PathPrefix            string `envconfig:"ATHENS_PATH_PREFIX"`
	NETRCPath             string `envconfig:"ATHENS_NETRC_PATH"`
	GithubToken           string `envconfig:"ATHENS_GITHUB_TOKEN"`
	HGRCPath              string `envconfig:"ATHENS_HGRC_PATH"`
}

// BasicAuth returns BasicAuthUser and BasicAuthPassword
// and ok if neither of them are empty
func (p *ProxyConfig) BasicAuth() (user, pass string, ok bool) {
	user = p.BasicAuthUser
	pass = p.BasicAuthPass
	ok = user != "" && pass != ""
	return user, pass, ok
}
