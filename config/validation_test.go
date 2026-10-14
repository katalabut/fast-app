package config_test

import (
	"testing"
	"time"

	"github.com/katalabut/fast-app/config"
	"github.com/katalabut/fast-app/configloader"
)

func TestDefaultsAndValidation(t *testing.T) {
	cfg, err := configloader.New[config.App](configloader.WithEnv("FA_VALIDATE"))
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if !cfg.Observability.Enabled || !cfg.Observability.Debug.Enabled || cfg.Observability.Port != 9090 || cfg.Observability.Health.Timeout != 30*time.Second {
		t.Fatalf("legacy defaults changed: %+v", cfg.Observability)
	}

	for _, change := range []func(*config.App){
		func(c *config.App) { c.Observability.Port = -1 },
		func(c *config.App) { c.Observability.Port = 65536 },
		func(c *config.App) { c.Observability.Health.Timeout = 0 },
		func(c *config.App) { c.Observability.Health.LivePath = c.Observability.Health.ReadyPath },
		func(c *config.App) { c.Observability.Metrics.Path = "/{bad" },
	} {
		copy := *cfg
		change(&copy)
		if err := copy.Validate(); err == nil {
			t.Fatalf("invalid config accepted: %+v", copy)
		}
	}
}
