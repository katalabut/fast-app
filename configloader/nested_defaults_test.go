package configloader

import (
	"fmt"
	"testing"
	"time"
)

type nestedDefaults struct {
	Name    string
	Port    int           `default:"7"`
	Timeout time.Duration `default:"2s"`
	Enabled bool          `default:"true"`
	Note    string        `default:"fallback"`
}
type dynamicDefaultsConfig struct {
	Optional     *nestedDefaults
	Backends     []nestedDefaults
	Pointers     []*nestedDefaults
	Pools        map[string]nestedDefaults
	PointerPools map[string]*nestedDefaults
}

func TestDynamicNestedDefaults(t *testing.T) {
	for _, tc := range []struct {
		name, fields string
		port         int
		timeout      time.Duration
		enabled      bool
		note         string
	}{
		{"omitted", "", 7, 2 * time.Second, true, "fallback"},
		{"explicit zero beside missing duration", ", port: 0, enabled: false, note: ''", 0, 2 * time.Second, false, ""},
		{"explicit duration beside missing port", ", timeout: 0s", 7, 0, true, "fallback"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := fmt.Sprintf("optional: {name: optional%s}\nbackends: [{name: slice%s}]\npointers: [{name: pointer-slice%s}]\npools: {main: {name: map%s}}\npointerpools: {main: {name: pointer-map%s}}\n", tc.fields, tc.fields, tc.fields, tc.fields, tc.fields)
			cfg, err := New[dynamicDefaultsConfig](WithEnv("FA_DYNAMIC"), WithFile(configFile(t, body)))
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Optional == nil || len(cfg.Backends) != 1 || len(cfg.Pointers) != 1 || cfg.Pointers[0] == nil || cfg.PointerPools["main"] == nil {
				t.Fatalf("nodes missing: %+v", cfg)
			}
			for name, node := range map[string]nestedDefaults{"optional": *cfg.Optional, "slice": cfg.Backends[0], "pointer slice": *cfg.Pointers[0], "map": cfg.Pools["main"], "pointer map": *cfg.PointerPools["main"]} {
				if node.Port != tc.port || node.Timeout != tc.timeout || node.Enabled != tc.enabled || node.Note != tc.note {
					t.Errorf("%s defaults/explicit values: %+v", name, node)
				}
			}
		})
	}
}

func TestDynamicPointerEnvPriority(t *testing.T) {
	t.Setenv("FA_DYNAMIC_OPTIONAL_PORT", "0")
	t.Setenv("FA_DYNAMIC_OPTIONAL_ENABLED", "false")
	cfg, err := New[dynamicDefaultsConfig](WithEnv("FA_DYNAMIC"), WithFile(configFile(t, "optional: {name: provided, port: 9, enabled: true}\n")))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Optional == nil || cfg.Optional.Port != 0 || cfg.Optional.Enabled || cfg.Optional.Timeout != 2*time.Second || cfg.Optional.Note != "fallback" {
		t.Fatalf("env/default precedence: %+v", cfg.Optional)
	}
}
