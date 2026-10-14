package config

import (
	"fmt"
	"net/http"
	"strings"
)

// Validate checks the library-owned configuration without reporting config values.
// Applications should validate their business configuration in their composition root.
func (c App) Validate() error {
	if c.AutoMaxProcs.Enabled && c.AutoMaxProcs.Min < 0 {
		return fmt.Errorf("automaxprocs.min must be non-negative")
	}
	return c.Observability.Validate()
}

// Validate rejects invalid listeners and invalid/conflicting endpoint patterns before startup.
func (c Observability) Validate() error {
	return c.ValidateWithPprof(false)
}

// ValidateWithPprof validates the actual route set for an explicit pprof opt-in.
// Custom prefixes retain their historical registration even without opt-in.
func (c Observability) ValidateWithPprof(pprofEnabled bool) error {
	if !c.Enabled {
		return nil
	}
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("observability.port must be between 0 and 65535")
	}

	mux := http.NewServeMux()
	add := func(field, pattern string) (err error) {
		if pattern == "" {
			return fmt.Errorf("%s must be a non-empty HTTP pattern", field)
		}
		defer func() {
			if recover() != nil {
				err = fmt.Errorf("%s is an invalid or conflicting HTTP path", field)
			}
		}()
		mux.HandleFunc(pattern, func(http.ResponseWriter, *http.Request) {})
		return nil
	}
	if c.Metrics.Enabled {
		if err := add("observability.metrics.path", c.Metrics.Path); err != nil {
			return err
		}
	}
	if c.Health.Enabled {
		if c.Health.Timeout <= 0 {
			return fmt.Errorf("observability.health.timeout must be positive")
		}
		for _, route := range []struct{ field, path string }{{"livepath", c.Health.LivePath}, {"readypath", c.Health.ReadyPath}, {"checkpath", c.Health.CheckPath}} {
			if err := add("observability.health."+route.field, route.path); err != nil {
				return err
			}
		}
	}
	if c.Debug.Enabled {
		if !strings.HasPrefix(c.Debug.PathPrefix, "/") || strings.ContainsAny(c.Debug.PathPrefix, "?#{} \t\r\n") {
			return fmt.Errorf("observability.debug.pathprefix must be a non-empty HTTP pattern")
		}
		if !pprofEnabled && c.Debug.PathPrefix == "/debug" {
			return nil // Dormant routes do not reserve patterns belonging to other endpoints.
		}
		prefix := strings.TrimSuffix(c.Debug.PathPrefix, "/") + "/pprof/"
		for _, suffix := range []string{"", "cmdline", "profile", "symbol", "trace"} {
			if err := add("observability.debug.pathprefix", prefix+suffix); err != nil {
				return err
			}
		}
	}
	return nil
}
