// Package configloader provides type-safe configuration loading with support
// for environment variables and configuration files. It uses Go generics
// to provide compile-time type safety for configuration structures.
package configloader

// New creates and loads a configuration of type T using the provided options.
// By default, it loads configuration from environment variables.
// Additional sources can be added using options like WithFile().
//
// Example:
//
//	type Config struct {
//	    Port int `default:"8080"`
//	    Host string `default:"localhost"`
//	}
//
//	cfg, err := configloader.New[Config]()
//	if err != nil {
//	    log.Fatal(err)
//	}
func New[T any](opts ...Option) (*T, error) {
	var cfg T
	ops := []Option{
		WithEnv(""),
	}
	ops = append(ops, opts...)

	parser, err := NewParser(ops...)
	if err != nil {
		return nil, err
	}

	err = parser.Parse(&cfg)

	return &cfg, err
}
