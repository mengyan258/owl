package foundation

import (
	"context"
	"go.uber.org/dig"
)

// Application 接口定义
type Application interface {
	// Get the version number of the application.
	Version() string

	// Get the base path of the Laravel installation.
	BasePath(path string) string

	// Get the path to the bootstrap directory.
	BootstrapPath(path string) string

	// Get the path to the application configuration files.
	ConfigPath(path string) string

	// Get the path to the database directory.
	DatabasePath(path string) string

	// Get the path to the language files.
	LangPath(path string) string

	// Get the path to the public directory.
	PublicPath(path string) string

	// Get the path to the resources directory.
	ResourcePath(path string) string

	// Get the path to the storage directory.
	StoragePath(path string) string

	// Get or check the current application environment.
	Environment(...string) (string, bool)

	// Determine if the application is running in the console.
	RunningInConsole() bool

	// Determine if the application is running unit tests.
	RunningUnitTests() bool

	// Determine if the application is running with debug mode enabled.
	HasDebugModeEnabled() bool

	// Get an instance of the maintenance mode manager implementation.
	MaintenanceMode() context.Context

	// Determine if the application is currently down for maintenance.
	IsDownForMaintenance() bool

	// Register a service provider with the application.
	Register(provider ...any)
	Provide(function interface{}, opts ...dig.ProvideOption) error
	Invoke(function interface{}, opts ...dig.InvokeOption) error
	RegisterServiceProviders(providers ...ServiceProvider)
	// Register a deferred provider and service.
	RegisterDeferredProvider(provider string, service *string)

	// Resolve a service provider instance from the class name.
	ResolveProvider(provider string) interface{}

	// Boot the application's service providers.
	Boot()

	// Register a new boot listener.
	Booting(callback func())

	// Register a new "booted" listener.
	Booted(callback func())

	// Run the given array of bootstrap classes.
	BootstrapWith(bootstrappers []string)

	// Get the current application locale.
	GetLocale() string

	// Get the application namespace.
	GetNamespace() string

	// Get the registered service provider instances if any exist.
	GetProviders(provider interface{}) []interface{}

	// Determine if the application has been bootstrapped before.
	HasBeenBootstrapped() bool

	// Load and boot all of the remaining deferred providers.
	LoadDeferredProviders()

	// Set the current application locale.
	SetLocale(locale string)

	// Determine if middleware has been disabled for the application.
	ShouldSkipMiddleware() bool

	// Register a terminating callback with the application.
	Terminating(callback interface{}) Application

	// Terminate the application.
	Terminate()
}
