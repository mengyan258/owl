package conf

// Repository 接口定义
type Repository interface {
	// Has Determine if the given configuration value exists.
	Has(key string) bool

	// Get the specified configuration value.
	Get(key interface{}, defaultVal interface{}) interface{}

	// All Get all of the configuration items for the application.
	All() map[string]interface{}

	// Set a given configuration value.
	Set(key interface{}, value interface{})

	// Prepend a value onto an array configuration value.
	Prepend(key string, value interface{})

	// Push a value onto an array configuration value.
	Push(key string, value interface{})
}
