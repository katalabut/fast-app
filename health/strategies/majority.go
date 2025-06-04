package strategies

import "github.com/katalabut/fast-app/health"

// MajorityHealthyStrategy requires majority of health checks to be healthy
type MajorityHealthyStrategy struct{}

// Aggregate returns healthy if majority of checks are healthy
func (s *MajorityHealthyStrategy) Aggregate(results map[string]health.HealthResult) health.HealthStatus {
	if len(results) == 0 {
		return health.StatusHealthy
	}

	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0

	for _, result := range results {
		switch result.Status {
		case health.StatusHealthy:
			healthyCount++
		case health.StatusDegraded:
			degradedCount++
		case health.StatusUnhealthy:
			unhealthyCount++
		}
	}

	total := len(results)
	majority := total/2 + 1

	// If majority is unhealthy, return unhealthy
	if unhealthyCount >= majority {
		return health.StatusUnhealthy
	}

	// If majority is healthy, return healthy
	if healthyCount >= majority {
		return health.StatusHealthy
	}

	// Otherwise, return degraded
	return health.StatusDegraded
}
