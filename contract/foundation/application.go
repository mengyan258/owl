package foundation

import (
	"context"
	"go.uber.org/dig"
)

// Application 接口定义
type Application interface {
	Version() string

	GetBasePath() string
	GetBootstrapPath() string
	GetConfigPath() string
	GetLangPath() string
	GetPublicPath() string
	GetResourcePath() string
	GetStoragePath() string

	Environment(...string) (string, bool)
	MaintenanceMode() context.Context
	IsDownForMaintenance() bool

	// Register a service provider with the application.
	Register(provider ...any)
	Provide(function interface{}, opts ...dig.ProvideOption) error
	Invoke(function interface{}, opts ...dig.InvokeOption) error

	// Boot the application's service providers.
	Boot()

	// Register a new boot listener.
	Booting(callback func(application Application))

	// Register a new "booted" listener.
	Booted(callback func(application Application))

	// Get the current application locale.
	GetLocale() string

	// Get the registered service provider instances if any exist.
	GetProviders(provider interface{}) []interface{}

	// Determine if the application has been bootstrapped before.
	HasBeenBootstrapped() bool

	// Set the current application locale.
	SetLocale(locale string)

	// Register a terminating callback with the application.
	Terminating(callback interface{}) Application

	// Terminate the application.
	Terminate()
}
