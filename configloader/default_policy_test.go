package configloader

import (
	"testing"
	"time"

	"github.com/katalabut/fast-app/config"
)

type policyChild struct {
	Name    string
	Port    int           `default:"7"`
	Enabled bool          `default:"true"`
	Timeout time.Duration `default:"2s"`
	Nested  struct {
		Enabled bool `default:"true"`
	}
}

type policyConfig struct {
	Ordinary     *policyChild
	Plain        policyChild                  `default:"-"`
	Pointer      *policyChild                 `default:"-"`
	Slice        []policyChild                `default:"-"`
	PointerSlice []*policyChild               `default:"-"`
	Map          map[string]policyChild       `default:"-"`
	PointerMap   map[string]*policyChild      `default:"-"`
	Ancestor     struct{ Child *policyChild } `default:"-"`
	Renamed      *policyChild                 `mapstructure:"custom_name" default:"-"`
	Remainder    map[string]policyChild       `mapstructure:",remain" default:"-"`
	App          config.App                   `default:"-"`
}

func TestDefaultPolicyByFieldAndAncestor(t *testing.T) {
	body := "ordinary: {name: allowed, port: 0, enabled: false}\nplain: {name: skip}\npointer: {name: skip}\nslice: [{name: skip}]\npointerslice: [{name: skip}]\nmap: {main: {name: skip}}\npointermap: {main: {name: skip}}\nancestor: {child: {name: skip}}\ncustom_name: {name: skip}\nextra: {name: skip}\napp: {logger: {level: info}}\n"
	cfg, err := New[policyConfig](WithEnv("FA_POLICY"), WithFile(configFile(t, body)))
	if err != nil {
		t.Fatal(err)
	}
	for name, node := range map[string]policyChild{"plain": cfg.Plain, "pointer": *cfg.Pointer, "slice": cfg.Slice[0], "pointer slice": *cfg.PointerSlice[0], "map": cfg.Map["main"], "pointer map": *cfg.PointerMap["main"], "ancestor": *cfg.Ancestor.Child, "renamed": *cfg.Renamed, "remain": cfg.Remainder["extra"]} {
		if node.Port != 0 || node.Enabled || node.Timeout != 0 || node.Nested.Enabled || node.Name != "skip" {
			t.Errorf("%s defaults policy bypassed: %+v", name, node)
		}
	}
	if cfg.Ordinary.Port != 0 || cfg.Ordinary.Enabled || cfg.Ordinary.Timeout != 2*time.Second || !cfg.Ordinary.Nested.Enabled {
		t.Fatalf("ordinary defaults/explicit values lost: %+v", cfg.Ordinary)
	}
	if cfg.App.Observability.Enabled || cfg.App.Observability.Port != 0 || cfg.App.Observability.Metrics.Enabled || cfg.App.Observability.Health.Enabled || cfg.App.Observability.Debug.Enabled {
		t.Fatal("App opt-out activated exposure")
	}
}

func TestOptOutDoesNotBlockExplicitEnvironment(t *testing.T) {
	t.Setenv("FA_POLICY_CUSTOM_NAME_PORT", "9")
	t.Setenv("FA_POLICY_CUSTOM_NAME_ENABLED", "true")
	t.Setenv("FA_POLICY_APP_OBSERVABILITY_ENABLED", "false")
	cfg, err := New[policyConfig](WithEnv("FA_POLICY"), WithFile(configFile(t, "custom_name: {name: skip, port: 0, enabled: false}\napp: {logger: {level: info}}\n")))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Renamed.Port != 9 || !cfg.Renamed.Enabled || cfg.Renamed.Timeout != 0 || cfg.App.Observability.Enabled || cfg.App.Observability.Port != 0 {
		t.Fatalf("opt-out env priority: %+v", cfg.Renamed)
	}
}

func TestSquashOptOutAndEmptyContainers(t *testing.T) {
	type Flattened struct {
		Port    int  `default:"7"`
		Enabled bool `default:"true"`
	}
	type Cfg struct {
		Flattened `mapstructure:",squash" default:"-"`
		Other     *policyChild
		Slice     []policyChild          `default:"[{\"Name\":\"fallback\"}]"`
		Map       map[string]policyChild `default:"{\"main\":{\"Name\":\"fallback\"}}"`
	}
	cfg, err := New[Cfg](WithEnv("FA_POLICY"), WithFile(configFile(t, "enabled: false\nother: {name: allowed}\nslice: []\nmap: {}\n")))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 0 || cfg.Enabled || cfg.Other == nil || cfg.Other.Port != 7 || len(cfg.Slice) != 0 || len(cfg.Map) != 0 {
		t.Fatalf("squash/container metadata: %+v", cfg)
	}
	if cfg.Other == nil || cfg.Other.Timeout != 2*time.Second {
		t.Fatal("sibling ordinary defaults missing")
	}
}

// The policy belongs to each field, even when decoder metadata maps the same
// child type to a renamed key, a squash, a remainder or an array element.
func TestDefaultPolicyAdjacentMetadata(t *testing.T) {
	type Flattened struct {
		Port    int  `default:"7"`
		Enabled bool `default:"true"`
	}
	type Cfg struct {
		Flattened `mapstructure:",squash"`
		Renamed   *policyChild           `mapstructure:"chosen"`
		Ignored   policyChild            `mapstructure:"-"`
		Skipped   policyChild            `mapstructure:"-" default:"-"`
		Array     [1]policyChild         `default:"-"`
		Remainder map[string]policyChild `mapstructure:",remain"`
	}
	cfg, err := New[Cfg](WithFile(configFile(t, "port: 0\nenabled: false\nchosen: {name: provided, port: 0, enabled: false}\narray: [{name: skip}]\nextra: {name: provided, port: 0, enabled: false}\n")))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 0 || cfg.Enabled {
		t.Fatalf("squash explicit values overwritten: %+v", cfg.Flattened)
	}
	for name, child := range map[string]policyChild{"renamed": *cfg.Renamed, "remain": cfg.Remainder["extra"]} {
		if child.Port != 0 || child.Enabled || child.Timeout != 2*time.Second || !child.Nested.Enabled {
			t.Errorf("%s source metadata: %+v", name, child)
		}
	}
	if cfg.Ignored.Port != 7 || !cfg.Ignored.Enabled {
		t.Errorf("ignored field defaults: %+v", cfg.Ignored)
	}
	if cfg.Skipped.Port != 0 || cfg.Skipped.Enabled || cfg.Array[0].Port != 0 || cfg.Array[0].Enabled || cfg.Array[0].Nested.Enabled {
		t.Fatal("ignored/array opt-out bypassed")
	}
}
