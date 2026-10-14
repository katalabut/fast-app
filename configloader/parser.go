package configloader

import (
	"reflect"
	"strings"

	"github.com/creasty/defaults"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Parser loads sources in registration order. A Parser is not safe for concurrent use.
type Parser struct {
	viper   *viper.Viper
	sources []Source
}

type Source interface {
	Name() string
	Load(*viper.Viper) error
}

func NewParser(opts ...Option) (*Parser, error) {
	p := &Parser{viper: viper.New()}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// Parse applies defaults before source values, preserving explicit false, zero and empty values.
// Environment values take precedence over file values. Custom sources load in registration order.
func (p *Parser) Parse(cfg interface{}) error {
	value := reflect.ValueOf(cfg)
	if value.Kind() != reflect.Ptr || value.IsNil() || value.Elem().Kind() != reflect.Struct {
		return errors.New("configuration must be a non-nil struct pointer")
	}
	if err := defaults.Set(cfg); err != nil {
		return errors.Wrap(err, "failed to set defaults")
	}

	// Start afresh so parsing another target does not retain stale keys/config from a previous parse.
	p.viper = viper.New()
	registerKeys(p.viper, value.Elem().Type(), "", make(map[reflect.Type]bool))
	for _, source := range p.sources {
		if err := source.Load(p.viper); err != nil {
			return err
		}
	}
	// Capture exactly one source snapshot, including environment precedence.
	return decodeConfig(p.viper.AllSettings(), value.Elem(), true)
}

// registerKeys makes struct leaves visible to AutomaticEnv even when absent from a file.
// A nil Viper default registers the key without overwriting the target's defaults or enabling env by itself.
func registerKeys(v *viper.Viper, typ reflect.Type, prefix string, visiting map[reflect.Type]bool) {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct || visiting[typ] {
		return
	}
	visiting[typ] = true
	defer delete(visiting, typ)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		tags := strings.Split(field.Tag.Get("mapstructure"), ",")
		if tags[0] == "-" {
			continue
		}
		name := field.Name
		if tags[0] != "" {
			name = tags[0]
		}
		key := prefix + strings.ToLower(name)
		ft := field.Type
		for ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct {
			nestedPrefix := key + "."
			for _, tag := range tags[1:] {
				if tag == "squash" {
					nestedPrefix = prefix
				}
			}
			registerKeys(v, ft, nestedPrefix, visiting)
		} else {
			v.SetDefault(key, nil)
		}
	}
}

// SetSource replaces a source with the same name and places it last in registration order.
func (p *Parser) SetSource(s Source) error {
	if s == nil || (reflect.ValueOf(s).Kind() == reflect.Ptr && reflect.ValueOf(s).IsNil()) {
		return errors.New("empty source")
	}
	for i, existing := range p.sources {
		if existing.Name() == s.Name() {
			p.sources = append(p.sources[:i], p.sources[i+1:]...)
			break
		}
	}
	p.sources = append(p.sources, s)
	return nil
}
