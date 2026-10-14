package configloader

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/creasty/defaults"
	"github.com/katalabut/fast-app/config"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type formSource map[string]interface{}

func (formSource) Name() string { return "decoder-forms" }
func (s formSource) Load(v *viper.Viper) error {
	for key, value := range s {
		v.Set(key, value)
	}
	return nil
}

type formsConfig[T any] struct {
	Value   T
	Skipped T `default:"-"`
}
type sparseTypedChild struct{ Name string }
type fullTypedChild struct {
	Name    string
	Port    int
	Enabled bool
	Timeout time.Duration
}
type omitTypedChild struct {
	Name    string
	Port    int           `mapstructure:",omitempty"`
	Enabled bool          `mapstructure:",omitempty"`
	Timeout time.Duration `mapstructure:",omitempty"`
}

func TestDecoderFormsDefaultsAndExplicitValues(t *testing.T) {
	forms := []struct {
		name    string
		input   interface{}
		port    int
		enabled bool
		timeout time.Duration
	}{
		{"map missing", map[string]interface{}{"name": "provided"}, 7, true, 2 * time.Second},
		{"map explicit zero", map[string]interface{}{"name": "provided", "port": 0, "enabled": false}, 0, false, 2 * time.Second},
		{"map explicit null", map[string]interface{}{"name": "provided", "port": nil, "enabled": nil, "timeout": nil}, 7, true, 2 * time.Second},
		{"weak scalar numbers", map[string]interface{}{"name": "provided", "port": float64(0), "enabled": int(0), "timeout": "0s"}, 0, false, 0},
		{"weak scalar bool", map[string]interface{}{"name": "provided", "port": false, "enabled": 0, "timeout": "0s"}, 0, false, 0},
		{"weak scalar strings", map[string]string{"name": "provided", "port": "0", "enabled": "false", "timeout": "0s"}, 0, false, 0},
		{"interface key map", map[interface{}]interface{}{"name": "provided", "port": 0, "enabled": false}, 0, false, 2 * time.Second},
		{"typed exact struct", policyChild{Name: "provided"}, 0, false, 0},
		{"typed exact pointer", &policyChild{Name: "provided"}, 0, false, 0},
		{"typed sparse struct", sparseTypedChild{Name: "provided"}, 7, true, 2 * time.Second},
		{"typed sparse pointer", &sparseTypedChild{Name: "provided"}, 7, true, 2 * time.Second},
		{"typed full struct", fullTypedChild{Name: "provided"}, 0, false, 0},
		{"typed omitempty", omitTypedChild{Name: "provided"}, 7, true, 2 * time.Second},
	}
	shapes := []struct {
		name   string
		target func() interface{}
		wrap   func(interface{}) interface{}
	}{
		{"struct", func() interface{} { return new(formsConfig[policyChild]) }, func(v interface{}) interface{} { return v }},
		{"pointer", func() interface{} { return new(formsConfig[*policyChild]) }, func(v interface{}) interface{} { return v }},
		{"slice", func() interface{} { return new(formsConfig[[]policyChild]) }, func(v interface{}) interface{} { return []interface{}{v} }},
		{"pointer slice", func() interface{} { return new(formsConfig[[]*policyChild]) }, func(v interface{}) interface{} { return []interface{}{v} }},
		{"array", func() interface{} { return new(formsConfig[[1]policyChild]) }, func(v interface{}) interface{} { return [1]interface{}{v} }},
		{"weak singleton slice", func() interface{} { return new(formsConfig[[]policyChild]) }, func(v interface{}) interface{} { return v }},
		{"weak singleton pointer slice", func() interface{} { return new(formsConfig[[]*policyChild]) }, func(v interface{}) interface{} { return v }},
		{"map", func() interface{} { return new(formsConfig[map[string]policyChild]) }, func(v interface{}) interface{} { return map[string]interface{}{"main": v} }},
		{"pointer map", func() interface{} { return new(formsConfig[map[string]*policyChild]) }, func(v interface{}) interface{} { return map[string]interface{}{"main": v} }},
	}
	for _, shape := range shapes {
		for _, form := range forms {
			t.Run(shape.name+"/"+form.name, func(t *testing.T) {
				p, err := NewParser()
				if err != nil {
					t.Fatal(err)
				}
				if err = p.SetSource(formSource{"value": shape.wrap(form.input), "skipped": shape.wrap(form.input)}); err != nil {
					t.Fatal(err)
				}
				cfg := shape.target()
				if err = p.Parse(cfg); err != nil {
					t.Fatal(err)
				}
				// Differential oracle: a suppressed subtree must have precisely the
				// native decoder's result for this same Viper snapshot and target type.
				reference := reflect.New(reflect.TypeOf(cfg).Elem())
				decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
					Result: reference.Interface(), WeaklyTypedInput: true,
					DecodeHook: mapstructure.ComposeDecodeHookFunc(mapstructure.StringToTimeDurationHookFunc(), mapstructure.StringToSliceHookFunc(",")),
				})
				if err != nil {
					t.Fatal(err)
				}
				if err = decoder.Decode(p.viper.AllSettings()); err != nil {
					t.Fatal(err)
				}
				root := reflect.ValueOf(cfg).Elem()
				if !reflect.DeepEqual(root.FieldByName("Skipped").Interface(), reference.Elem().FieldByName("Skipped").Interface()) {
					t.Fatal("suppressed subtree differs from native decoder")
				}

				node := formNode(root.FieldByName("Value"))
				skip := formNode(root.FieldByName("Skipped"))
				if node.Name != "provided" || node.Port != form.port || node.Enabled != form.enabled || node.Timeout != form.timeout {
					t.Errorf("ordinary=%+v; want port%d enabled%v timeout%v", node, form.port, form.enabled, form.timeout)
				}
				if skip.Name != "provided" || skip.Port != 0 || skip.Enabled || skip.Timeout != 0 || skip.Nested.Enabled {
					t.Errorf("skip subtree=%+v", skip)
				}
			})
		}
	}
}
func formNode(value reflect.Value) policyChild {
	switch value.Kind() {
	case reflect.Ptr:
		return formNode(value.Elem())
	case reflect.Map:
		return formNode(value.MapIndex(reflect.ValueOf("main")))
	case reflect.Slice, reflect.Array:
		return formNode(value.Index(0))
	default:
		return value.Interface().(policyChild)
	}
}

