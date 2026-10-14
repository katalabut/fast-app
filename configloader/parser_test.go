package configloader

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/spf13/viper"
)

type testConfig struct {
	Enabled bool          `default:"true"`
	Port    int           `default:"1234"`
	Name    string        `default:"fallback"`
	Timeout time.Duration `default:"2s"`
	Tags    []string      `default:"[\"fallback\"]"`
	Flag    *bool         `default:"true"`
	Nested  struct {
		Port int `default:"2345"`
	}
	Optional *struct{ Value string }
}

func configFile(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestConfigDefaultsAndOverrides(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		cfg, err := New[testConfig](WithEnv("FA_PROD"))
		if err != nil {
			t.Fatal(err)
		}
		if !cfg.Enabled || cfg.Port != 1234 || cfg.Timeout != 2*time.Second || cfg.Nested.Port != 2345 || cfg.Flag == nil || !*cfg.Flag || cfg.Optional != nil {
			t.Fatalf("defaults not applied: %+v", cfg)
		}
	})
	t.Run("file explicit zero", func(t *testing.T) {
		cfg, err := New[testConfig](WithEnv("FA_PROD"), WithFile(configFile(t, "enabled: false\nport: 0\nname: ''\nflag: false\ntags: []\n")))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Enabled || cfg.Port != 0 || cfg.Name != "" || cfg.Flag == nil || *cfg.Flag || len(cfg.Tags) != 0 {
			t.Fatalf("explicit zero overridden: %+v", cfg)
		}
		if cfg.Nested.Port != 2345 {
			t.Fatalf("missing field lost default: %+v", cfg)
		}
	})
	for _, prefix := range []string{"FA_PROD", "FA_PROD_"} {
		t.Run("env only "+prefix, func(t *testing.T) {
			t.Setenv("FA_PROD_PORT", "4321")
			t.Setenv("FA_PROD_NESTED_PORT", "5432")
			t.Setenv("FA_PROD_ENABLED", "false")
			t.Setenv("FA_PROD_NAME", "")
			t.Setenv("FA_PROD_TIMEOUT", "7s")
			t.Setenv("FA_PROD_TAGS", "a,b")
			t.Setenv("FA_PROD_FLAG", "false")
			t.Setenv("FA_PROD_OPTIONAL_VALUE", "present")
			cfg, err := New[testConfig](WithEnv(prefix))
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Port != 4321 || cfg.Nested.Port != 5432 || cfg.Enabled || cfg.Name != "" || cfg.Timeout != 7*time.Second || !reflect.DeepEqual(cfg.Tags, []string{"a", "b"}) || cfg.Flag == nil || *cfg.Flag || cfg.Optional == nil || cfg.Optional.Value != "present" {
				t.Fatalf("env missing/overridden: %+v", cfg)
			}
		})
	}
	t.Run("env overrides known and absent file keys", func(t *testing.T) {
		t.Setenv("FA_PROD_PORT", "4321")
		t.Setenv("FA_PROD_NESTED_PORT", "5432")
		cfg, err := New[testConfig](WithFile(configFile(t, "port: 9999\nname: file\n")), WithEnv("FA_PROD"))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Port != 4321 || cfg.Nested.Port != 5432 || cfg.Name != "file" {
			t.Fatalf("wrong precedence: %+v", cfg)
		}
	})
}

type namedSource struct {
	name, value string
	calls       *[]string
}

func (s namedSource) Name() string { return s.name }
func (s namedSource) Load(v *viper.Viper) error {
	*s.calls = append(*s.calls, s.name)
	return v.MergeConfigMap(map[string]interface{}{"value": s.value})
}

