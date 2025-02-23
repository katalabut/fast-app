package config

import (
	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Parser struct {
	viper   *viper.Viper
	sources map[string]Source
}

type Source interface {
	Name() string
	Load(*viper.Viper) error
}

func NewParser(opts ...Option) (*Parser, error) {
	p := &Parser{
		viper:   viper.New(),
		sources: make(map[string]Source),
	}

	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (p *Parser) Parse(cfg interface{}) error {
	for _, source := range p.sources {
		if err := source.Load(p.viper); err != nil {
			return err
		}
	}

	hooks := []mapstructure.DecodeHookFunc{
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	}

	viperOpts := viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(hooks...))

	if err := p.viper.Unmarshal(cfg, viperOpts); err != nil {
		return err
	}

	if err := defaults.Set(cfg); err != nil {
		return errors.Wrap(err, "failed to set defaults")
	}

	// todo: validate cfg

	return nil
}

func (p *Parser) SetSource(s Source) error {
	if s == nil {
		return errors.New("empty source")
	}

	p.sources[s.Name()] = s

	return nil
}
