package strategies

import (
	"testing"

	"github.com/katalabut/fast-app/health"
)

func TestAllHealthyStrategy(t *testing.T) {
	strategy := &health.AllHealthyStrategy{}

	t.Run("EmptyResults", func(t *testing.T) {
		results := make(map[string]health.HealthResult)
		status := strategy.Aggregate(results)
		if status != health.StatusHealthy {
			t.Errorf("Expected %s for empty results, got %s", health.StatusHealthy, status)
		}
	})

	t.Run("AllHealthy", func(t *testing.T) {
		results := map[string]health.HealthResult{
			"check1": health.NewHealthyResult("ok"),
			"check2": health.NewHealthyResult("ok"),
			"check3": health.NewHealthyResult("ok"),
		}
		status := strategy.Aggregate(results)
		if status != health.StatusHealthy {
			t.Errorf("Expected %s for all healthy, got %s", health.StatusHealthy, status)
		}
	})

	t.Run("OneDegraded", func(t *testing.T) {
		results := map[string]health.HealthResult{
			"check1": health.NewHealthyResult("ok"),
			"check2": health.NewDegradedResult("slow"),
			"check3": health.NewHealthyResult("ok"),
		}
		status := strategy.Aggregate(results)
		if status != health.StatusDegraded {
			t.Errorf("Expected %s for one degraded, got %s", health.StatusDegraded, status)
		}
	})

	t.Run("OneUnhealthy", func(t *testing.T) {
		results := map[string]health.HealthResult{
			"check1": health.NewHealthyResult("ok"),
			"check2": health.NewUnhealthyResult("failed"),
			"check3": health.NewHealthyResult("ok"),
		}
		status := strategy.Aggregate(results)
		if status != health.StatusUnhealthy {
			t.Errorf("Expected %s for one unhealthy, got %s", health.StatusUnhealthy, status)
		}
	})

	t.Run("MixedStatusesWithUnhealthy", func(t *testing.T) {
		results := map[string]health.HealthResult{
			"check1": health.NewHealthyResult("ok"),
			"check2": health.NewDegradedResult("slow"),
			"check3": health.NewUnhealthyResult("failed"),
		}
		status := strategy.Aggregate(results)
		// Unhealthy takes precedence over degraded
		if status != health.StatusUnhealthy {
			t.Errorf("Expected %s for mixed statuses with unhealthy, got %s", health.StatusUnhealthy, status)
		}
	})

	t.Run("MixedStatusesWithoutUnhealthy", func(t *testing.T) {
		results := map[string]health.HealthResult{
			"check1": health.NewHealthyResult("ok"),
			"check2": health.NewDegradedResult("slow"),
			"check3": health.NewHealthyResult("ok"),
		}
		status := strategy.Aggregate(results)
		// Should be degraded when no unhealthy but has degraded
		if status != health.StatusDegraded {
			t.Errorf("Expected %s for mixed statuses without unhealthy, got %s", health.StatusDegraded, status)
		}
	})
}

func TestMajorityHealthyStrategy(t *testing.T) {
	strategy := &MajorityHealthyStrategy{}

	t.Run("EmptyResults", func(t *testing.T) {
		results := make(map[string]health.HealthResult)
		status := strategy.Aggregate(results)
		if status != health.StatusHealthy {
			t.Errorf("Expected %s for empty results, got %s", health.StatusHealthy, status)
		}
	})

	t.Run("MajorityHealthy", func(t *testing.T) {
		results := map[string]health.HealthResult{
			"check1": health.NewHealthyResult("ok"),
			"check2": health.NewHealthyResult("ok"),
			"check3": health.NewUnhealthyResult("failed"),
		}
		status := strategy.Aggregate(results)
		if status != health.StatusHealthy {
			t.Errorf("Expected %s for majority healthy, got %s", health.StatusHealthy, status)
		}
	})

	t.Run("MajorityUnhealthy", func(t *testing.T) {
		results := map[string]health.HealthResult{
			"check1": health.NewUnhealthyResult("failed"),
			"check2": health.NewUnhealthyResult("failed"),
			"check3": health.NewHealthyResult("ok"),
		}
		status := strategy.Aggregate(results)
		if status != health.StatusUnhealthy {
			t.Errorf("Expected %s for majority unhealthy, got %s", health.StatusUnhealthy, status)
		}
	})

	t.Run("EqualSplit", func(t *testing.T) {
		results := map[string]health.HealthResult{
			"check1": health.NewHealthyResult("ok"),
			"check2": health.NewUnhealthyResult("failed"),
		}
		status := strategy.Aggregate(results)
		// With equal split, should return degraded
		if status != health.StatusDegraded {
			t.Errorf("Expected %s for equal split, got %s", health.StatusDegraded, status)
		}
	})

	t.Run("MajorityDegraded", func(t *testing.T) {
		results := map[string]health.HealthResult{
			"check1": health.NewDegradedResult("slow"),
			"check2": health.NewDegradedResult("slow"),
			"check3": health.NewHealthyResult("ok"),
		}
		status := strategy.Aggregate(results)
		// No majority healthy or unhealthy, should be degraded
		if status != health.StatusDegraded {
			t.Errorf("Expected %s for majority degraded, got %s", health.StatusDegraded, status)
		}
	})

	t.Run("SingleCheck", func(t *testing.T) {
		results := map[string]health.HealthResult{
			"check1": health.NewHealthyResult("ok"),
		}
		status := strategy.Aggregate(results)
		if status != health.StatusHealthy {
			t.Errorf("Expected %s for single healthy check, got %s", health.StatusHealthy, status)
		}
	})
}