func TestDecoderNullAndEmptyContainers(t *testing.T) {
	cases := []struct {
		name  string
		value interface{}
	}{
		{"nil", nil}, {"typed nil pointer", (*policyChild)(nil)},
		{"typed nil map", map[string]interface{}(nil)},
		{"empty map", map[string]interface{}{}}, {"empty slice", []interface{}{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := NewParser()
			if err := p.SetSource(formSource{"value": tc.value, "skipped": tc.value}); err != nil {
				t.Fatal(err)
			}
			var cfg formsConfig[[]policyChild]
			if err := p.Parse(&cfg); err != nil {
				t.Fatal(err)
			}
			if len(cfg.Value) != 0 || len(cfg.Skipped) != 0 {
				t.Fatalf("empty/null must not invent child: %+v", cfg)
			}
		})
	}
	t.Run("null elements", func(t *testing.T) {
		p, _ := NewParser()
		_ = p.SetSource(formSource{"value": []interface{}{nil}, "skipped": []interface{}{nil}})
		var cfg formsConfig[[]policyChild]
		if err := p.Parse(&cfg); err != nil {
			t.Fatal(err)
		}
		if len(cfg.Value) != 1 || cfg.Value[0].Port != 0 || cfg.Value[0].Enabled || cfg.Skipped[0].Port != 0 {
			t.Fatalf("null element replaced: %+v", cfg)
		}
	})
}

func TestDecoderWeakMapMergingAndTypedKeys(t *testing.T) {
	type Cfg struct {
		Values  map[string]policyChild `default:"{\"old\":{\"Name\":\"fallback\"}}"`
		Numbers map[int]policyChild
		Skipped map[string]policyChild `default:"-"`
	}
	input := []interface{}{
		map[string]interface{}{"first": map[string]interface{}{"name": "provided", "port": 0, "enabled": false}},
		map[string]interface{}{"second": map[string]interface{}{"name": "missing"}},
	}
	p, _ := NewParser()
	_ = p.SetSource(formSource{"values": input, "skipped": input, "numbers": map[string]interface{}{"12": map[string]interface{}{"name": "provided", "port": 0, "enabled": false}}})
	var cfg Cfg
	if err := p.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Values) != 2 || len(cfg.Skipped) != 2 {
		t.Fatalf("weak map merge lost entries: %+v", cfg)
	}
	if cfg.Values["first"].Port != 0 || cfg.Values["first"].Enabled || cfg.Values["second"].Port != 7 || !cfg.Values["second"].Enabled {
		t.Fatalf("weak map defaults: %+v", cfg.Values)
	}
	if cfg.Skipped["second"].Port != 0 || cfg.Skipped["second"].Enabled || cfg.Numbers[12].Port != 0 || cfg.Numbers[12].Enabled {
		t.Fatal("skip/key conversion overwrote zeros")
	}
}

