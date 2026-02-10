package services

import (
	"github.com/gclkaze/evamon/cmd/internal/config"
	"github.com/gclkaze/evamon/cmd/internal/output"
	"github.com/gclkaze/evamon/cmd/internal/wsclient"
	"github.com/magiconair/properties"
)

type MainSetup interface {
	GetEndpointResolver() *config.EndpointResolver
	GetToken() string
	GetProperties() *properties.Properties
	GetPrinter() output.Printer
	GetWSClient() *wsclient.Client
	GetWidgetPath() string
}
