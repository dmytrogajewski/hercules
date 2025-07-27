package cohesion

import (
	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/analyzers/common"
)

// CohesionAggregator aggregates results from multiple cohesion analyses
type CohesionAggregator struct {
	*common.Aggregator
}

// NewCohesionAggregator creates a new CohesionAggregator
func NewCohesionAggregator() *CohesionAggregator {
	numericKeys := []string{"lcom", "cohesion_score", "function_cohesion"}
	countKeys := []string{"total_functions"}

	messageBuilder := func(score float64) string {
		if score >= 0.8 {
			return "Excellent overall cohesion across all analyzed code"
		}
		if score >= 0.6 {
			return "Good overall cohesion with room for improvement"
		}
		if score >= 0.3 {
			return "Fair overall cohesion - consider refactoring some functions"
		}
		return "Poor overall cohesion - significant refactoring recommended"
	}

	emptyResultBuilder := func() analyze.Report {
		return analyze.Report{
			"total_functions":   0,
			"lcom":              0.0,
			"cohesion_score":    1.0,
			"function_cohesion": 1.0,
			"message":           "No functions found",
		}
	}

	return &CohesionAggregator{
		Aggregator: common.NewAggregatorWithCustomEmptyResult(
			"cohesion",
			numericKeys,
			countKeys,
			"functions",
			"name",
			messageBuilder,
			emptyResultBuilder,
		),
	}
}