func TestDecoderTypedAppAndEnvPriority(t *testing.T) {
	type Cfg struct {
		App     config.App
		Skipped config.App `default:"-"`
	}
	app := config.App{Logger: config.Logger{Level: "info"}, Observability: config.Observability{Port: 9090}}
	p, _ := NewParser()
	_ = p.SetSource(formSource{"app": app, "skipped": &app})
	var cfg Cfg
	if err := p.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	for name, obs := range map[string]config.Observability{"ordinary": cfg.App.Observability, "skipped": cfg.Skipped.Observability} {
		if obs.Enabled || obs.Metrics.Enabled || obs.Health.Enabled || obs.Debug.Enabled || obs.Port != 9090 {
			t.Errorf("%s typed disabled exposure: %+v", name, obs)
		}
	}
	for _, key := range []string{"VALUE", "SKIPPED"} {
		t.Setenv("FA_FORMS_"+key+"_PORT", "0")
		t.Setenv("FA_FORMS_"+key+"_ENABLED", "false")
	}
	p, err := NewParser(WithEnv("FA_FORMS"))
	if err != nil {
		t.Fatal(err)
	}
	// MergeConfigMap has file priority; Viper.Set intentionally has higher priority than env.
	if err := p.SetSource(mergeFormsSource{}); err != nil {
		t.Fatal(err)
	}
	var env formsConfig[*policyChild]
	if err := p.Parse(&env); err != nil {
		t.Fatal(err)
	}
	if env.Value == nil || env.Skipped == nil || env.Value.Port != 0 || env.Value.Enabled || env.Value.Timeout != 2*time.Second || env.Skipped.Port != 0 || env.Skipped.Enabled || env.Skipped.Timeout != 0 {
		t.Fatalf("env/source/default precedence: %+v", env)
	}
}

type mergeFormsSource struct{}

func (mergeFormsSource) Name() string { return "merge-forms" }
func (mergeFormsSource) Load(v *viper.Viper) error {
	return v.MergeConfigMap(map[string]interface{}{"value": map[string]interface{}{"port": 9, "enabled": true}, "skipped": map[string]interface{}{"port": 9, "enabled": true}})
}

func TestDecoderExistingOptOutStateAndSourceOwnership(t *testing.T) {
	type Nested struct {
		Skipped  *policyChild `default:"-"`
		Ordinary *policyChild
	}
	type Cfg struct {
		Nested  *Nested
		Skipped map[string]*policyChild `default:"-"`
	}
	original := &policyChild{Name: "retained", Port: 11, Enabled: true}
	cfg := Cfg{Nested: &Nested{Skipped: original}, Skipped: map[string]*policyChild{"old": original}}
	p, _ := NewParser()
	_ = p.SetSource(formSource{"nested": map[string]interface{}{"ordinary": map[string]interface{}{"name": "provided"}}})
	if err := p.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Nested.Skipped.Port != 11 || !cfg.Nested.Skipped.Enabled || cfg.Skipped["old"].Port != 11 || cfg.Nested.Ordinary.Port != 7 {
		t.Fatalf("existing state lost: %+v", cfg.Nested)
	}
	raw := policyChild{Name: "provided"}
	p, _ = NewParser()
	_ = p.SetSource(formSource{"value": &raw, "skipped": &raw})
	var result formsConfig[*policyChild]
	if err := p.Parse(&result); err != nil {
		t.Fatal(err)
	}
	if raw.Port != 0 || raw.Enabled || result.Value.Port != 0 || result.Skipped.Port != 0 {
		t.Fatalf("typed source mutated: %s", fmt.Sprint(raw))
	}
}

var formSetterCalls atomic.Int32

type formSetterChild struct {
	Name string
	Port int
}

