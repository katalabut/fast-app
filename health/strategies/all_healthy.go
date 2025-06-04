package strategies

import "github.com/katalabut/fast-app/health"

// AllHealthyStrategy requires all health checks to be healthy
type AllHealthyStrategy struct{}

// Aggregate returns healthy only if all checks are healthy
func (s *AllHealthyStrategy) Aggregate(results map[string]health.HealthResult) health.HealthStatus {
	if len(results) == 0 {
		return health.StatusHealthy
	}

	for _, result := range results {
		if result.IsUnhealthy() {
			return health.StatusUnhealthy
		}
		if result.IsDegraded() {
			return health.StatusDegraded
		}
	}

	return health.StatusHealthy
}
