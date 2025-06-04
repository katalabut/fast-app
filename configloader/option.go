package configloader

import (
	"os"

	"github.com/katalabut/fast-app/configloader/source"
)

const (
	envFilePath = "CONFIG_FILE"
)

type Option func(*Parser) error

func WithEnv(prefix string) Option {
	return func(p *Parser) error {
		src := source.NewEnv(prefix)

		return p.SetSource(src)
	}
}

func WithFile(paths ...string) Option {
	return func(p *Parser) error {
		src, err := source.NewFile(paths...)
		if err != nil {
			return err
		}
		return p.SetSource(src)
	}
}

func WithFileFromEnv(paths ...string) Option {
	var pathsState []string
	if path := os.Getenv(envFilePath); path != "" {
		pathsState = append(pathsState, path)
	}
	pathsState = append(pathsState, paths...)

	return WithFile(pathsState...)
}