func (c *formSetterChild) SetDefaults() {
	formSetterCalls.Add(1)
	if c.Port == 0 {
		c.Port = 7
	}
}
func TestDecoderDefaultsSetterPolicy(t *testing.T) {
	type Cfg struct {
		Existing formSetterChild
		Skipped  formSetterChild `default:"-"`
		Dynamic  []*formSetterChild
		Denied   []*formSetterChild `default:"-"`
	}
	formSetterCalls.Store(0)
	p, _ := NewParser()
	_ = p.SetSource(formSource{"existing": map[string]interface{}{"name": "provided", "port": 0}, "skipped": map[string]interface{}{"name": "provided"}, "dynamic": []interface{}{map[string]interface{}{"name": "provided", "port": 0}}, "denied": []interface{}{map[string]interface{}{"name": "provided"}}})
	var cfg Cfg
	if err := p.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	if calls := formSetterCalls.Load(); calls != 2 {
		t.Fatalf("SetDefaults calls=%d; want existing+dynamic, no skipped or replay calls", calls)
	}
	if cfg.Existing.Port != 0 || cfg.Dynamic[0].Port != 0 || cfg.Skipped.Port != 0 || cfg.Denied[0].Port != 0 {
		t.Fatalf("setter overwrote source/policy: %+v", cfg)
	}
}
func TestDecoderRemainMixedScopes(t *testing.T) {
	type Outer struct{ Child *policyChild }
	type Cfg struct {
		Existing   *Outer `default:"{}"`
		Optional   *Outer
		Containers []struct {
			Ordinary  *policyChild
			Remainder map[string]*policyChild `mapstructure:",remain" default:"-"`
		}
		Remainder map[string]policyChild `mapstructure:",remain" default:"-"`
	}
	child := map[string]interface{}{"name": "provided"}
	p, _ := NewParser()
	_ = p.SetSource(formSource{"existing": map[string]interface{}{"child": child}, "optional": map[string]interface{}{"child": child}, "containers": []interface{}{map[string]interface{}{"ordinary": child, "extra": child}, map[string]interface{}{"ordinary": child, "extra": child}}, "extra": child})
	var cfg Cfg
	if err := p.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Existing.Child.Port != 7 || cfg.Optional.Child.Port != 7 || cfg.Remainder["extra"].Port != 0 {
		t.Fatalf("remaining map leaked policy into siblings: %+v", cfg)
	}
	for _, container := range cfg.Containers {
		if container.Ordinary.Port != 7 || container.Remainder["extra"].Port != 0 {
			t.Fatalf("remaining map leaked across elements: %+v", container)
		}
	}
}
func TestDecoderCaseTagsAndMapKeys(t *testing.T) {
	type Child struct {
		Port    int  `mapstructure:"Listen_Port" default:"7"`
		Enabled bool `mapstructure:"Switch" default:"true"`
	}
	type Config struct {
		Values  map[string]Child
		Other   *Child           `mapstructure:"other_name"`
		Skipped map[string]Child `default:"-"`
	}
	child := map[interface{}]interface{}{"LISTEN_PORT": 0, "sWITCH": false}
	p, _ := NewParser()
	_ = p.SetSource(formSource{"values": map[string]interface{}{"a.b[0]": child, "case": child}, "other_name": child, "skipped": map[string]interface{}{"a.b[0]": child}})
	var cfg Config
	if err := p.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Other == nil || cfg.Other.Port != 0 || cfg.Other.Enabled {
		t.Fatalf("rename/case source: %+v", cfg.Other)
	}
	for _, values := range []map[string]Child{cfg.Values, cfg.Skipped} {
		for _, node := range values {
			if node.Port != 0 || node.Enabled {
				t.Errorf("map key/case defaults overwrite: %+v", values)
			}
		}
	}
}

