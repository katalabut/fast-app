package fastapp

import (
	"time"

	"github.com/katalabut/fast-app/logger"
)

type (
	Config struct {
		Logger logger.Config

		AutoMaxProcs struct {
			Enabled bool
			Min     int
		}

		Health HealthConfig
	}

	HealthConfig struct {
		Enabled   bool          `default:"true"`
		Port      int           `default:"8080"`
		LivePath  string        `default:"/health/live"`
		ReadyPath string        `default:"/health/ready"`
		CheckPath string        `default:"/health/checks"`
		Timeout   time.Duration `default:"30s"`
		CacheTTL  time.Duration `default:"5s"`
	}
)
