package config

import "net/url"

type CDNConfig struct {
	Endpoint       string
	TimeoutSeconds int `validate:"required"`
}

// CDNEndpointWithDefault returns CDN endpoint if set
// if not it should default to clouds default blob storage endpoint e.g
func (c *CDNConfig) CDNEndpointWithDefault(value *url.URL) *url.URL {
	if c.Endpoint == "" {
		return value
	}
	rawURI := c.Endpoint

	uri, err := url.Parse(rawURI)
	if err != nil {
		return value
	}
	return uri
}