func TestDecoderNullPointerMatchesNative(t *testing.T) {
	type Cfg struct {
		Value   *policyChild `default:"{}"`
		Skipped *policyChild `default:"-"`
	}
	for _, input := range []interface{}{nil, (*policyChild)(nil), map[string]interface{}(nil), map[string]interface{}{}} {
		t.Run(fmt.Sprintf("%T/%v", input, input), func(t *testing.T) {
			cfg := Cfg{Skipped: &policyChild{Name: "retained", Port: 11, Enabled: true}}
			p, _ := NewParser()
			_ = p.SetSource(formSource{"value": input, "skipped": input})
			if err := p.Parse(&cfg); err != nil {
				t.Fatal(err)
			}
			reference := Cfg{Skipped: &policyChild{Name: "retained", Port: 11, Enabled: true}}
			if err := defaults.Set(&reference); err != nil {
				t.Fatal(err)
			}
			decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Result: &reference, WeaklyTypedInput: true})
			if err != nil {
				t.Fatal(err)
			}
			if err = decoder.Decode(p.viper.AllSettings()); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(cfg, reference) {
				t.Fatalf("nil conversion differs: actual=%+v reference=%+v", cfg, reference)
			}
		})
	}
	// A collection preserves typed nil maps as elements in its source snapshot.
	p, _ := NewParser()
	_ = p.SetSource(formSource{"value": []interface{}{map[string]interface{}(nil)}, "skipped": []interface{}{map[string]interface{}(nil)}})
	var cfg formsConfig[[]*policyChild]
	if err := p.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Value) != 1 || cfg.Value[0] != nil || len(cfg.Skipped) != 1 || cfg.Skipped[0] != nil {
		t.Fatalf("nil-map pointer element became allocated: %+v", cfg)
	}
}

func TestDecoderInterfacesAndTypedOmitTags(t *testing.T) {
	type Cfg struct {
		Value   interface{}
		Skipped interface{} `default:"-"`
	}
	cfg := Cfg{Value: new(policyChild), Skipped: new(policyChild)}
	p, _ := NewParser()
	_ = p.SetSource(formSource{"value": map[string]interface{}{"name": "provided", "port": 0, "enabled": false}, "skipped": map[string]interface{}{"name": "provided"}})
	if err := p.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	value := cfg.Value.(*policyChild)
	skip := cfg.Skipped.(*policyChild)
	if value.Port != 0 || value.Enabled || value.Timeout != 2*time.Second || skip.Port != 0 || skip.Enabled || skip.Timeout != 0 {
		t.Fatalf("interface destinations lost policy: value=%+v skip=%+v", value, skip)
	}
	type Omit struct {
		Name    string
		Port    int  `mapstructure:",omitempty" default:"7"`
		Enabled bool `mapstructure:",omitempty" default:"true"`
	}
	p, _ = NewParser()
	_ = p.SetSource(formSource{"value": Omit{Name: "provided"}, "skipped": Omit{Name: "provided"}})
	var typed formsConfig[Omit]
	if err := p.Parse(&typed); err != nil {
		t.Fatal(err)
	}
	if typed.Value.Port != 0 || typed.Value.Enabled || typed.Skipped.Port != 0 || typed.Skipped.Enabled {
		t.Fatalf("same-type assignment must bypass omitempty: %+v", typed)
	}
}

func TestDecoderCustomOverrideKeepsViperPriority(t *testing.T) {
	t.Setenv("FA_FORMS_VALUE_PORT", "0")
	t.Setenv("FA_FORMS_VALUE_ENABLED", "false")
	p, err := NewParser(WithEnv("FA_FORMS"))
	if err != nil {
		t.Fatal(err)
	}
	if err = p.SetSource(formSource{"value.port": 9, "value.enabled": true}); err != nil {
		t.Fatal(err)
	}
	var cfg formsConfig[*policyChild]
	if err = p.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Value == nil || cfg.Value.Port != 9 || !cfg.Value.Enabled || cfg.Value.Timeout != 2*time.Second {
		t.Fatalf("Viper.Set priority changed: %+v", cfg.Value)
	}
}

func TestDecoderExistingArrayNestedDefaults(t *testing.T) {
	type Cfg struct {
		Value   [1]policyChild
		Skipped [1]policyChild `default:"-"`
	}
	cfg := Cfg{Value: [1]policyChild{{Name: "existing"}}, Skipped: [1]policyChild{{Name: "existing"}}}
	p, _ := NewParser()
	_ = p.SetSource(formSource{"value": []interface{}{map[string]interface{}{"nested": map[string]interface{}{}}}, "skipped": []interface{}{map[string]interface{}{"nested": map[string]interface{}{}}}})
	if err := p.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	if !cfg.Value[0].Nested.Enabled || cfg.Skipped[0].Nested.Enabled || cfg.Value[0].Name != "existing" {
		t.Fatalf("existing array nested policy: %+v", cfg)
	}
}