func TestSourceRegistrationOrder(t *testing.T) {
	for i := 0; i < 50; i++ {
		var calls []string
		p, err := NewParser()
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range []namedSource{{"first", "old", &calls}, {"second", "second", &calls}, {"first", "replacement", &calls}} {
			if err := p.SetSource(s); err != nil {
				t.Fatal(err)
			}
		}
		var cfg struct{ Value string }
		if err := p.Parse(&cfg); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(calls, []string{"second", "first"}) || cfg.Value != "replacement" {
			t.Fatalf("nondeterministic/replaced source: calls=%v value=%q", calls, cfg.Value)
		}
	}
}

func TestConfigTagsAndInvalidTargets(t *testing.T) {
	t.Run("renamed and squashed fields", func(t *testing.T) {
		type Embedded struct{ Name string }
		type Config struct {
			Embedded `mapstructure:",squash"`
			Port     int    `mapstructure:"listen_port"`
			Ignored  string `mapstructure:"-"`
		}
		t.Setenv("FA_PROD_LISTEN_PORT", "8081")
		t.Setenv("FA_PROD_NAME", "tagged")
		t.Setenv("FA_PROD_IGNORED", "should not load")
		cfg, err := New[Config](WithEnv("FA_PROD"))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Port != 8081 || cfg.Name != "tagged" || cfg.Ignored != "" {
			t.Fatalf("incorrect tag handling: %+v", cfg)
		}
	})
	p, err := NewParser()
	if err != nil {
		t.Fatal(err)
	}
	var nilCfg *testConfig
	for _, cfg := range []interface{}{nil, nilCfg, 1, new(int)} {
		func() {
			defer func() {
				if x := recover(); x != nil {
					t.Errorf("Parse panicked for %T: %v", cfg, x)
				}
			}()
			if err := p.Parse(cfg); err == nil {
				t.Errorf("Parse accepted %T", cfg)
			}
		}()
	}
}

func TestProvidedContainersReplaceDefaults(t *testing.T) {
	type Config struct {
		Values map[string]string `default:"{\"baseline\":\"yes\"}"`
		Tags   []string          `default:"[\"one\",\"two\"]"`
		Nested *struct {
			Port    int  `default:"7"`
			Enabled bool `default:"true"`
		} `default:"{}"`
	}
	cfg, err := New[Config](WithEnv("FA_PROD"), WithFile(configFile(t, "values: {user: yes}\ntags: [new]\nnested: {enabled: false}\n")))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.Values, map[string]string{"user": "yes"}) || !reflect.DeepEqual(cfg.Tags, []string{"new"}) || cfg.Nested == nil || cfg.Nested.Port != 7 || cfg.Nested.Enabled {
		t.Fatalf("defaults leaked or were lost: %+v", cfg)
	}
}

func TestParserDoesNotRetainPreviousFile(t *testing.T) {
	p, err := NewParser(WithFile(configFile(t, "port: 4321\nname: first\n")))
	if err != nil {
		t.Fatal(err)
	}
	var first testConfig
	if err := p.Parse(&first); err != nil {
		t.Fatal(err)
	}
	if err := WithFile(configFile(t, "name: second\n"))(p); err != nil {
		t.Fatal(err)
	}
	var second testConfig
	if err := p.Parse(&second); err != nil {
		t.Fatal(err)
	}
	if first.Port != 4321 || second.Port != 1234 || second.Name != "second" {
		t.Fatalf("stale file values: first=%+v second=%+v", first, second)
	}
}

func TestInvalidEnvDuration(t *testing.T) {
	t.Setenv("FA_PROD_TIMEOUT", "invalid")
	if _, err := New[testConfig](WithEnv("FA_PROD")); err == nil {
		t.Fatal("invalid duration accepted")
	}
}

func TestExplicitConfigFileMustNotSilentlyFallback(t *testing.T) {
	t.Setenv("CONFIG_FILE", filepath.Join(t.TempDir(), "missing.yaml"))
	if _, err := New[testConfig](WithFileFromEnv(configFile(t, "name: fallback\n"))); err == nil {
		t.Fatal("explicit missing CONFIG_FILE silently ignored")
	}
}
