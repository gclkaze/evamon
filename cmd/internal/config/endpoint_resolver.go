package config

import (
	"fmt"

	"github.com/magiconair/properties"
	"github.com/spf13/viper"
)

type Endpoint struct {
	Hostname string
	Port     int
}

// EndpointResolver resolves hostname/port either from application.properties (*properties.Properties)
// or from viper flags/config. It prefers properties when provided.
type EndpointResolver struct {
	props *properties.Properties

	// Keys in application.properties
	hostKey string
	portKey string

	// Keys in viper
	viperHostKey string
	viperPortKey string

	defaultHost string
	defaultPort int
}

func NewEndpointResolver(props *properties.Properties) *EndpointResolver {
	return &EndpointResolver{
		props: props,

		// You can change these keys to match your application.properties schema
		hostKey: "hostname",
		portKey: "port",

		// And these to match your viper flag binding
		viperHostKey: "hostname",
		viperPortKey: "port",

		defaultHost: "127.0.0.1",
		defaultPort: 0, // require explicit port
	}
}

// Optionally allow customizing keys (if you use server.hostname/server.port).
func (r *EndpointResolver) WithKeys(propsHostKey, propsPortKey, viperHostKey, viperPortKey string) *EndpointResolver {
	r.hostKey = propsHostKey
	r.portKey = propsPortKey
	r.viperHostKey = viperHostKey
	r.viperPortKey = viperPortKey
	return r
}

func (r *EndpointResolver) Resolve() (Endpoint, error) {
	host := ""
	port := 0

	// Prefer properties if provided
	if r.props != nil {
		host = r.props.GetString(r.hostKey, "")
		port = r.props.GetInt(r.portKey, 0)
	}

	// Fallback to viper
	if host == "" {
		host = viper.GetString(r.viperHostKey)
	}
	if port == 0 {
		port = viper.GetInt(r.viperPortKey)
	}

	// Defaults
	if host == "" {
		host = r.defaultHost
	}
	if port == 0 {
		if r.defaultPort != 0 {
			port = r.defaultPort
		} else {
			return Endpoint{}, fmt.Errorf("port is not configured (checked properties key %q and viper key %q)", r.portKey, r.viperPortKey)
		}
	}

	return Endpoint{Hostname: host, Port: port}, nil
}

func (e Endpoint) WSURL() string {
	return fmt.Sprintf("ws://%s:%d/ws", e.Hostname, e.Port)
}