func TestWeightedStrategy(t *testing.T) {
	t.Run("NewWeightedStrategy", func(t *testing.T) {
		weights := map[string]health.ComponentImportance{
			"database": health.Critical,
			"cache":    health.Important,
			"metrics":  health.Optional,
		}
		strategy := NewWeightedStrategy(weights)

		if strategy.Weights["database"] != health.Critical {
			t.Errorf("Expected database to be Critical, got %v", strategy.Weights["database"])
		}
	})

	t.Run("CriticalUnhealthy", func(t *testing.T) {
		weights := map[string]health.ComponentImportance{
			"database": health.Critical,
			"cache":    health.Important,
		}
		strategy := NewWeightedStrategy(weights)

		results := map[string]health.HealthResult{
			"database": health.NewUnhealthyResult("db down"),
			"cache":    health.NewHealthyResult("ok"),
		}

		status := strategy.Aggregate(results)
		if status != health.StatusUnhealthy {
			t.Errorf("Expected %s when critical component unhealthy, got %s", health.StatusUnhealthy, status)
		}
	})

	t.Run("ImportantUnhealthy", func(t *testing.T) {
		weights := map[string]health.ComponentImportance{
			"database": health.Critical,
			"cache":    health.Important,
		}
		strategy := NewWeightedStrategy(weights)

		results := map[string]health.HealthResult{
			"database": health.NewHealthyResult("ok"),
			"cache":    health.NewUnhealthyResult("cache down"),
		}

		status := strategy.Aggregate(results)
		if status != health.StatusDegraded {
			t.Errorf("Expected %s when important component unhealthy, got %s", health.StatusDegraded, status)
		}
	})

	t.Run("OptionalUnhealthy", func(t *testing.T) {
		weights := map[string]health.ComponentImportance{
			"database": health.Critical,
			"metrics":  health.Optional,
		}
		strategy := NewWeightedStrategy(weights)

		results := map[string]health.HealthResult{
			"database": health.NewHealthyResult("ok"),
			"metrics":  health.NewUnhealthyResult("metrics down"),
		}

		status := strategy.Aggregate(results)
		if status != health.StatusHealthy {
			t.Errorf("Expected %s when only optional component unhealthy, got %s", health.StatusHealthy, status)
		}
	})

	t.Run("DegradedComponent", func(t *testing.T) {
		weights := map[string]health.ComponentImportance{
			"database": health.Critical,
			"cache":    health.Important,
		}
		strategy := NewWeightedStrategy(weights)

		results := map[string]health.HealthResult{
			"database": health.NewHealthyResult("ok"),
			"cache":    health.NewDegradedResult("slow"),
		}

		status := strategy.Aggregate(results)
		if status != health.StatusDegraded {
			t.Errorf("Expected %s when component degraded, got %s", health.StatusDegraded, status)
		}
	})

	t.Run("UnknownComponent", func(t *testing.T) {
		weights := map[string]health.ComponentImportance{
			"database": health.Critical,
		}
		strategy := NewWeightedStrategy(weights)

		results := map[string]health.HealthResult{
			"database":        health.NewHealthyResult("ok"),
			"unknown-service": health.NewUnhealthyResult("failed"),
		}

		status := strategy.Aggregate(results)
		// Unknown component defaults to Important, so should be degraded
		if status != health.StatusDegraded {
			t.Errorf("Expected %s when unknown component unhealthy, got %s", health.StatusDegraded, status)
		}
	})

	t.Run("AllHealthy", func(t *testing.T) {
		weights := map[string]health.ComponentImportance{
			"database": health.Critical,
			"cache":    health.Important,
			"metrics":  health.Optional,
		}
		strategy := NewWeightedStrategy(weights)

		results := map[string]health.HealthResult{
			"database": health.NewHealthyResult("ok"),
			"cache":    health.NewHealthyResult("ok"),
			"metrics":  health.NewHealthyResult("ok"),
		}

		status := strategy.Aggregate(results)
		if status != health.StatusHealthy {
			t.Errorf("Expected %s when all components healthy, got %s", health.StatusHealthy, status)
		}
	})
}
