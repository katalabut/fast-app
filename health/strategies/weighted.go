package strategies

import "github.com/katalabut/fast-app/health"

// WeightedStrategy uses component importance to determine overall health
type WeightedStrategy struct {
	Weights map[string]health.ComponentImportance
}

// NewWeightedStrategy creates a new weighted strategy
func NewWeightedStrategy(weights map[string]health.ComponentImportance) *WeightedStrategy {
	return &WeightedStrategy{
		Weights: weights,
	}
}

// Aggregate returns health status based on component importance
func (s *WeightedStrategy) Aggregate(results map[string]health.HealthResult) health.HealthStatus {
	if len(results) == 0 {
		return health.StatusHealthy
	}

	hasCriticalUnhealthy := false
	hasImportantUnhealthy := false
	hasDegraded := false

	for name, result := range results {
		importance, exists := s.Weights[name]
		if !exists {
			importance = health.Important // default importance
		}

		switch result.Status {
		case health.StatusUnhealthy:
			switch importance {
			case health.Critical:
				hasCriticalUnhealthy = true
			case health.Important:
				hasImportantUnhealthy = true
			}
		case health.StatusDegraded:
			hasDegraded = true
		}
	}

	// If any critical component is unhealthy, overall is unhealthy
	if hasCriticalUnhealthy {
		return health.StatusUnhealthy
	}

	// If any important component is unhealthy, overall is degraded
	if hasImportantUnhealthy {
		return health.StatusDegraded
	}

	// If any component is degraded, overall is degraded
	if hasDegraded {
		return health.StatusDegraded
	}

	return health.StatusHealthy
}
