package configloader

import (
	"fmt"
	"os"

	"github.com/katalabut/fast-app/configloader/source"
)

const (
	envFilePath = "CONFIG_FILE"
)

// Option is a functional option for configuring the configuration parser.
type Option func(*Parser) error

// WithEnv adds environment variable support to the configuration loader.
// If prefix is provided, only environment variables with that prefix will be considered.
// For example, WithEnv("APP_") will look for variables like APP_PORT, APP_HOST, etc.
func WithEnv(prefix string) Option {
	return func(p *Parser) error {
		src := source.NewEnv(prefix)

		return p.SetSource(src)
	}
}

// WithFile adds configuration file support to the configuration loader.
// It accepts multiple file paths and will use the first existing file.
// Supported formats include JSON, YAML, TOML, and other formats supported by Viper.
func WithFile(paths ...string) Option {
	return func(p *Parser) error {
		src, err := source.NewFile(paths...)
		if err != nil {
			return err
		}
		return p.SetSource(src)
	}
}

// WithFileFromEnv adds configuration file support with the file path taken from
// the CONFIG_FILE environment variable. If the environment variable is not set,
// it falls back to the provided paths. A non-empty explicit path never falls back. This is useful for containerized environments
// where the config file location may vary.
func WithFileFromEnv(paths ...string) Option {
	if path := os.Getenv(envFilePath); path != "" {
		option := WithFile(path)
		return func(p *Parser) error {
			if err := option(p); err != nil {
				return fmt.Errorf("CONFIG_FILE must identify a readable supported configuration file: %w", err)
			}
			return nil
		}
	}
	return WithFile(paths...)
}
